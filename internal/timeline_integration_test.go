package internal

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/opsybot/opsybot/internal/config"
	"github.com/opsybot/opsybot/internal/entity"
	"github.com/opsybot/opsybot/internal/pkg/postgres"
	"github.com/opsybot/opsybot/internal/pkg/secretbox"
	"github.com/opsybot/opsybot/internal/repository/incident"
	"github.com/opsybot/opsybot/internal/repository/transactor"
	"github.com/opsybot/opsybot/internal/repository/user"
	"github.com/opsybot/opsybot/internal/repository/workspace"
)

func TestIntegrationIncidentTimeline(t *testing.T) {
	dbURL := os.Getenv("OPSYBOT_TEST_POSTGRES_URL")
	if dbURL == "" {
		t.Skip("OPSYBOT_TEST_POSTGRES_URL not set")
	}

	client, cleanup, err := postgres.New(config.Postgres{
		URL:            dbURL,
		MaxOpenConns:   4,
		MaxIdleConns:   4,
		ConnectTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(cleanup)

	ctx := context.Background()
	migrator := &Migrator{Log: slog.New(slog.NewTextHandler(io.Discard, nil)), PG: client}
	if err := migrator.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	box, err := secretbox.New(config.Auth{})
	if err != nil {
		t.Fatalf("secretbox: %v", err)
	}
	users := user.New(client, box)
	workspaces := workspace.New(client)
	incidents := incident.New(client)

	stamp := time.Now().UnixNano()
	owner, err := users.Create(ctx, entity.NewUser{
		Email:    fmt.Sprintf("timeline+%d@example.test", stamp),
		Name:     "Timeline Owner",
		Timezone: "UTC",
	}, "argon2-not-verified-here")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = client.Querier(context.Background()).ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, owner.ID)
	})

	ws, err := workspaces.Create(ctx, entity.NewWorkspace{
		Name:     "Timeline IT",
		Slug:     fmt.Sprintf("timeline-it-%d", stamp),
		Timezone: "UTC",
	}, owner.ID)
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	t.Cleanup(func() {
		_, _ = client.Querier(context.Background()).ExecContext(context.Background(), `DELETE FROM workspaces WHERE id = $1`, ws.ID)
	})

	number, err := incidents.NextNumber(ctx, ws.ID)
	if err != nil {
		t.Fatalf("next number: %v", err)
	}
	declaredAt := time.Now().UTC().Truncate(time.Second)
	inc, err := incidents.Create(ctx, entity.Incident{
		WorkspaceID:   ws.ID,
		Number:        number,
		Name:          "Timeline integration",
		SeverityLevel: "SEV3",
		Status:        entity.IncidentStatusDeclared,
		DeclaredAt:    declaredAt,
	})
	if err != nil {
		t.Fatalf("create incident: %v", err)
	}

	t.Run("idempotency key replay inside a transaction returns the first entry", func(t *testing.T) {
		tx := transactor.New(client)
		entry := entity.IncidentEvent{
			IncidentID:     inc.ID,
			WorkspaceID:    ws.ID,
			At:             declaredAt,
			Kind:           entity.IncidentEventNote,
			Category:       entity.IncidentCategoryCommunication,
			Source:         entity.IncidentSourceChat,
			Text:           "Relayed from chat",
			Actor:          "Priya",
			IdempotencyKey: fmt.Sprintf("chat-%d", stamp),
		}
		var first, replay entity.IncidentEvent
		if err := tx.WithTx(ctx, func(ctx context.Context) error {
			var err error
			first, err = incidents.AppendEvent(ctx, entry)
			return err
		}); err != nil {
			t.Fatalf("append: %v", err)
		}
		entry.Text = "Relayed from chat (retry)"
		if err := tx.WithTx(ctx, func(ctx context.Context) error {
			var err error
			replay, err = incidents.AppendEvent(ctx, entry)
			return err
		}); err != nil {
			t.Fatalf("replay append: %v", err)
		}
		if replay.ID != first.ID {
			t.Fatalf("replay created a second entry: %s vs %s", replay.ID, first.ID)
		}
		if replay.Text != "Relayed from chat" {
			t.Fatalf("replay overwrote the original text: %q", replay.Text)
		}
	})

	t.Run("concurrent replay of one key yields one entry", func(t *testing.T) {
		tx := transactor.New(client)
		key := fmt.Sprintf("race-%d", stamp)
		const racers = 4

		start := make(chan struct{})
		ids := make([]string, racers)
		errs := make([]error, racers)
		var wg sync.WaitGroup
		for i := range racers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				errs[i] = tx.WithTx(ctx, func(ctx context.Context) error {
					got, err := incidents.AppendEvent(ctx, entity.IncidentEvent{
						IncidentID:     inc.ID,
						WorkspaceID:    ws.ID,
						At:             declaredAt,
						Kind:           entity.IncidentEventNote,
						Category:       entity.IncidentCategoryCommunication,
						Source:         entity.IncidentSourceChat,
						Text:           fmt.Sprintf("racer %d", i),
						IdempotencyKey: key,
					})
					ids[i] = got.ID
					return err
				})
			}()
		}
		close(start)
		wg.Wait()

		for i, err := range errs {
			if err != nil {
				t.Fatalf("racer %d failed instead of resolving to the winning entry: %v", i, err)
			}
		}
		for i, id := range ids {
			if id == "" || id != ids[0] {
				t.Fatalf("racer %d returned %q, want the same entry as %q", i, id, ids[0])
			}
		}

		var stored int
		row := client.Querier(ctx).QueryRowContext(ctx,
			`SELECT count(*) FROM incident_events WHERE incident_id = $1 AND idempotency_key = $2`, inc.ID, key)
		if err := row.Scan(&stored); err != nil {
			t.Fatalf("count entries: %v", err)
		}
		if stored != 1 {
			t.Fatalf("%d rows stored for one idempotency key, want 1", stored)
		}
	})

	t.Run("keyset pagination returns every entry exactly once", func(t *testing.T) {
		var existing int
		row := client.Querier(ctx).QueryRowContext(ctx,
			`SELECT count(*) FROM incident_events WHERE incident_id = $1`, inc.ID)
		if err := row.Scan(&existing); err != nil {
			t.Fatalf("count existing entries: %v", err)
		}

		const total = 7
		for i := range total {
			if _, err := incidents.AppendEvent(ctx, entity.IncidentEvent{
				IncidentID:  inc.ID,
				WorkspaceID: ws.ID,
				At:          declaredAt.Add(time.Duration(i) * time.Second),
				Kind:        entity.IncidentEventNote,
				Category:    entity.IncidentCategoryObservation,
				Source:      entity.IncidentSourceUI,
				Text:        fmt.Sprintf("page probe %d", i),
			}); err != nil {
				t.Fatalf("append probe %d: %v", i, err)
			}
		}

		seen := map[string]int{}
		cursor := entity.TimelineCursor{}
		for pages := 0; pages < existing+total+2; pages++ {
			page, err := incidents.ListEvents(ctx, ws.ID, inc.ID, nil, cursor, 2)
			if err != nil {
				t.Fatalf("list events: %v", err)
			}
			if len(page) == 0 {
				break
			}
			for _, e := range page {
				seen[e.ID]++
			}
			last := page[len(page)-1]
			cursor = entity.TimelineCursor{At: last.At, ID: last.ID}
		}

		if len(seen) != existing+total {
			t.Fatalf("paged %d distinct entries, want %d", len(seen), existing+total)
		}
		for id, count := range seen {
			if count != 1 {
				t.Fatalf("entry %s returned %d times", id, count)
			}
		}
	})

	t.Run("repository serves the page-plus-lookahead row", func(t *testing.T) {
		var existing int
		row := client.Querier(ctx).QueryRowContext(ctx,
			`SELECT count(*) FROM incident_events WHERE incident_id = $1`, inc.ID)
		if err := row.Scan(&existing); err != nil {
			t.Fatalf("count existing entries: %v", err)
		}
		for i := existing; i < entity.TimelineFetchLimit; i++ {
			if _, err := incidents.AppendEvent(ctx, entity.IncidentEvent{
				IncidentID:  inc.ID,
				WorkspaceID: ws.ID,
				At:          declaredAt.Add(time.Duration(i) * time.Millisecond),
				Kind:        entity.IncidentEventNote,
				Category:    entity.IncidentCategoryObservation,
				Source:      entity.IncidentSourceUI,
				Text:        fmt.Sprintf("lookahead filler %d", i),
			}); err != nil {
				t.Fatalf("append filler %d: %v", i, err)
			}
		}

		page, err := incidents.ListEvents(ctx, ws.ID, inc.ID, nil, entity.TimelineCursor{}, entity.TimelineFetchLimit)
		if err != nil {
			t.Fatalf("list events: %v", err)
		}
		if len(page) != entity.TimelineFetchLimit {
			t.Fatalf("repository returned %d rows for a %d-row request; the lookahead row that "+
				"signals another page exists is being clamped away", len(page), entity.TimelineFetchLimit)
		}
	})

	t.Run("concurrent edits serialise instead of losing one", func(t *testing.T) {
		tx := transactor.New(client)
		entry, err := incidents.AppendEvent(ctx, entity.IncidentEvent{
			IncidentID:  inc.ID,
			WorkspaceID: ws.ID,
			At:          declaredAt.Add(3 * time.Minute),
			Kind:        entity.IncidentEventNote,
			Category:    entity.IncidentCategoryObservation,
			Source:      entity.IncidentSourceUI,
			Text:        "A",
		})
		if err != nil {
			t.Fatalf("append entry: %v", err)
		}

		edit := func(ctx context.Context, next string) error {
			locked, err := incidents.GetEventForUpdate(ctx, ws.ID, entry.ID)
			if err != nil {
				return err
			}
			if err := incidents.AppendRevision(ctx, entity.IncidentEventRevision{
				EventID:     entry.ID,
				WorkspaceID: ws.ID,
				At:          declaredAt.Add(4 * time.Minute),
				EditorLabel: next,
				Text:        locked.Text,
				Category:    locked.Category,
			}); err != nil {
				return err
			}
			return incidents.UpdateEvent(ctx, ws.ID, entry.ID, entity.TimelineEdit{
				Text: next, Category: entity.IncidentCategoryObservation,
			}, declaredAt.Add(4*time.Minute), "")
		}

		firstRead := make(chan struct{})
		secondStarted := make(chan struct{})
		errs := make([]error, 2)
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[0] = tx.WithTx(ctx, func(ctx context.Context) error {
				locked, err := incidents.GetEventForUpdate(ctx, ws.ID, entry.ID)
				if err != nil {
					return err
				}
				close(firstRead)
				<-secondStarted
				time.Sleep(100 * time.Millisecond)
				if err := incidents.AppendRevision(ctx, entity.IncidentEventRevision{
					EventID:     entry.ID,
					WorkspaceID: ws.ID,
					At:          declaredAt.Add(4 * time.Minute),
					EditorLabel: "B",
					Text:        locked.Text,
					Category:    locked.Category,
				}); err != nil {
					return err
				}
				return incidents.UpdateEvent(ctx, ws.ID, entry.ID, entity.TimelineEdit{
					Text: "B", Category: entity.IncidentCategoryObservation,
				}, declaredAt.Add(4*time.Minute), "")
			})
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			<-firstRead
			close(secondStarted)
			errs[1] = tx.WithTx(ctx, func(ctx context.Context) error { return edit(ctx, "C") })
		}()
		wg.Wait()

		for i, err := range errs {
			if err != nil {
				t.Fatalf("edit %d: %v", i, err)
			}
		}

		revisions, err := incidents.ListRevisions(ctx, ws.ID, entry.ID)
		if err != nil {
			t.Fatalf("list revisions: %v", err)
		}
		if len(revisions) != 2 {
			t.Fatalf("%d revisions, want 2", len(revisions))
		}
		if revisions[0].Text == revisions[1].Text {
			t.Fatalf("both revisions captured %q: the second edit read a stale pre-image, so one edit vanished "+
				"from the audit trail", revisions[0].Text)
		}
		final, err := incidents.GetEvent(ctx, ws.ID, entry.ID)
		if err != nil {
			t.Fatalf("get event: %v", err)
		}
		if final.Text != "C" {
			t.Fatalf("final text = %q, want the second edit to have applied on top of the first", final.Text)
		}
	})

	t.Run("revisions and attachments survive an edit", func(t *testing.T) {
		entry, err := incidents.AppendEvent(ctx, entity.IncidentEvent{
			IncidentID:  inc.ID,
			WorkspaceID: ws.ID,
			At:          declaredAt.Add(time.Minute),
			Kind:        entity.IncidentEventNote,
			Category:    entity.IncidentCategoryObservation,
			Source:      entity.IncidentSourceUI,
			Text:        "Original observation",
		})
		if err != nil {
			t.Fatalf("append entry: %v", err)
		}
		if _, err := incidents.AddAttachment(ctx, entity.IncidentEventAttachment{
			EventID:     entry.ID,
			WorkspaceID: ws.ID,
			Kind:        entity.AttachmentLink,
			Label:       "Dashboard",
			URL:         "https://example.test/d/1",
		}); err != nil {
			t.Fatalf("add attachment: %v", err)
		}
		if err := incidents.AppendRevision(ctx, entity.IncidentEventRevision{
			EventID:     entry.ID,
			WorkspaceID: ws.ID,
			At:          declaredAt.Add(2 * time.Minute),
			EditorLabel: "Priya",
			Text:        entry.Text,
			Category:    entry.Category,
		}); err != nil {
			t.Fatalf("append revision: %v", err)
		}
		if err := incidents.UpdateEvent(ctx, ws.ID, entry.ID, entity.TimelineEdit{
			Text:     "Revised observation",
			Category: entity.IncidentCategoryDecision,
		}, declaredAt.Add(2*time.Minute), ""); err != nil {
			t.Fatalf("update event: %v", err)
		}

		updated, err := incidents.GetEvent(ctx, ws.ID, entry.ID)
		if err != nil {
			t.Fatalf("get event: %v", err)
		}
		if updated.Text != "Revised observation" || updated.Category != entity.IncidentCategoryDecision {
			t.Fatalf("edit not persisted: %+v", updated)
		}
		if updated.EditedAt.IsZero() {
			t.Fatalf("edited_at not stamped: %+v", updated)
		}
		if len(updated.Attachments) != 1 || updated.Attachments[0].Label != "Dashboard" {
			t.Fatalf("attachments lost on edit: %+v", updated.Attachments)
		}

		revisions, err := incidents.ListRevisions(ctx, ws.ID, entry.ID)
		if err != nil {
			t.Fatalf("list revisions: %v", err)
		}
		if len(revisions) != 1 || revisions[0].Text != "Original observation" {
			t.Fatalf("revision does not hold the previous text: %+v", revisions)
		}
	})
}
