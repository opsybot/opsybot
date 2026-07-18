package auth

import (
	"context"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/workspace"
)

func TestCheckSlug(t *testing.T) {
	ctx := context.Background()

	t.Run("free slug is available with no suggestion", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ws := workspace.NewMockWorkspace(ctrl)
		ws.EXPECT().GetBySlug(gomock.Any(), "acme").Return(entity.Workspace{}, entity.ErrWorkspaceNotFound)

		available, suggestion, err := (&srv{workspaces: ws}).CheckSlug(ctx, "acme")
		if err != nil || !available || suggestion != "" {
			t.Fatalf("got available=%v suggestion=%q err=%v; want available=true, no suggestion", available, suggestion, err)
		}
	})

	t.Run("taken slug suggests a valid random alternative", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ws := workspace.NewMockWorkspace(ctrl)
		ws.EXPECT().GetBySlug(gomock.Any(), "acme").Return(entity.Workspace{ID: "w1"}, nil)
		ws.EXPECT().GetBySlug(gomock.Any(), gomock.Not("acme")).Return(entity.Workspace{}, entity.ErrWorkspaceNotFound)

		available, suggestion, err := (&srv{workspaces: ws}).CheckSlug(ctx, "acme")
		if err != nil || available {
			t.Fatalf("got available=%v err=%v; want unavailable", available, err)
		}
		if !strings.HasPrefix(suggestion, "acme-") || !entity.ValidSlugFormat(suggestion) {
			t.Fatalf("suggestion %q is not a valid acme- alternative", suggestion)
		}
	})

	t.Run("reserved slug is unavailable and suggested around", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ws := workspace.NewMockWorkspace(ctrl)
		ws.EXPECT().GetBySlug(gomock.Any(), gomock.Any()).Return(entity.Workspace{}, entity.ErrWorkspaceNotFound)

		available, suggestion, err := (&srv{workspaces: ws}).CheckSlug(ctx, "login")
		if err != nil || available {
			t.Fatalf("got available=%v err=%v; want unavailable", available, err)
		}
		if !strings.HasPrefix(suggestion, "login-") {
			t.Fatalf("suggestion %q should extend the reserved base", suggestion)
		}
	})

	t.Run("malformed slug is rejected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ws := workspace.NewMockWorkspace(ctrl)

		_, _, err := (&srv{workspaces: ws}).CheckSlug(ctx, "Acme Corp")
		if err != entity.ErrWorkspaceSlugInvalid {
			t.Fatalf("got err=%v; want ErrWorkspaceSlugInvalid", err)
		}
	})
}
