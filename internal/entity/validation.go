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

	errSourceSlug         = validation.NewError("source_slug_invalid", "A source URL uses lowercase letters, numbers, and hyphens, and starts with a letter.")
	errSourceSlugReserved = validation.NewError("source_slug_reserved", "That name is reserved. Pick another.")
	errSourceName         = validation.NewError("source_name_invalid", "Enter a source name of 60 characters or fewer.")
	errSourceFormat       = validation.NewError("source_format_invalid", "Choose a supported format: Alertmanager, Grafana, Uptime Kuma, heartbeat, or generic JSON.")
	errSourceAutoResolve  = validation.NewError("source_auto_resolve_invalid", "Auto-resolve must be between 0 and 30 days.")
	errSourceMapping      = validation.NewError("source_mapping_invalid", "Each mapping needs a known field and a path, listed once.")
	errSourceMappingTitle = validation.NewError("source_mapping_title", "Map the title field so alerts have something to show.")
	errAlertSeverity      = validation.NewError("alert_severity_invalid", "Choose critical, high, or warning.")
	errAlertStatus        = validation.NewError("alert_status_invalid", "Choose open, acked, or resolved.")
	errPolicyRef          = validation.NewError("policy_ref_invalid", "Pick an escalation policy.")
	errRouteCondition     = validation.NewError("route_condition_invalid", "Each condition needs a known field, an operator, and a value.")
	errRouteConditions    = validation.NewError("route_conditions_empty", "A rule needs at least one condition with a value.")
	errConditionOp        = validation.NewError("condition_op_invalid", "Choose is, is not, contains, or matches.")
	errSilenceCondition   = validation.NewError("silence_condition_invalid", "Each scope needs a source, service, or label value.")
	errSilenceReason      = validation.NewError("silence_reason_invalid", "Keep the reason to 200 characters or fewer.")
	errGroupRuleFields    = validation.NewError("group_rule_invalid", "A grouping rule needs between 1 and 5 known fields.")
	errGroupRuleWindow    = validation.NewError("group_window_invalid", "A grouping window runs from 1 minute to 24 hours.")
	errMonitorInterval    = validation.NewError("monitor_interval_invalid", "A check-in interval runs from 1 minute to 30 days.")
	errMonitorGrace       = validation.NewError("monitor_grace_invalid", "A grace period runs from 0 up to 24 hours.")
	errMonitorState       = validation.NewError("monitor_state_invalid", "Choose a valid monitor state.")
	errEscalationSlug     = validation.NewError("escalation_slug_invalid", "A policy URL uses lowercase letters, numbers, and hyphens, and starts with a letter.")
	errEscalationName     = validation.NewError("escalation_name_invalid", "Enter a policy name of 60 characters or fewer.")
	errEscalationRepeat   = validation.NewError("escalation_repeat_invalid", "A policy repeats between 0 and 3 times.")
	errEscalationAckTTL   = validation.NewError("escalation_ack_timeout_invalid", "An acknowledgement expiry runs up to 24 hours.")
	errEscalationEmpty    = validation.NewError("escalation_nodes_empty", "A policy needs at least one level.")
	errEscalationLevel    = validation.NewError("escalation_level_invalid", "Every level needs between 1 and 20 valid targets.")
	errEscalationWait     = validation.NewError("escalation_wait_invalid", "A level waits between 1 minute and 1 hour for an acknowledgement.")
	errEscalationMode     = validation.NewError("escalation_mode_invalid", "Choose all at once or round-robin.")
	errEscalationTarget   = validation.NewError("escalation_target_invalid", "Every target needs a valid type and a reference.")
	errEscalationBranch   = validation.NewError("escalation_branch_invalid", "A branch splits by priority or working hours into its two lanes.")
	errEscalationDeadEnd  = validation.NewError("escalation_lane_empty", "Every branch lane needs at least one level.")
	errEscalationHours    = validation.NewError("escalation_hours_invalid", "A working-hours window needs at least one weekday, a valid time range, and a real timezone.")
	errEscalationDepth    = validation.NewError("escalation_depth_invalid", "The policy tree is nested too deeply.")
	errEscalationHookName = validation.NewError("escalation_webhook_name_invalid", "Enter a webhook name of 60 characters or fewer.")
	errEscalationHookURL  = validation.NewError("escalation_webhook_url_invalid", "Enter a valid https URL for the webhook.")
	errEscalationGone     = validation.NewError("escalation_target_unknown", "A target no longer exists. Remove it and pick another.")
	errEscalationDark     = validation.NewError("escalation_level_unreachable", "A level only pages people who can't be paged. Fix its targets.")
	errEscalationNoReach  = validation.NewError("escalation_no_reach", "This policy can never notify anyone. Add at least one reachable target.")

	errNotificationSteps      = validation.NewError("notification_steps_invalid", "A rule takes up to 12 steps, each delayed by 0 to 60 minutes.")
	errNotificationFirstStep  = validation.NewError("notification_first_step", "The first step always fires immediately.")
	errNotificationChannel    = validation.NewError("notification_channel_invalid", "Pick a channel you've connected.")
	errNotificationQuietHours = validation.NewError("notification_quiet_hours_invalid", "Quiet hours need at least one day, a start and end that differ, and a real timezone.")
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

