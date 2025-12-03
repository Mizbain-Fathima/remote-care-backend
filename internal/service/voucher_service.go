package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	pb "github.com/Mizbain-Fathima/remote-care-backend/gen"
	"github.com/Mizbain-Fathima/remote-care-backend/internal/cache"
	"github.com/Mizbain-Fathima/remote-care-backend/internal/db"
	"github.com/Mizbain-Fathima/remote-care-backend/internal/mq"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type VoucherService struct {
	pb.UnimplementedVoucherServiceServer
	db        *db.DB
	cache     *cache.Cache
	publisher *mq.Publisher
	logger    *zap.Logger
}

func NewVoucherService(
	db *db.DB,
	cache *cache.Cache,
	pub *mq.Publisher,
	logger *zap.Logger,
) *VoucherService {
	return &VoucherService{
		db:        db,
		cache:     cache,
		publisher: pub,
		logger:    logger,
	}
}

// SearchVouchers uses Redis lazy cache
func (s *VoucherService) SearchVouchers(ctx context.Context, req *pb.SearchVouchersRequest) (*pb.SearchVouchersResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	key := fmt.Sprintf("search:%s:%s:%d:%d", req.Query, req.Category, req.Page, req.PageSize)

	if val, err := s.cache.Get(ctx, key); err == nil {
		var resp pb.SearchVouchersResponse
		if err := json.Unmarshal([]byte(val), &resp); err == nil {
			return &resp, nil
		}
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, name, description, category, brand,
		       discount_percent, price, currency, active
		FROM vouchers
		WHERE active = true
		  AND ($1 = '' OR category = $1)
		  AND ($2 = '' OR name ILIKE '%' || $2 || '%')
		LIMIT $3 OFFSET $4`,
		req.Category,
		req.Query,
		req.PageSize,
		(req.Page-1)*req.PageSize,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vouchers []*pb.Voucher
	for rows.Next() {
		var v pb.Voucher
		if err := rows.Scan(
			&v.Id, &v.Name, &v.Description, &v.Category,
			&v.Brand, &v.DiscountPercent, &v.Price, &v.Currency, &v.Active,
		); err != nil {
			return nil, err
		}
		vouchers = append(vouchers, &v)
	}

	resp := &pb.SearchVouchersResponse{
		Vouchers: vouchers,
		Total:    int32(len(vouchers)), // simple total
	}

	if b, err := json.Marshal(resp); err == nil {
		_ = s.cache.Set(ctx, key, string(b), 5*time.Minute)
	}

	return resp, nil
}

// GetBalance returns wallet balance or zero if no wallet
func (s *VoucherService) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	row := s.db.Pool.QueryRow(ctx, `
		SELECT balance, currency
		FROM wallets
		WHERE user_id = $1`,
		req.UserId,
	)

	var balance float64
	var currency string
	if err := row.Scan(&balance, &currency); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &pb.GetBalanceResponse{
				UserId:   req.UserId,
				Balance:  0,
				Currency: "INR",
			}, nil
		}
		return nil, err
	}

	return &pb.GetBalanceResponse{
		UserId:   req.UserId,
		Balance:  balance,
		Currency: currency,
	}, nil
}

// BuyVoucher uses ACID transaction + idempotency + RabbitMQ event
func (s *VoucherService) BuyVoucher(ctx context.Context, req *pb.BuyVoucherRequest) (*pb.BuyVoucherResponse, error) {
	// mock UPI validation
	if req.PaymentMethod != "MOCK_UPI" || !isValidUPI(req.UpiId) {
		return &pb.BuyVoucherResponse{
			Status: "FAILED",
		}, nil
	}

	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var txnID, status string

	// idempotency: request_id unique
	err = tx.QueryRow(ctx, `
		SELECT id, status
		FROM transactions
		WHERE request_id = $1`, req.RequestId).Scan(&txnID, &status)

	if err == nil {
		// already processed
		return &pb.BuyVoucherResponse{
			TransactionId: txnID,
			Status:        status,
		}, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// voucher details
	var voucherPrice float64
	var currency string
	err = tx.QueryRow(ctx, `
		SELECT price, currency
		FROM vouchers
		WHERE id = $1 AND active = true AND stock > 0
		FOR UPDATE`,
		req.VoucherId,
	).Scan(&voucherPrice, &currency)
	if err != nil {
		return nil, fmt.Errorf("voucher not available: %w", err)
	}

	// wallet balance
	var balance float64
	err = tx.QueryRow(ctx, `
		SELECT balance
		FROM wallets
		WHERE user_id = $1
		FOR UPDATE`,
		req.UserId,
	).Scan(&balance)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	if balance < voucherPrice {
		status = "FAILED"
		err = tx.QueryRow(ctx, `
			INSERT INTO transactions (user_id, voucher_id, amount, currency, status, payment_method, upi_id, request_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id`,
			req.UserId, req.VoucherId, voucherPrice, currency, status,
			req.PaymentMethod, req.UpiId, req.RequestId,
		).Scan(&txnID)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return &pb.BuyVoucherResponse{TransactionId: txnID, Status: status}, nil
	}

	// deduct balance & stock
	_, err = tx.Exec(ctx, `
		UPDATE wallets
		SET balance = balance - $1, updated_at = now()
		WHERE user_id = $2`, voucherPrice, req.UserId)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE vouchers
		SET stock = stock - 1
		WHERE id = $1`, req.VoucherId)
	if err != nil {
		return nil, err
	}

	status = "SUCCESS"
	err = tx.QueryRow(ctx, `
		INSERT INTO transactions (user_id, voucher_id, amount, currency, status, payment_method, upi_id, request_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id`,
		req.UserId, req.VoucherId, voucherPrice, currency, status,
		req.PaymentMethod, req.UpiId, req.RequestId,
	).Scan(&txnID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// publish event (fire-and-forget)
	event := map[string]any{
		"transaction_id": txnID,
		"user_id":        req.UserId,
		"voucher_id":     req.VoucherId,
		"amount":         voucherPrice,
		"currency":       currency,
		"status":         status,
		"payment_method": req.PaymentMethod,
		"upi_id":         req.UpiId,
	}
	if b, err := json.Marshal(event); err == nil {
		_ = s.publisher.PublishPurchase(ctx, "purchase.completed", b)
	}

	return &pb.BuyVoucherResponse{
		TransactionId: txnID,
		Status:        status,
	}, nil
}

func (s *VoucherService) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, user_id, voucher_id, amount, currency, status, created_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		req.UserId,
		req.PageSize,
		(req.Page-1)*req.PageSize,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []*pb.Transaction
	for rows.Next() {
		var t pb.Transaction
		if err := rows.Scan(
			&t.Id, &t.UserId, &t.VoucherId,
			&t.Amount, &t.Currency, &t.Status, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		txns = append(txns, &t)
	}
	return &pb.ListTransactionsResponse{
		Transactions: txns,
		Total:        int32(len(txns)),
	}, nil
}

// simple mock: contains "@upi"
func isValidUPI(id string) bool {
	return len(id) > 4 && (id[len(id)-4:] == "@upi" || id[len(id)-8:] == "@okhdfcb")
}
