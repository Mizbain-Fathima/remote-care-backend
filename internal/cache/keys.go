package cache

import "fmt"

// Search results key
func SearchKey(category, query string, page, pageSize int32) string {
	return fmt.Sprintf("search:%s:%s:%d:%d", category, query, page, pageSize)
}

// Single voucher key
func VoucherKey(id string) string {
	return "voucher:" + id
}

// Wallet balance key
func WalletKey(userID string) string {
	return "wallet:" + userID
}