func sourceSlugField(value any) error {
	s, _ := value.(string)
	if !ValidSlugFormat(s) || len(s) > SourceSlugMaxLength {
		return errSourceSlug
	}
	if slices.Contains(SourceReservedSlugs, s) {
		return errSourceSlugReserved
	}
	return nil
}

func sourceNameField(value any) error {
	s, _ := value.(string)
	name := strings.TrimSpace(s)
	if name == "" || len(name) > SourceNameMaxLength {
		return errSourceName
	}
	return nil
}

func sourceFormatField(value any) error {
	f, _ := value.(SourceFormat)
	switch f {
	case SourceFormatAlertmanager, SourceFormatGrafana, SourceFormatKuma, SourceFormatHeartbeat, SourceFormatGeneric:
		return nil
	default:
		return errSourceFormat
	}
}

func alertSeverityField(value any) error {
	s, _ := value.(AlertSeverity)
	switch s {
	case SeverityCritical, SeverityHigh, SeverityWarning:
		return nil
	default:
		return errAlertSeverity
	}
}

func alertStatusField(value any) error {
	s, _ := value.(AlertStatus)
	switch s {
	case AlertStatusOpen, AlertStatusAcked, AlertStatusResolved:
		return nil
	default:
		return errAlertStatus
	}
}

func autoResolveField(value any) error {
	d, _ := value.(time.Duration)
	if d < 0 || d > SourceMaxAutoResolve {
		return errSourceAutoResolve
	}
	return nil
}

func policyRefField(value any) error {
	s, _ := value.(string)
	ref := strings.TrimSpace(s)
	if ref == "" || len(ref) > PolicyRefMaxLength || !ValidSlugFormat(ref) {
		return errPolicyRef
	}
	return nil
}

func sourceMappingsField(value any) error {
	mappings, _ := value.([]SourceMapping)
	if len(mappings) > SourceMaxMappings {
		return errSourceMapping
	}
	seen := make(map[string]struct{}, len(mappings))
	for _, m := range mappings {
		if !slices.Contains(MappingFields, m.Field) {
			return errSourceMapping
		}
		if _, dup := seen[m.Field]; dup {
			return errSourceMapping
		}
		seen[m.Field] = struct{}{}
		path := strings.TrimSpace(m.Path)
		if path == "" || len(path) > SourceMappingPathMax {
			return errSourceMapping
		}
	}
	if _, ok := seen[MappingFieldTitle]; !ok {
		return errSourceMappingTitle
	}
	return nil
}

func conditionOpField(value any) error {
	op, _ := value.(ConditionOp)
	switch op {
	case ConditionIs, ConditionIsNot, ConditionContains, ConditionMatches:
		return nil
	default:
		return errConditionOp
	}
}

