package entity

import "testing"

func TestIncidentStatusCanTransition(t *testing.T) {
	legal := map[IncidentStatus]map[IncidentStatus]bool{
		IncidentStatusDeclared: {
			IncidentStatusInvestigating: true,
		},
		IncidentStatusInvestigating: {
			IncidentStatusIdentified: true,
		},
		IncidentStatusIdentified: {
			IncidentStatusInvestigating: true,
			IncidentStatusMonitoring:    true,
		},
		IncidentStatusMonitoring: {
			IncidentStatusIdentified: true,
			IncidentStatusResolved:   true,
		},
		IncidentStatusResolved: {},
	}

	all := append([]IncidentStatus{}, IncidentStatusOrder...)
	for _, from := range all {
		for _, to := range all {
			want := legal[from][to]
			if got := from.CanTransition(to); got != want {
				t.Errorf("CanTransition(%s -> %s) = %v, want %v", from, to, got, want)
			}
		}
	}
}

func TestIncidentStatusCanTransitionRejectsUnknown(t *testing.T) {
	if IncidentStatus("bogus").CanTransition(IncidentStatusInvestigating) {
		t.Error("unknown source status must not transition")
	}
	if IncidentStatusDeclared.CanTransition("bogus") {
		t.Error("unknown target status must not transition")
	}
	if IncidentStatusDeclared.CanTransition(IncidentStatusDeclared) {
		t.Error("identity transition must be rejected")
	}
}

func TestIncidentStatusNoBackToDeclared(t *testing.T) {
	if IncidentStatusInvestigating.CanTransition(IncidentStatusDeclared) {
		t.Error("must never step back into declared")
	}
}

func TestIncidentStatusResolvedIsTerminal(t *testing.T) {
	for _, to := range IncidentStatusOrder {
		if IncidentStatusResolved.CanTransition(to) {
			t.Errorf("resolved must be terminal, allowed -> %s", to)
		}
	}
}

func TestIncidentStatusNoStageSkips(t *testing.T) {
	if IncidentStatusDeclared.CanTransition(IncidentStatusIdentified) {
		t.Error("must not skip investigating")
	}
	if IncidentStatusInvestigating.CanTransition(IncidentStatusResolved) {
		t.Error("must not resolve without monitoring")
	}
	if IncidentStatusDeclared.CanTransition(IncidentStatusResolved) {
		t.Error("must not resolve straight from declared")
	}
}

func TestIncidentStatusActive(t *testing.T) {
	for _, s := range IncidentStatusOrder {
		want := s != IncidentStatusResolved
		if got := s.Active(); got != want {
			t.Errorf("Active(%s) = %v, want %v", s, got, want)
		}
	}
	if IncidentStatus("bogus").Active() {
		t.Error("unknown status is not active")
	}
}

func TestIncidentSeverityForAlert(t *testing.T) {
	cases := map[AlertSeverity]string{
		SeverityCritical: "SEV1",
		SeverityHigh:     "SEV2",
		SeverityWarning:  "SEV3",
	}
	for sev, want := range cases {
		if got := IncidentSeverityForAlert(sev); got != want {
			t.Errorf("IncidentSeverityForAlert(%s) = %s, want %s", sev, got, want)
		}
	}
}
