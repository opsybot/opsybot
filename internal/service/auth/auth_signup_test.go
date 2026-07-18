package auth

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/workspace"
)

func TestUniqueWorkspaceSlug(t *testing.T) {
	ctx := context.Background()

	t.Run("free slug is used as-is", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ws := workspace.NewMockWorkspace(ctrl)
		ws.EXPECT().GetBySlug(gomock.Any(), "acme-corp").Return(entity.Workspace{}, entity.ErrWorkspaceNotFound)

		got, err := (&srv{workspaces: ws}).uniqueWorkspaceSlug(ctx, "Acme Corp")
		if err != nil || got != "acme-corp" {
			t.Fatalf("got %q, err %v; want acme-corp", got, err)
		}
	})

	t.Run("taken slug is deduped with a suffix", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ws := workspace.NewMockWorkspace(ctrl)
		ws.EXPECT().GetBySlug(gomock.Any(), "acme-corp").Return(entity.Workspace{ID: "w1"}, nil)
		ws.EXPECT().GetBySlug(gomock.Any(), "acme-corp-2").Return(entity.Workspace{}, entity.ErrWorkspaceNotFound)

		got, err := (&srv{workspaces: ws}).uniqueWorkspaceSlug(ctx, "Acme Corp")
		if err != nil || got != "acme-corp-2" {
			t.Fatalf("got %q, err %v; want acme-corp-2", got, err)
		}
	})

	t.Run("reserved slug is skipped", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ws := workspace.NewMockWorkspace(ctrl)
		ws.EXPECT().GetBySlug(gomock.Any(), "login-2").Return(entity.Workspace{}, entity.ErrWorkspaceNotFound)

		got, err := (&srv{workspaces: ws}).uniqueWorkspaceSlug(ctx, "login")
		if err != nil || got != "login-2" {
			t.Fatalf("got %q, err %v; want login-2 (reserved 'login' skipped)", got, err)
		}
	})
}