func routeConditionsField(value any) error {
	conditions, _ := value.([]RouteCondition)
	if len(conditions) == 0 || len(conditions) > RouteMaxConditions {
		return errRouteConditions
	}
	for _, c := range conditions {
		if err := conditionOpField(c.Op); err != nil {
			return err
		}
		if !routeFieldKnown(c.Field) {
			return errRouteCondition
		}
		v := strings.TrimSpace(c.Value)
		if v == "" || len(v) > RouteValueMaxLength {
			return errRouteCondition
		}
	}
	return nil
}

func routeFieldKnown(field string) bool {
	if strings.HasPrefix(field, "labels.") {
		return len(field) > len("labels.")
	}
	return slices.Contains(RouteFields, field)
}

func silenceConditionsField(value any) error {
	conditions, _ := value.([]SilenceCondition)
	if len(conditions) == 0 || len(conditions) > SilenceMaxConditions {
		return errSilenceCondition
	}
	for _, c := range conditions {
		if !slices.Contains(SilenceScopeFields, c.Field) {
			return errSilenceCondition
		}
		if strings.TrimSpace(c.Value) == "" {
			return errSilenceCondition
		}
	}
	return nil
}

func silenceReasonField(value any) error {
	s, _ := value.(string)
	if len(s) > SilenceReasonMax {
		return errSilenceReason
	}
	return nil
}

func groupRuleFieldsField(value any) error {
	fields, _ := value.([]string)
	if len(fields) == 0 || len(fields) > GroupRuleMaxFields {
		return errGroupRuleFields
	}
	seen := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		if !routeFieldKnown(f) {
			return errGroupRuleFields
		}
		if _, dup := seen[f]; dup {
			return errGroupRuleFields
		}
		seen[f] = struct{}{}
	}
	return nil
}

func groupRuleWindowField(value any) error {
	d, _ := value.(time.Duration)
	if d < GroupWindowMin || d > GroupWindowMax {
		return errGroupRuleWindow
	}
	return nil
}

func monitorIntervalField(value any) error {
	d, _ := value.(time.Duration)
	if d < MonitorIntervalMin || d > MonitorIntervalMax {
		return errMonitorInterval
	}
	return nil
}

func monitorGraceField(value any) error {
	d, _ := value.(time.Duration)
	if d < 0 || d > MonitorGraceMax {
		return errMonitorGrace
	}
	return nil
}

func monitorStateField(value any) error {
	s, _ := value.(MonitorState)
	switch s {
	case MonitorStateHealthy, MonitorStateMissed, MonitorStatePaused:
		return nil
	default:
		return errMonitorState
	}
}

func escalationSlugField(value any) error {
	s, _ := value.(string)
	if !ValidSlugFormat(s) || len(s) > EscalationSlugMaxLength {
		return errEscalationSlug
	}
	if slices.Contains(EscalationReservedSlugs, s) {
		return errEscalationSlug
	}
	return nil
}

func escalationNameField(value any) error {
	s, _ := value.(string)
	name := strings.TrimSpace(s)
	if name == "" || len(name) > EscalationNameMaxLength {
		return errEscalationName
	}
	return nil
}

func escalationRepeatField(value any) error {
	n, _ := value.(int)
	if n < 0 || n > EscalationRepeatMax {
		return errEscalationRepeat
	}
	return nil
}

func escalationAckTimeoutField(value any) error {
	d, _ := value.(time.Duration)
	if d < 0 || d > EscalationAckTimeoutMax {
		return errEscalationAckTTL
	}
	return nil
}

func escalationTargetTypeField(value any) error {
	t, _ := value.(EscalationTargetType)
	switch t {
	case EscalationTargetPerson, EscalationTargetSchedule, EscalationTargetTeam, EscalationTargetWebhook:
		return nil
	default:
		return errEscalationTarget
	}
}

func notifyModeField(value any) error {
	m, _ := value.(NotifyMode)
	switch m {
	case NotifyModeAll, NotifyModeRoundRobin:
		return nil
	default:
		return errEscalationMode
	}
}

func escalationBranchKindField(value any) error {
	k, _ := value.(EscalationBranchKind)
	switch k {
	case EscalationBranchPriority, EscalationBranchHours:
		return nil
	default:
		return errEscalationBranch
	}
}

