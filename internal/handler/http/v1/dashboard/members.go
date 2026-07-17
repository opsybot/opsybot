package dashboard

import (
	"context"
	"errors"
	"net/http"

	"github.com/opsybot/opsybot/internal/entity"
	api "github.com/opsybot/opsybot/pkg/http/v1/dashboard"
)

func (h *handler) ListMembers(ctx context.Context, request api.ListMembersRequestObject) (api.ListMembersResponseObject, error) {
	list, err := h.members.List(ctx, request.WorkspaceId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.ListMembers403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "You don't have permission to view members.", "")), nil
		case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.ListMembers404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such workspace.", "")), nil
		default:
			return nil, err
		}
	}
	items := make([]api.Member, 0, len(list))
	for _, m := range list {
		items = append(items, memberDTO(m))
	}
	return api.ListMembers200JSONResponse{Items: items}, nil
}

func (h *handler) GetMember(ctx context.Context, request api.GetMemberRequestObject) (api.GetMemberResponseObject, error) {
	m, err := h.members.Get(ctx, request.WorkspaceId, request.UserId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.GetMember403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "You don't have permission to view members.", "")), nil
		case errors.Is(err, entity.ErrMemberNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.GetMember404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such member.", "")), nil
		default:
			return nil, err
		}
	}
	return api.GetMember200JSONResponse(memberDTO(m)), nil
}

func (h *handler) ListInvites(ctx context.Context, request api.ListInvitesRequestObject) (api.ListInvitesResponseObject, error) {
	list, err := h.members.ListInvites(ctx, request.WorkspaceId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.ListInvites403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can view invitations.", "")), nil
		case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.ListInvites404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such workspace.", "")), nil
		default:
			return nil, err
		}
	}
	items := make([]api.Invite, 0, len(list))
	for _, inv := range list {
		items = append(items, inviteDTO(inv))
	}
	return api.ListInvites200JSONResponse{Items: items}, nil
}

func (h *handler) InviteMember(ctx context.Context, request api.InviteMemberRequestObject) (api.InviteMemberResponseObject, error) {
	if request.Body == nil {
		return api.InviteMember400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	inv, acceptURL, err := h.members.Invite(ctx, request.WorkspaceId, request.Body.Email, entity.Role(request.Body.Role))
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.InviteMember403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can invite members.", "")), nil
		case errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.InviteMember404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such workspace.", "")), nil
		case errors.Is(err, entity.ErrMemberAlreadyExists):
			return api.InviteMember409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Already a member", "That person is already a member of this workspace.", "")), nil
		case errors.Is(err, entity.ErrInvitePending), errors.Is(err, entity.ErrMemberDeactivated):
			return api.InviteMember409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Invite exists", "That email already has a pending or deactivated membership here.", "")), nil
		case isValidation(err), errors.Is(err, entity.ErrRoleInvalid):
			return api.InviteMember400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid invitation", validationDetail(err), "")), nil
		default:
			return nil, err
		}
	}
	return api.InviteMember201JSONResponse{Invite: inviteDTO(inv), AcceptUrl: acceptURL}, nil
}

func (h *handler) ResendInvite(ctx context.Context, request api.ResendInviteRequestObject) (api.ResendInviteResponseObject, error) {
	inv, acceptURL, err := h.members.ResendInvite(ctx, request.WorkspaceId, request.UserId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.ResendInvite403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can resend invitations.", "")), nil
		case errors.Is(err, entity.ErrInviteNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.ResendInvite404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No pending invitation for that member.", "")), nil
		default:
			return nil, err
		}
	}
	return api.ResendInvite200JSONResponse{Invite: inviteDTO(inv), AcceptUrl: acceptURL}, nil
}

func (h *handler) RevokeInvite(ctx context.Context, request api.RevokeInviteRequestObject) (api.RevokeInviteResponseObject, error) {
	err := h.members.RevokeInvite(ctx, request.WorkspaceId, request.UserId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.RevokeInvite403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can revoke invitations.", "")), nil
		case errors.Is(err, entity.ErrInviteNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.RevokeInvite404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No pending invitation for that member.", "")), nil
		default:
			return nil, err
		}
	}
	return api.RevokeInvite204Response{}, nil
}

func (h *handler) ChangeMemberRole(ctx context.Context, request api.ChangeMemberRoleRequestObject) (api.ChangeMemberRoleResponseObject, error) {
	if request.Body == nil {
		return api.ChangeMemberRole400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	err := h.members.ChangeRole(ctx, request.WorkspaceId, request.UserId, entity.Role(request.Body.Role))
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.ChangeMemberRole403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can change roles.", "")), nil
		case errors.Is(err, entity.ErrMemberLastAdmin):
			return api.ChangeMemberRole409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Last admin", "You can't remove the last admin. Promote another member to admin first.", "last-admin")), nil
		case errors.Is(err, entity.ErrMemberNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.ChangeMemberRole404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such member.", "")), nil
		case errors.Is(err, entity.ErrRoleInvalid), errors.Is(err, entity.ErrMemberDeactivated):
			return api.ChangeMemberRole400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid role change", "That role change isn't allowed.", "")), nil
		default:
			return nil, err
		}
	}
	return h.memberResponse(ctx, request.WorkspaceId, request.UserId)
}

