package service

//go:generate go tool mockgen -source=references.go -destination=./references/references_mock.go -package=references

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
)

type ReferenceSource interface {
	ListByUser(ctx context.Context, workspaceID, userID string) ([]entity.MemberReference, error)
	Reassign(ctx context.Context, workspaceID, referenceID, toUserID string) error
}

type References interface {
	ListByUser(ctx context.Context, workspaceID, userID string) ([]entity.MemberReference, error)
	ReassignAll(ctx context.Context, workspaceID, userID string, replacements map[string]string) error
}
