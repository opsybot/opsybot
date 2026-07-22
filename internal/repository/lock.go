package repository

//go:generate go tool mockgen -source=lock.go -destination=./lock/lock_mock.go -package=lock

import "context"

type Lock interface {
	Workspace(ctx context.Context, workspaceID string) error
	Instance(ctx context.Context) error
	TryJob(ctx context.Context, name string) (bool, error)
}
