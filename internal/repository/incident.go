package repository

//go:generate go tool mockgen -source=incident.go -destination=./incident/incident_mock.go -package=incident

import (
	"context"
	"time"

	"github.com/opsybot/opsybot/internal/entity"
)

type Incident interface {
	NextNumber(ctx context.Context, workspaceID string) (int, error)
	Create(ctx context.Context, in entity.Incident) (entity.Incident, error)
	GetByID(ctx context.Context, workspaceID, id string) (entity.Incident, error)
	List(ctx context.Context, workspaceID string, filter entity.IncidentFilter) (entity.IncidentPage, error)
	Patch(ctx context.Context, workspaceID, id string, patch entity.IncidentPatch) (bool, error)
	SetStatus(ctx context.Context, workspaceID, id string, from, to entity.IncidentStatus, at time.Time, resolution string) (bool, error)
	Reopen(ctx context.Context, workspaceID, id string, to entity.IncidentStatus, at time.Time) (bool, error)
	SetCustomFields(ctx context.Context, workspaceID, id string, fields map[string]string) (bool, error)
	ReplaceServices(ctx context.Context, workspaceID, incidentID string, serviceIDs []string) error
	LinkAlert(ctx context.Context, workspaceID, incidentID, alertID string) error
	UnlinkAlert(ctx context.Context, workspaceID, incidentID, alertID string) error
	LinkedAlertIDs(ctx context.Context, incidentID string) ([]string, error)
	Relate(ctx context.Context, workspaceID, incidentID, relatedID string, kind entity.IncidentRelationKind) (entity.IncidentRelation, error)
	Unrelate(ctx context.Context, workspaceID, incidentID, relationID string) error
	AppendEvent(ctx context.Context, event entity.IncidentEvent) (entity.IncidentEvent, error)
	ListEvents(ctx context.Context, workspaceID, incidentID string, categories []entity.IncidentEventCategory, after entity.TimelineCursor, limit int) ([]entity.IncidentEvent, error)
	GetEvent(ctx context.Context, workspaceID, eventID string) (entity.IncidentEvent, error)
	GetEventForUpdate(ctx context.Context, workspaceID, eventID string) (entity.IncidentEvent, error)
	UpdateEvent(ctx context.Context, workspaceID, eventID string, edit entity.TimelineEdit, editedAt time.Time, editorUserID string) error
	AppendRevision(ctx context.Context, revision entity.IncidentEventRevision) error
	ListRevisions(ctx context.Context, workspaceID, eventID string) ([]entity.IncidentEventRevision, error)
	AddAttachment(ctx context.Context, attachment entity.IncidentEventAttachment) (entity.IncidentEventAttachment, error)
	GetAttachment(ctx context.Context, workspaceID, attachmentID string) (entity.IncidentEventAttachment, error)
	RemoveAttachment(ctx context.Context, workspaceID, attachmentID string) error
	CountAttachments(ctx context.Context, workspaceID, eventID string) (int, error)
	AddFollowup(ctx context.Context, f entity.IncidentFollowup) (entity.IncidentFollowup, error)
	SetFollowupDone(ctx context.Context, workspaceID, id string, done bool, at time.Time) (entity.IncidentFollowup, error)
	ListFollowups(ctx context.Context, workspaceID string) ([]entity.IncidentFollowup, error)
}
