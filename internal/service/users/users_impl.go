package users

import (
	"context"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository"
	"github.com/opsybot/opsybot/internal/service"
)

type srv struct {
	tx       repository.Transactor
	users    repository.User
	recovery repository.RecoveryCode
	sessions repository.Session
}

func New(tx repository.Transactor, users repository.User, recovery repository.RecoveryCode, sessions repository.Session) service.Users {
	return &srv{tx: tx, users: users, recovery: recovery, sessions: sessions}
}

func (s *srv) identity(ctx context.Context) (entity.Identity, error) {
	id, ok := entity.IdentityFrom(ctx)
	if !ok || id.Kind != entity.IdentityKindSession {
		return entity.Identity{}, entity.ErrUnauthenticated
	}
	return id, nil
}

func (s *srv) UpdateProfile(ctx context.Context, in entity.ProfileUpdate) (entity.User, error) {
	id, err := s.identity(ctx)
	if err != nil {
		return entity.User{}, err
	}
	if err := in.Validate(); err != nil {
		return entity.User{}, err
	}
	if err := s.users.UpdateProfile(ctx, id.UserID, in); err != nil {
		return entity.User{}, err
	}
	return s.users.GetByID(ctx, id.UserID)
}

func (s *srv) ChangePassword(ctx context.Context, current, next string) error {
	id, err := s.identity(ctx)
	if err != nil {
		return err
	}
	if err := entity.ValidatePassword(next); err != nil {
		return err
	}
	hash, err := s.users.PasswordHash(ctx, id.UserID)
	if err != nil {
		return err
	}
	if err := entity.VerifyPassword(hash, current); err != nil {
		return entity.ErrInvalidCredentials
	}
	newHash, err := entity.HashPassword(next)
	if err != nil {
		return err
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.users.UpdatePassword(ctx, id.UserID, newHash); err != nil {
			return err
		}
		return s.sessions.DeleteOthers(ctx, id.UserID, id.SessionID)
	})
}

func (s *srv) BeginTOTP(ctx context.Context) (entity.TOTPEnrollment, error) {
	id, err := s.identity(ctx)
	if err != nil {
		return entity.TOTPEnrollment{}, err
	}
	user, err := s.users.GetByID(ctx, id.UserID)
	if err != nil {
		return entity.TOTPEnrollment{}, err
	}
	if user.TOTPEnabled {
		return entity.TOTPEnrollment{}, entity.ErrTOTPAlreadySetUp
	}
	generated, err := totp.Generate(totp.GenerateOpts{Issuer: entity.TOTPIssuer, AccountName: user.Email})
	if err != nil {
		return entity.TOTPEnrollment{}, err
	}
	if err := s.users.SetTOTP(ctx, id.UserID, generated.Secret()); err != nil {
		return entity.TOTPEnrollment{}, err
	}
	return entity.TOTPEnrollment{Secret: generated.Secret(), OTPAuthURI: generated.URL()}, nil
}

func (s *srv) ConfirmTOTP(ctx context.Context, code string) ([]string, error) {
	id, err := s.identity(ctx)
	if err != nil {
		return nil, err
	}
	if !entity.ValidTOTPCode(code) {
		return nil, entity.ErrTOTPInvalidCode
	}
	secret, err := s.users.TOTPSecret(ctx, id.UserID)
	if err != nil {
		return nil, err
	}
	if !totp.Validate(code, secret) {
		return nil, entity.ErrTOTPInvalidCode
	}
	return s.enableWithRecovery(ctx, id, func(ctx context.Context) error {
		return s.users.EnableTOTP(ctx, id.UserID)
	})
}

func (s *srv) RegenerateRecoveryCodes(ctx context.Context, code string) ([]string, error) {
	id, err := s.identity(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.requireTOTP(ctx, id.UserID, code); err != nil {
		return nil, err
	}
	return s.enableWithRecovery(ctx, id, nil)
}

func (s *srv) enableWithRecovery(ctx context.Context, id entity.Identity, before func(context.Context) error) ([]string, error) {
	codes, err := entity.GenerateRecoveryCodes(entity.RecoveryCodeCount)
	if err != nil {
		return nil, err
	}
	hashes := make([]string, 0, len(codes))
	for _, c := range codes {
		h, err := entity.HashPassword(c)
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, h)
	}
	err = s.tx.WithTx(ctx, func(ctx context.Context) error {
		if before != nil {
			if err := before(ctx); err != nil {
				return err
			}
		}
		if err := s.recovery.Replace(ctx, id.UserID, hashes); err != nil {
			return err
		}
		return s.sessions.DeleteOthers(ctx, id.UserID, id.SessionID)
	})
	if err != nil {
		return nil, err
	}
	return codes, nil
}

func (s *srv) DisableTOTP(ctx context.Context, code string) error {
	id, err := s.identity(ctx)
	if err != nil {
		return err
	}
	if err := s.requireTOTP(ctx, id.UserID, code); err != nil {
		return err
	}
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		if err := s.users.DisableTOTP(ctx, id.UserID); err != nil {
			return err
		}
		if err := s.recovery.Replace(ctx, id.UserID, nil); err != nil {
			return err
		}
		return s.sessions.DeleteOthers(ctx, id.UserID, id.SessionID)
	})
}

func (s *srv) requireTOTP(ctx context.Context, userID, code string) error {
	if !entity.ValidTOTPCode(code) {
		return entity.ErrTOTPInvalidCode
	}
	secret, err := s.users.TOTPSecret(ctx, userID)
	if err != nil {
		return err
	}
	if !totp.Validate(code, secret) {
		return entity.ErrTOTPInvalidCode
	}
	step := time.Now().Unix() / entity.TOTPPeriodSeconds
	accepted, err := s.users.AcceptTOTPStep(ctx, userID, step)
	if err != nil {
		return err
	}
	if !accepted {
		return entity.ErrTOTPInvalidCode
	}
	return nil
}

func (s *srv) ListSessions(ctx context.Context) ([]entity.Session, error) {
	id, err := s.identity(ctx)
	if err != nil {
		return nil, err
	}
	return s.sessions.ListByUser(ctx, id.UserID)
}

func (s *srv) RevokeSession(ctx context.Context, sessionID string) error {
	id, err := s.identity(ctx)
	if err != nil {
		return err
	}
	owned, err := s.sessions.OwnedBy(ctx, sessionID, id.UserID)
	if err != nil {
		return err
	}
	if !owned {
		return entity.ErrSessionNotFound
	}
	return s.sessions.Delete(ctx, sessionID)
}