func escalationHoursField(h HoursWindow) error {
	if len(h.Days) == 0 {
		return errEscalationHours
	}
	for _, d := range h.Days {
		if d < 0 || d > EscalationHoursDayMax {
			return errEscalationHours
		}
	}
	if h.StartMinute < 0 || h.StartMinute >= EscalationMinutesPerDay ||
		h.EndMinute < 0 || h.EndMinute >= EscalationMinutesPerDay ||
		h.StartMinute == h.EndMinute {
		return errEscalationHours
	}
	if _, err := time.LoadLocation(h.Timezone); err != nil || h.Timezone == "" {
		return errEscalationHours
	}
	return nil
}

func notificationStepsField(value any) error {
	steps, _ := value.([]NotificationStep)
	if len(steps) > NotificationMaxSteps {
		return errNotificationSteps
	}
	for i, s := range steps {
		if !s.Channel.Valid() {
			return errNotificationChannel
		}
		if s.Delay < 0 || s.Delay > NotificationStepDelayMax {
			return errNotificationSteps
		}
		if i == 0 && s.Delay != 0 {
			return errNotificationFirstStep
		}
	}
	return nil
}

func notificationQuietHoursField(value any) error {
	q, _ := value.(QuietHours)
	if !q.Enabled {
		return nil
	}
	if err := escalationHoursField(q.Window); err != nil {
		return errNotificationQuietHours
	}
	return nil
}

func escalationLevelField(l EscalationLevel) error {
	if len(l.Targets) == 0 || len(l.Targets) > EscalationMaxTargets {
		return errEscalationLevel
	}
	for _, t := range l.Targets {
		if err := escalationTargetTypeField(t.Type); err != nil {
			return err
		}
		if strings.TrimSpace(t.Ref) == "" {
			return errEscalationTarget
		}
	}
	if err := notifyModeField(l.Mode); err != nil {
		return err
	}
	if l.Wait < EscalationWaitMin || l.Wait > EscalationWaitMax {
		return errEscalationWait
	}
	return nil
}

func escalationLaneKeys(kind EscalationBranchKind) []string {
	if kind == EscalationBranchHours {
		return []string{EscalationLaneBusiness, EscalationLaneOff}
	}
	return []string{EscalationLaneHigh, EscalationLaneLow}
}

func escalationBranchField(b EscalationBranch, depth int) error {
	if err := escalationBranchKindField(b.On); err != nil {
		return err
	}
	if b.On == EscalationBranchHours {
		if err := escalationHoursField(b.Hours); err != nil {
			return err
		}
	}
	keys := escalationLaneKeys(b.On)
	if len(b.Lanes) != len(keys) {
		return errEscalationBranch
	}
	for i, lane := range b.Lanes {
		if lane.Key != keys[i] {
			return errEscalationBranch
		}
		if len(lane.Nodes) == 0 {
			return errEscalationDeadEnd
		}
		if err := escalationWalk(lane.Nodes, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func escalationWalk(nodes []EscalationNode, depth int) error {
	if depth > EscalationMaxDepth {
		return errEscalationDepth
	}
	for _, node := range nodes {
		switch {
		case node.Level != nil:
			if err := escalationLevelField(*node.Level); err != nil {
				return err
			}
		case node.Branch != nil:
			if err := escalationBranchField(*node.Branch, depth); err != nil {
				return err
			}
		default:
			return errEscalationBranch
		}
	}
	return nil
}

func escalationNodesField(value any) error {
	nodes, _ := value.([]EscalationNode)
	if len(nodes) == 0 {
		return errEscalationEmpty
	}
	if err := escalationWalk(nodes, 1); err != nil {
		return err
	}
	if len(collectTargets(nodes)) == 0 {
		return errEscalationEmpty
	}
	return nil
}

func escalationWebhookNameField(value any) error {
	s, _ := value.(string)
	name := strings.TrimSpace(s)
	if name == "" || len(name) > EscalationWebhookNameMax {
		return errEscalationHookName
	}
	return nil
}

func escalationWebhookURLField(value any) error {
	s, _ := value.(string)
	raw := strings.TrimSpace(s)
	if raw == "" || len(raw) > EscalationWebhookURLMax {
		return errEscalationHookURL
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return errEscalationHookURL
	}
	return nil
}
