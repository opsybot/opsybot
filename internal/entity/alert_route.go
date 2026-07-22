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
	PolicyID    string
	PolicySlug  string
	Conditions  []RouteCondition
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type NewAlertRoute struct {
	PolicySlug string
	PolicyID   string
	Conditions []RouteCondition
}

type GroupRule struct {
	ID          string
	WorkspaceID string
	Fields      []string
	Window      time.Duration
	Position    int
}

func (g GroupRule) Validate() error {
	return validation.ValidateStruct(&g,
		validation.Field(&g.Fields, validation.By(groupRuleFieldsField)),
		validation.Field(&g.Window, validation.By(groupRuleWindowField)),
	)
}

func ValidateGroupRules(rules []GroupRule) error {
	if len(rules) > GroupRuleMaxRules {
		return ErrGroupRuleLimit
	}
	for _, rule := range rules {
		if err := rule.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type AlertSettings struct {
	WorkspaceID       string
	DefaultPolicyID   string
	DefaultPolicySlug string
}

type RoutePreview struct {
	MatchedRouteID string
	Position       int
	PolicySlug     string
	GroupFields    []string
}

var (
	ErrAlertRouteNotFound   = errors.New("alert route not found")
	ErrAlertRouteConditions = errors.New("alert route needs at least one condition")
	ErrGroupRuleNotFound    = errors.New("group rule not found")
	ErrGroupRuleLimit       = errors.New("group rule limit reached")
)

func ValidatePolicyRef(ref string) error {
	return policyRefField(ref)
}

func (o ConditionOp) Validate() error {
	return conditionOpField(o)
}

func (n NewAlertRoute) Validate() error {
	return validation.ValidateStruct(&n,
		validation.Field(&n.PolicySlug, validation.By(policyRefField)),
		validation.Field(&n.Conditions, validation.By(routeConditionsField)),
	)
}
