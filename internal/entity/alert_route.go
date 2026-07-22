package entity

import (
	"errors"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ConditionOp string

const (
	ConditionIs       ConditionOp = "is"
	ConditionIsNot    ConditionOp = "is not"
	ConditionContains ConditionOp = "contains"
	ConditionMatches  ConditionOp = "matches"
)

const (
	RouteMaxConditions   = 10
	RouteMaxRules        = 100
	RouteValueMaxLength  = 200
	GroupRuleMaxFields   = 5
	GroupRuleMaxRules    = 20
	GroupWindowMin       = time.Minute
	GroupWindowMax       = 24 * time.Hour
	GroupWindowDefault   = 5 * time.Minute
	SilenceMaxConditions = 10
	SilenceReasonMax     = 200
)

const DefaultPolicyRef = "platform-default"

var RouteFields = []string{"source", "service", "severity", "title", "labels"}

type RouteCondition struct {
	Field string
	Op    ConditionOp
	Value string
}

type AlertRoute struct {
	ID          string
	WorkspaceID string
	Position    int
	PolicyRef   string
	Conditions  []RouteCondition
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NewAlertRoute struct {
	PolicyRef  string
	Conditions []RouteCondition
}

type GroupRule struct {
	ID          string
	WorkspaceID string
	Fields      []string
	Window      time.Duration
	Position    int
}

type AlertSettings struct {
	WorkspaceID      string
	DefaultPolicyRef string
}

var (
	ErrAlertRouteNotFound   = errors.New("alert route not found")
	ErrAlertRouteConditions = errors.New("alert route needs at least one condition")
	ErrGroupRuleNotFound    = errors.New("group rule not found")
)

func ValidatePolicyRef(ref string) error {
	return policyRefField(ref)
}

func (o ConditionOp) Validate() error {
	return conditionOpField(o)
}

func (n NewAlertRoute) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.PolicyRef, validation.By(policyRefField)),
		validation.Field(&n.Conditions, validation.By(routeConditionsField)),
	)
}