func (h *handler) DeactivateMember(ctx context.Context, request api.DeactivateMemberRequestObject) (api.DeactivateMemberResponseObject, error) {
	if request.Body == nil {
		return api.DeactivateMember400ApplicationProblemPlusJSONResponse(prob(http.StatusBadRequest, "Invalid request", "The request body was empty.", "")), nil
	}
	replacements := map[string]string{}
	if request.Body.Replacements != nil {
		replacements = request.Body.Replacements
	}
	err := h.members.Deactivate(ctx, request.WorkspaceId, request.UserId, replacements)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.DeactivateMember403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can deactivate members.", "")), nil
		case errors.Is(err, entity.ErrMemberLastAdmin):
			return api.DeactivateMember409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Last admin", "You can't deactivate the last admin. Promote another member to admin first.", "last-admin")), nil
		case errors.Is(err, entity.ErrMemberReplacementsIncomplete), errors.Is(err, entity.ErrMemberReplacementInvalid):
			return api.DeactivateMember409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Reassignment required", "This member still owns schedules or policies. Reassign every one to an active member before deactivating.", "unresolved-references")), nil
		case errors.Is(err, entity.ErrMemberNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember), errors.Is(err, entity.ErrMemberNotDeactivated):
			return api.DeactivateMember404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No active member to deactivate.", "")), nil
		default:
			return nil, err
		}
	}
	m, err := h.members.Get(ctx, request.WorkspaceId, request.UserId)
	if err != nil {
		return nil, err
	}
	return api.DeactivateMember200JSONResponse(memberDTO(m)), nil
}

func (h *handler) ReactivateMember(ctx context.Context, request api.ReactivateMemberRequestObject) (api.ReactivateMemberResponseObject, error) {
	err := h.members.Reactivate(ctx, request.WorkspaceId, request.UserId)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrForbidden):
			return api.ReactivateMember403ApplicationProblemPlusJSONResponse(prob(http.StatusForbidden, "Forbidden", "Only admins can reactivate members.", "")), nil
		case errors.Is(err, entity.ErrMemberNotDeactivated):
			return api.ReactivateMember409ApplicationProblemPlusJSONResponse(prob(http.StatusConflict, "Not deactivated", "That member is not deactivated.", "")), nil
		case errors.Is(err, entity.ErrMemberNotFound), errors.Is(err, entity.ErrWorkspaceNotFound), errors.Is(err, entity.ErrNotMember):
			return api.ReactivateMember404ApplicationProblemPlusJSONResponse(prob(http.StatusNotFound, "Not found", "No such member.", "")), nil
		default:
			return nil, err
		}
	}
	return h.reactivateResponse(ctx, request.WorkspaceId, request.UserId)
}

func (h *handler) memberResponse(ctx context.Context, workspaceID, userID string) (api.ChangeMemberRoleResponseObject, error) {
	m, err := h.members.Get(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	return api.ChangeMemberRole200JSONResponse(memberDTO(m)), nil
}

func (h *handler) reactivateResponse(ctx context.Context, workspaceID, userID string) (api.ReactivateMemberResponseObject, error) {
	m, err := h.members.Get(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	return api.ReactivateMember200JSONResponse(memberDTO(m)), nil
}

func memberDTO(m entity.Member) api.Member {
	dto := api.Member{
		UserId:      m.UserID,
		Name:        m.Name,
		Email:       m.Email,
		Role:        api.Role(m.Role),
		Status:      api.MemberStatus(m.Status),
		AuthMethod:  api.MemberAuthMethod(m.Auth()),
		TwoFactor:   m.TOTPEnabled,
		Deactivated: m.Status == entity.MemberStatusDeactivated,
	}
	if !m.LastActiveAt.IsZero() {
		dto.LastActiveAt = &m.LastActiveAt
	}
	if len(m.References) > 0 {
		refs := make([]api.MemberReference, 0, len(m.References))
		for _, r := range m.References {
			refs = append(refs, api.MemberReference{Id: r.ID, Kind: r.Kind, Icon: ptr(r.Icon), Label: r.Label, Detail: r.Detail})
		}
		dto.References = &refs
	}
	return dto
}

func inviteDTO(inv entity.Invite) api.Invite {
	return api.Invite{
		UserId:    inv.UserID,
		Email:     inv.Email,
		Role:      api.Role(inv.Role),
		InvitedBy: inv.InvitedByName,
		SentAt:    inv.CreatedAt,
		ExpiresAt: inv.ExpiresAt,
	}
}
