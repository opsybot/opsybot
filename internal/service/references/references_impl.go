package references

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	sources []service.ReferenceSource
}

func New(sources []service.ReferenceSource) service.References {
	return &srv{sources: sources}
}

func (s *srv) ListByUser(ctx context.Context, workspaceID, userID string) ([]entity.MemberReference, error) {
	var all []entity.MemberReference
	for _, src := range s.sources {
		refs, err := src.ListByUser(ctx, workspaceID, userID)
		if err != nil {
			return nil, err
		}
		all = append(all, refs...)
	}
	return all, nil
}

func (s *srv) ReassignAll(ctx context.Context, workspaceID, userID string, replacements map[string]string) error {
	for _, src := range s.sources {
		refs, err := src.ListByUser(ctx, workspaceID, userID)
		if err != nil {
			return err
		}
		for _, ref := range refs {
			to, ok := replacements[ref.ID]
			if !ok {
				return entity.ErrMemberReplacementsIncomplete
			}
			if err := src.Reassign(ctx, workspaceID, ref.ID, to); err != nil {
				return err
			}
		}
	}
	return nil
}
