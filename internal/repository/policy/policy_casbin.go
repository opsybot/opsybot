package policy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/casbin"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/repository"
)

const subjectPrefix = "user:"

type repo struct {
	enforcer casbin.Client
	db       postgres.Client
}

func New(enforcer casbin.Client, db postgres.Client) repository.Policy {
	return &repo{enforcer: enforcer, db: db}
}

func userSubject(userID string) string { return subjectPrefix + userID }

func (r *repo) Allowed(ctx context.Context, subject, workspaceID string, obj entity.PolicyObject, act entity.PolicyAction) (bool, error) {
	ok, err := r.enforcer.Enforce(subject, workspaceID, string(obj), string(act))
	if err != nil {
		return false, fmt.Errorf("enforce: %w", err)
	}
	return ok, nil
}

func (r *repo) RoleOf(ctx context.Context, userID, workspaceID string) (entity.Role, bool, error) {
	roles := r.enforcer.GetRolesForUserInDomain(userSubject(userID), workspaceID)
	if len(roles) == 0 {
		return "", false, nil
	}
	return entity.Role(roles[0]), true, nil
}

func (r *repo) RolesByWorkspace(ctx context.Context, workspaceID string) (map[string]entity.Role, error) {
	out := make(map[string]entity.Role)
	for _, role := range []entity.Role{entity.RoleAdmin, entity.RoleMember} {
		for _, subject := range r.enforcer.GetUsersForRoleInDomain(string(role), workspaceID) {
			if id, ok := strings.CutPrefix(subject, subjectPrefix); ok {
				out[id] = role
			}
		}
	}
	return out, nil
}

func (r *repo) AssignRole(ctx context.Context, userID, workspaceID string, role entity.Role) error {
	if _, err := r.enforcer.AddRoleForUserInDomain(userSubject(userID), string(role), workspaceID); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}
	return nil
}

func (r *repo) ReplaceRole(ctx context.Context, userID, workspaceID string, from, to entity.Role) error {
	if _, err := r.enforcer.AddRoleForUserInDomain(userSubject(userID), string(to), workspaceID); err != nil {
		return fmt.Errorf("replace role add: %w", err)
	}
	if _, err := r.enforcer.DeleteRoleForUserInDomain(userSubject(userID), string(from), workspaceID); err != nil {
		if _, rerr := r.enforcer.DeleteRoleForUserInDomain(userSubject(userID), string(to), workspaceID); rerr != nil {
			return fmt.Errorf("replace role delete (restore failed: %v): %w", rerr, err)
		}
		return fmt.Errorf("replace role delete: %w", err)
	}
	return nil
}

func (r *repo) RemoveRole(ctx context.Context, userID, workspaceID string) error {
	role, ok, err := r.RoleOf(ctx, userID, workspaceID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if _, err := r.enforcer.DeleteRoleForUserInDomain(userSubject(userID), string(role), workspaceID); err != nil {
		return fmt.Errorf("remove role: %w", err)
	}
	return nil
}

func (r *repo) AssignAgentRole(ctx context.Context, workspaceID string, role entity.Role) error {
	if _, err := r.enforcer.AddRoleForUserInDomain("wsagent:"+workspaceID, string(role), workspaceID); err != nil {
		return fmt.Errorf("assign agent role: %w", err)
	}
	return nil
}

func (r *repo) SeedWorkspace(ctx context.Context, workspaceID string) error {
	var rules [][]string
	for _, role := range []entity.Role{entity.RoleAdmin, entity.RoleMember} {
		for _, rule := range entity.RolePolicies(role) {
			rules = append(rules, []string{string(role), workspaceID, string(rule.Object), string(rule.Action)})
		}
	}
	if _, err := r.enforcer.AddPoliciesEx(rules); err != nil {
		return fmt.Errorf("seed workspace policies: %w", err)
	}
	return nil
}

func (r *repo) RoleOfTx(ctx context.Context, userID, workspaceID string) (entity.Role, bool, error) {
	var role string
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT v1 FROM casbin_rule WHERE p_type = 'g' AND v0 = $1 AND v2 = $2`,
		userSubject(userID), workspaceID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("role of tx: %w", err)
	}
	return entity.Role(role), true, nil
}

func (r *repo) CountActiveAdminsTx(ctx context.Context, workspaceID string) (int, error) {
	var count int
	err := r.db.Querier(ctx).QueryRowContext(ctx,
		`SELECT count(*) FROM casbin_rule r
		 JOIN workspace_members m
		   ON ('user:' || m.user_id::text) = r.v0 AND m.workspace_id::text = r.v2
		 WHERE r.p_type = 'g' AND r.v1 = 'admin' AND m.workspace_id = $1 AND m.status = 'active'`,
		workspaceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active admins: %w", err)
	}
	return count, nil
}
