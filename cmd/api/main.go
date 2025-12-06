package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mizbain-Fathima/remote-care-backend/internal/cache"
	"github.com/Mizbain-Fathima/remote-care-backend/internal/config"
	"github.com/Mizbain-Fathima/remote-care-backend/internal/db"
	"github.com/Mizbain-Fathima/remote-care-backend/internal/mq"
	"github.com/Mizbain-Fathima/remote-care-backend/internal/observability"
	"github.com/Mizbain-Fathima/remote-care-backend/internal/service"
	"github.com/Mizbain-Fathima/remote-care-backend/internal/transport"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	pp "net/http/pprof"

	"go.uber.org/zap"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	} else {
		log.Println("Loaded .env successfully")
	}

	cfg := config.Load()

	logger, err := observability.NewLogger()
	if err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer logger.Sync()

	ctx := context.Background()

	// OpenTelemetry (optional)
	if os.Getenv("OTEL_ENABLED") == "false" {
		logger.Info("OTEL disabled — skipping tracer initialization")
	} else {
		tp, err := observability.InitOTEL(cfg.ServiceName)
		if err != nil {
			logger.Fatal("otel init failed", zap.Error(err))
		}
		defer tp.Shutdown(ctx)
	}

	// PostgreSQL connection
	dbConn, err := db.New(ctx, cfg.PostgresURL)
	if err != nil {
		logger.Fatal("db connection failed", zap.Error(err))
	}
	logger.Info("Connected to PostgreSQL", zap.String("url", cfg.PostgresURL))
	defer dbConn.Close()

	// Redis client
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logger.Fatal("Failed to parse Redis URL", zap.Error(err))
	}
	cacheClient := cache.NewWithOptions(opt, 5*time.Minute)

	logger.Info("Initialized Redis client", zap.String("addr", cfg.RedisAddr))
	defer cacheClient.Close()

	// Test Redis write
	err = cacheClient.Set(context.Background(), "test_key", "123", time.Minute)
	log.Println("Redis SET error:", err)

	// RabbitMQ publisher
	pub, err := mq.NewPublisher(cfg.RabbitURL, "voucher.events")
	if err != nil {
		logger.Fatal("rabbitmq connection failed", zap.Error(err))
	}
	logger.Info("Connected to RabbitMQ", zap.String("url", cfg.RabbitURL))
	defer pub.Close()

	// Initialize voucher service
	voucherSvc := service.NewVoucherService(dbConn, cacheClient, pub, logger)

	// Create gRPC server
	grpcServer := transport.NewGRPCServer(voucherSvc, logger)
	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Fatal("listen", zap.Error(err))
	}

	// ---------- NEW HTTP SERVER (LOGIN + VOUCHERS) ----------
	httpSrv := transport.NewHTTPServer(voucherSvc, logger)
	httpHandler := withCORS(httpSrv.Handler())

	// Attach routes including pprof
	mux := http.NewServeMux()
	mux.Handle("/", httpHandler)
	mux.HandleFunc("/debug/pprof/", pp.Index)
	mux.HandleFunc("/debug/pprof/profile", pp.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pp.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pp.Trace)

	httpServer := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: mux,
	}

	// Start gRPC server
	go func() {
		logger.Info("Starting gRPC server", zap.String("port", cfg.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("grpc serve failed", zap.Error(err))
		}
	}()

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server", zap.String("port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http serve failed", zap.Error(err))
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownGrace)
	defer cancel()

	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP shutdown failed", zap.Error(err))
	}

	logger.Info("Shutdown complete")
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
