package entity

import (
	"errors"
	"net/mail"
	"net/url"
	"slices"
	"sort"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := errors.AsType[validation.Errors](err); ok {
		return true
	}
	_, ok := errors.AsType[validation.Error](err)
	return ok
}

func ValidationMessage(err error) string {
	if verrs, ok := errors.AsType[validation.Errors](err); ok {
		keys := make([]string, 0, len(verrs))
		for k := range verrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		msgs := make([]string, 0, len(keys))
		for _, k := range keys {
			msgs = append(msgs, verrs[k].Error())
		}
		if len(msgs) > 0 {
			return strings.Join(msgs, " ")
		}
	}
	if verr, ok := errors.AsType[validation.Error](err); ok {
		return verr.Error()
	}
	return "One or more fields are invalid. Check your input and try again."
}

var (
	errName          = validation.NewError("name_invalid", "Enter a name of 80 characters or fewer.")
	errEmail         = validation.NewError("email_invalid", "That email address isn't valid. Enter a valid address like name@example.com.")
	errPassword      = validation.NewError("password_weak", "Password must be at least 12 characters. Choose a longer password and try again.")
	errTimezone      = validation.NewError("timezone_invalid", "That timezone isn't recognised. Pick an IANA timezone such as Europe/Berlin.")
	errWorkspaceName = validation.NewError("workspace_name_invalid", "Enter a workspace name of 80 characters or fewer.")
	errSlugReserved  = validation.NewError("slug_reserved", "That workspace URL is reserved. Choose a different one.")
	errSlugFormat    = validation.NewError("slug_invalid", "A workspace URL uses lowercase letters, numbers, and hyphens, and starts with a letter.")
	errTeamName      = validation.NewError("team_name_invalid", "Enter a team name of 60 characters or fewer.")
	errTeamMembers   = validation.NewError("team_members_max", "A team can have at most 50 members.")
	errChannelDetail = validation.NewError("channel_invalid", "Enter a valid destination: a real email address or an https URL.")
	errKeyName       = validation.NewError("key_name_invalid", "Enter a key name of 60 characters or fewer.")
	errKeyKind       = validation.NewError("key_kind_invalid", "Choose a valid key type.")
	errKeyScope      = validation.NewError("key_scope_invalid", "Choose at least one valid scope.")
	errRole          = validation.NewError("role_invalid", "Choose a valid role.")
	errSSOMode       = validation.NewError("sso_mode_invalid", "Choose OIDC or SAML.")
	errSSOURL        = validation.NewError("sso_url_invalid", "Enter a valid https URL.")
	errSSOClientID   = validation.NewError("sso_client_id", "Enter the client ID.")
	errSSODomain     = validation.NewError("sso_domain_invalid", "Enter valid email domains without @ or spaces.")

	errScheduleSlug         = validation.NewError("schedule_slug_invalid", "A schedule URL uses lowercase letters, numbers, and hyphens, and starts with a letter.")
	errScheduleSlugReserved = validation.NewError("schedule_slug_reserved", "That name is reserved. Pick another.")
	errScheduleTeam         = validation.NewError("schedule_team_invalid", "Pick a team for this schedule.")
	errScheduleRotation     = validation.NewError("schedule_rotation_invalid", "Choose a daily, weekly, or custom rotation.")
	errScheduleLayers       = validation.NewError("schedule_layers_invalid", "A schedule needs between 1 and 10 layers, each with at least one participant.")
	errScheduleLayer        = validation.NewError("schedule_layer_invalid", "Check each layer: a rotation, a handover hour, a start date, and restriction hours between 00:00 and 24:00.")
	errOverrideUser         = validation.NewError("override_user_invalid", "Choose who takes the override.")
	errOverrideReason       = validation.NewError("override_reason_invalid", "Keep the reason to 200 characters or fewer.")
)

func nameField(value any) error {
	s, _ := value.(string)
	name := strings.TrimSpace(s)
	if name == "" || len(name) > NameMaxLength {
		return errName
	}
	return nil
}

func emailField(value any) error {
	s, _ := value.(string)
	email := strings.TrimSpace(s)
	if email == "" || len(email) > EmailMaxLength {
		return errEmail
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Name != "" || addr.Address != email {
		return errEmail
	}
	return nil
}

func passwordField(value any) error {
	s, _ := value.(string)
	if len(s) < PasswordMinLength || len(s) > PasswordMaxLength {
		return errPassword
	}
	return nil
}

func timezoneField(value any) error {
	s, _ := value.(string)
	if strings.TrimSpace(s) == "" {
		return errTimezone
	}
	if _, err := time.LoadLocation(s); err != nil {
		return errTimezone
	}
	return nil
}

func workspaceNameField(value any) error {
	s, _ := value.(string)
	name := strings.TrimSpace(s)
	if name == "" || len(name) > WorkspaceNameMaxLength {
		return errWorkspaceName
	}
	return nil
}

func slugField(value any) error {
	s, _ := value.(string)
	if !ValidSlugFormat(s) {
		return errSlugFormat
	}
	if slices.Contains(WorkspaceReservedSlugs, s) {
		return errSlugReserved
	}
	return nil
}

func teamNameField(value any) error {
	s, _ := value.(string)
	name := strings.TrimSpace(s)
	if name == "" || len(name) > TeamNameMaxLength {
		return errTeamName
	}
	return nil
}

func teamMembersField(value any) error {
	ids, _ := value.([]string)
	if len(ids) > TeamMaxMembers {
		return errTeamMembers
	}
	return nil
}

func channelDetailFor(t ChannelType, value any) error {
	s, _ := value.(string)
	detail := strings.TrimSpace(s)
	if detail == "" || len(detail) > ChannelDetailMaxLength {
		return errChannelDetail
	}
	switch t {
	case ChannelTypeEmail:
		if emailField(detail) != nil {
			return errChannelDetail
		}
	case ChannelTypeWebhook, ChannelTypeNtfy:
		u, err := url.ParseRequestURI(detail)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
			return errChannelDetail
		}
	}
	return nil
}

