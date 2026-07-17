package sso

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/repository/audit"
	"github.com/opsybot/opsybot/internal/repository/member"
	"github.com/opsybot/opsybot/internal/repository/policy"
	"github.com/opsybot/opsybot/internal/repository/user"
	"github.com/opsybot/opsybot/internal/repository/user_identity"
)

type provisionMocks struct {
	users      *user.MockUser
	members    *member.MockMember
	policy     *policy.MockPolicy
	audit      *audit.MockAudit
	identities *user_identity.MockUserIdentity
}

func newProvisionSrv(ctrl *gomock.Controller) (*srv, provisionMocks) {
	m := provisionMocks{
		users:      user.NewMockUser(ctrl),
		members:    member.NewMockMember(ctrl),
		policy:     policy.NewMockPolicy(ctrl),
		audit:      audit.NewMockAudit(ctrl),
		identities: user_identity.NewMockUserIdentity(ctrl),
	}
	return &srv{users: m.users, members: m.members, policy: m.policy, audit: m.audit, identities: m.identities}, m
}

func TestProvisionNewUserJITDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	s, m := newProvisionSrv(ctrl)
	ws := entity.Workspace{ID: "ws1"}
	conn := entity.SSOConnection{ID: "c1", JITProvisioning: false}

	m.identities.EXPECT().GetBySubject(gomock.Any(), "c1", "sub-1").Return(entity.UserIdentity{}, entity.ErrUserIdentityNotFound)
	m.users.EXPECT().GetByEmail(gomock.Any(), "new@acme.dev").Return(entity.User{}, entity.ErrUserNotFound)

	_, assigned, err := s.provision(context.Background(), ws, conn, "sub-1", "new@acme.dev", "New")
	if err != entity.ErrSSOProvisioningDisabled {
		t.Fatalf("err = %v, want ErrSSOProvisioningDisabled", err)
	}
	if assigned {
		t.Error("assigned should be false when nothing was provisioned")
	}
}

func TestProvisionNewUserDomainNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	s, m := newProvisionSrv(ctrl)
	ws := entity.Workspace{ID: "ws1"}
	conn := entity.SSOConnection{ID: "c1", JITProvisioning: true, AllowedEmailDomains: []string{"acme.dev"}}

	m.identities.EXPECT().GetBySubject(gomock.Any(), "c1", "sub-1").Return(entity.UserIdentity{}, entity.ErrUserIdentityNotFound)
	m.users.EXPECT().GetByEmail(gomock.Any(), "intruder@evil.com").Return(entity.User{}, entity.ErrUserNotFound)

	_, _, err := s.provision(context.Background(), ws, conn, "sub-1", "intruder@evil.com", "Intruder")
	if err != entity.ErrSSODomainNotAllowed {
		t.Fatalf("err = %v, want ErrSSODomainNotAllowed", err)
	}
}

func TestProvisionNewUserJITCreates(t *testing.T) {
	ctrl := gomock.NewController(t)
	s, m := newProvisionSrv(ctrl)
	ws := entity.Workspace{ID: "ws1"}
	conn := entity.SSOConnection{ID: "c1", JITProvisioning: true, AllowedEmailDomains: []string{"acme.dev"}}
	created := entity.User{ID: "u1", Email: "new@acme.dev", Name: "New"}

	m.identities.EXPECT().GetBySubject(gomock.Any(), "c1", "sub-1").Return(entity.UserIdentity{}, entity.ErrUserIdentityNotFound)
	m.users.EXPECT().GetByEmail(gomock.Any(), "new@acme.dev").Return(entity.User{}, entity.ErrUserNotFound)
	m.users.EXPECT().CreateSSO(gomock.Any(), "new@acme.dev", "New").Return(created, nil)
	m.members.EXPECT().Create(gomock.Any(), "ws1", "u1", entity.MemberStatusActive).Return(nil)
	m.policy.EXPECT().AssignRole(gomock.Any(), "u1", "ws1", entity.RoleMember).Return(nil)
	m.audit.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	m.identities.EXPECT().Create(gomock.Any(), "u1", "c1", "sub-1", "new@acme.dev").Return(nil)

	user, assigned, err := s.provision(context.Background(), ws, conn, "sub-1", "new@acme.dev", "New")
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if user.ID != "u1" || !assigned {
		t.Errorf("user=%+v assigned=%v, want u1/true", user, assigned)
	}
}

func TestProvisionReturningUserByIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	s, m := newProvisionSrv(ctrl)
	ws := entity.Workspace{ID: "ws1"}
	conn := entity.SSOConnection{ID: "c1"}

	m.identities.EXPECT().GetBySubject(gomock.Any(), "c1", "sub-9").Return(entity.UserIdentity{UserID: "u9"}, nil)
	m.users.EXPECT().GetByID(gomock.Any(), "u9").Return(entity.User{ID: "u9"}, nil)
	m.members.EXPECT().Get(gomock.Any(), "ws1", "u9").Return(entity.Member{UserID: "u9", Status: entity.MemberStatusActive}, nil)

	user, assigned, err := s.provision(context.Background(), ws, conn, "sub-9", "known@acme.dev", "Known")
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if user.ID != "u9" || assigned {
		t.Errorf("user=%+v assigned=%v, want u9/false", user, assigned)
	}
}
