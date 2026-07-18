package channels

import (
	"context"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	channels repository.Channel
	audit    repository.Audit
}

func New(channels repository.Channel, audit repository.Audit) service.Channels {
	return &srv{channels: channels, audit: audit}
}

func (s *srv) userID(ctx context.Context) (string, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok || id.Kind != entity.IdentityKindSession {
		return "", entity.ErrUnauthenticated
	}
	return id.UserID, nil
}

func (s *srv) List(ctx context.Context) ([]entity.Channel, error) {
	userID, err := s.userID(ctx)
	if err != nil {
		return nil, err
	}
	return s.channels.ListByUser(ctx, userID)
}

func (s *srv) Add(ctx context.Context, in entity.NewChannel) (entity.Channel, error) {
	userID, err := s.userID(ctx)
	if err != nil {
		return entity.Channel{}, err
	}
	if err := in.Validate(); err != nil {
		return entity.Channel{}, err
	}
	ch, err := s.channels.Create(ctx, userID, in)
	if err != nil {
		return entity.Channel{}, err
	}
	_ = s.audit.Create(ctx, entity.AuditEvent{
		ActorType: entity.AuditActorUser, ActorUserID: userID,
		Action: entity.ActionChannelAdded, Target: string(ch.Type),
	})
	return ch, nil
}

func (s *srv) Verify(ctx context.Context, channelID string) error {
	userID, err := s.userID(ctx)
	if err != nil {
		return err
	}
	return s.channels.MarkVerified(ctx, channelID, userID)
}

func (s *srv) Remove(ctx context.Context, channelID string) error {
	userID, err := s.userID(ctx)
	if err != nil {
		return err
	}
	if err := s.channels.Delete(ctx, channelID, userID); err != nil {
		return err
	}
	_ = s.audit.Create(ctx, entity.AuditEvent{
		ActorType: entity.AuditActorUser, ActorUserID: userID,
		Action: entity.ActionChannelRemoved, Target: channelID,
	})
	return nil
}
