package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	pb "github.com/Mizbain-Fathima/remote-care-backend/gen/github.com/Mizbain-Fathima/remote-care-backend/gen"

	"github.com/Mizbain-Fathima/remote-care-backend/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// ----------------------
// Request / Response DTO
// ----------------------
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"userId"`
	Email  string `json:"email"`
}

// ----------------------
// HTTP Server Struct
// ----------------------
type HTTPServer struct {
	mux    *http.ServeMux
	svc    *service.VoucherService
	logger *zap.Logger
}

// ----------------------
// JWT Secret Key
// ----------------------
var jwtKey = []byte("SUPER_SECRET_KEY_CHANGE_ME")

// ----------------------
// Login Handler
// ----------------------
func (s *HTTPServer) handleLogin(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email required", http.StatusBadRequest)
		return
	}

	// Static user (matches your frontend)
	userId := "0024fd90-b092-4a5e-aee6-d48d10d8a66e"

	// Create JWT
	claims := jwt.MapClaims{
		"userId": userId,
		"email":  req.Email,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Token error", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		Token:  signed,
		UserID: userId,
		Email:  req.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ------------------------------
// GET BALANCE
// ------------------------------
func (s *HTTPServer) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}

	resp, err := s.svc.GetBalance(r.Context(), &pb.GetBalanceRequest{UserId: userId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// ------------------------------
// SEARCH VOUCHERS
// ------------------------------
func (s *HTTPServer) handleSearchVouchers(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	query := r.URL.Query().Get("query")
	page := r.URL.Query().Get("page")
	size := r.URL.Query().Get("page_size")

	req := &pb.SearchVouchersRequest{
		Category: category,
		Query:    query,
		Page:     parseInt(page, 1),
		PageSize: parseInt(size, 20),
	}

	resp, err := s.svc.SearchVouchers(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// ------------------------------
// LIST TRANSACTIONS
// ------------------------------
func (s *HTTPServer) handleListTransactions(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")
	page := parseInt(r.URL.Query().Get("page"), 1)
	size := parseInt(r.URL.Query().Get("page_size"), 20)

	req := &pb.ListTransactionsRequest{
		UserId:   userId,
		Page:     page,
		PageSize: size,
	}

	resp, err := s.svc.ListTransactions(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// ------------------------------
// BUY VOUCHER
// ------------------------------
func (s *HTTPServer) handleBuyVoucher(w http.ResponseWriter, r *http.Request) {
	var req pb.BuyVoucherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	resp, err := s.svc.BuyVoucher(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, resp)
}

// ------------------------------
// Helper: JSON writer
// ------------------------------
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// Helper: safe integer parser
func parseInt(s string, def int32) int32 {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return int32(n)
}

func NewHTTPServer(svc *service.VoucherService, logger *zap.Logger) *HTTPServer {
	mux := http.NewServeMux()
	s := &HTTPServer{mux: mux, svc: svc, logger: logger}

	// REGISTER ALL ENDPOINTS HERE
	mux.HandleFunc("/v1/login", s.handleLogin)
	mux.HandleFunc("/v1/getbalance", s.handleGetBalance)
	mux.HandleFunc("/v1/searchvouchers", s.handleSearchVouchers)
	mux.HandleFunc("/v1/listtransactions", s.handleListTransactions)
	mux.HandleFunc("/v1/buyvoucher", s.handleBuyVoucher)

	return s
}

func (s *HTTPServer) Handler() http.Handler {
	return s.mux
}
