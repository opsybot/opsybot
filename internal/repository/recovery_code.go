package repository

//go:generate go tool mockgen -source=recovery_code.go -destination=./recovery_code/recovery_code_mock.go -package=recovery_code

import "context"

type RecoveryCode interface {
	Replace(ctx context.Context, userID string, codeHashes []string) error
	ListUnusedHashes(ctx context.Context, userID string) ([]string, error)
	MarkUsed(ctx context.Context, userID, codeHash string) (bool, error)
	CountUnused(ctx context.Context, userID string) (int, error)
}