func keyNameField(value any) error {
	s, _ := value.(string)
	name := strings.TrimSpace(s)
	if name == "" || len(name) > APIKeyNameMaxLength {
		return errKeyName
	}
	return nil
}

func keyKindField(value any) error {
	k, _ := value.(KeyKind)
	if k != KeyKindPersonal && k != KeyKindWorkspace {
		return errKeyKind
	}
	return nil
}

func keyScopesField(value any) error {
	scopes, _ := value.([]Scope)
	if len(scopes) == 0 {
		return errKeyScope
	}
	for _, s := range scopes {
		if !ScopeValid(s) {
			return errKeyScope
		}
	}
	return nil
}

func channelTypeField(value any) error {
	t, _ := value.(ChannelType)
	if !t.Valid() {
		return errChannelDetail
	}
	return nil
}

func roleField(value any) error {
	r, _ := value.(Role)
	if r != RoleAdmin && r != RoleMember {
		return errRole
	}
	return nil
}

func ssoModeField(value any) error {
	m, _ := value.(SSOMode)
	if m != SSOModeOIDC && m != SSOModeSAML {
		return errSSOMode
	}
	return nil
}

func httpsURLField(value any) error {
	s, _ := value.(string)
	if !validHTTPSURL(s) {
		return errSSOURL
	}
	return nil
}

func clientIDField(value any) error {
	s, _ := value.(string)
	if strings.TrimSpace(s) == "" {
		return errSSOClientID
	}
	return nil
}

func domainField(value any) error {
	d, _ := value.(string)
	if strings.TrimSpace(d) == "" || strings.ContainsAny(d, "@ ") {
		return errSSODomain
	}
	return nil
}

func scheduleSlugField(value any) error {
	s, _ := value.(string)
	if !ValidSlugFormat(s) || len(s) > ScheduleSlugMaxLength {
		return errScheduleSlug
	}
	if slices.Contains(ScheduleReservedSlugs, s) {
		return errScheduleSlugReserved
	}
	return nil
}

func scheduleTeamField(value any) error {
	s, _ := value.(string)
	if strings.TrimSpace(s) == "" {
		return errScheduleTeam
	}
	return nil
}

func rotationField(value any) error {
	r, _ := value.(Rotation)
	if r != RotationDaily && r != RotationWeekly && r != RotationCustom {
		return errScheduleRotation
	}
	return nil
}

func scheduleLayersField(value any) error {
	layers, _ := value.([]NewScheduleLayer)
	if len(layers) < 1 || len(layers) > ScheduleMaxLayers {
		return errScheduleLayers
	}
	for _, layer := range layers {
		if err := scheduleLayer(layer); err != nil {
			return err
		}
	}
	return nil
}

func scheduleLayer(layer NewScheduleLayer) error {
	if n := len(layer.Participants); n < 1 || n > LayerMaxParticipants {
		return errScheduleLayers
	}
	for _, p := range layer.Participants {
		if strings.TrimSpace(p) == "" {
			return errScheduleLayers
		}
	}
	if rotationField(layer.Rotation) != nil {
		return errScheduleLayer
	}
	if layer.Rotation == RotationCustom &&
		(layer.IntervalDays < LayerMinIntervalDays || layer.IntervalDays > LayerMaxIntervalDays) {
		return errScheduleLayer
	}
	if layer.HandoverHour < 0 || layer.HandoverHour > 23 {
		return errScheduleLayer
	}
	if layer.StartsOn.IsZero() {
		return errScheduleLayer
	}
	if len(layer.Restrictions) > LayerMaxRestrictions {
		return errScheduleLayer
	}
	for _, r := range layer.Restrictions {
		if r.Weekday < 0 || r.Weekday > 6 {
			return errScheduleLayer
		}
		if r.StartMinute < 0 || r.EndMinute > MinutesPerDay || r.StartMinute >= r.EndMinute {
			return errScheduleLayer
		}
	}
	return nil
}

func overrideUserField(value any) error {
	s, _ := value.(string)
	if strings.TrimSpace(s) == "" {
		return errOverrideUser
	}
	return nil
}

func overrideReasonField(value any) error {
	s, _ := value.(string)
	if len(s) > ScheduleReasonMaxLength {
		return errOverrideReason
	}
	return nil
}
