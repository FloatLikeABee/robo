package mcp_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"idongivaflyinfa/auth"
	"idongivaflyinfa/mcp"
)

func TestOwnerScopeEveryKindAndSearch(t *testing.T) {
	path, writer := seedTwoUsers(t)
	ctx := context.Background()
	db, err := mcp.OpenReadWrite(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	kinds := []mcp.CreateNoteArgs{
		{Kind: mcp.KindNote, Title: "Ada note", Body: "alpha dock"},
		{Kind: mcp.KindTodo, Title: "Ada todo", Body: "walk north"},
		{Kind: mcp.KindBigNote, Title: "Ada story", Body: "handover keys", Idea: "night"},
		{Kind: mcp.KindTimeline, Title: "Ada timeline", Body: "09:00 opened the gate"},
		{Kind: mcp.KindResearch, Title: "Ada research", Body: "tide delay", Prompt: "tide delay"},
		{Kind: mcp.KindStick, Title: "Ada stick", Body: "west gate"},
		{Kind: mcp.KindTask, Title: "Ada task", Body: "board item"},
	}
	var created []mcp.NoteRecord
	for _, in := range kinds {
		rec, err := mcp.CreateNote(ctx, db, "ada-id", in)
		if err != nil {
			t.Fatalf("create %s: %v", in.Kind, err)
		}
		if rec.ID <= 0 || rec.RecordID != mcp.RecordID(rec.Kind, rec.ID) || rec.Visibility != "private" {
			t.Fatalf("created = %+v", rec)
		}
		got, err := mcp.GetNote(ctx, db, "ada-id", rec.RecordID)
		if err != nil || got.Kind != rec.Kind || got.ID != rec.ID || got.Title == "" {
			t.Fatalf("get %s = %+v err=%v", in.Kind, got, err)
		}
		if _, err := mcp.GetNote(ctx, db, "bea-id", rec.RecordID); !errors.Is(err, mcp.ErrNotFound) {
			t.Fatalf("bea get %s = %v", in.Kind, err)
		}
		if strings.Contains(errString(err), "forbidden") {
			t.Fatalf("bea error leaked forbidden: %v", err)
		}
		created = append(created, rec)
	}

	listed, err := mcp.ListNotes(ctx, db, "ada-id", mcp.KindStick, 10)
	if err != nil || len(listed.Notes) != 1 {
		t.Fatalf("ada stick list = %+v err=%v", listed, err)
	}
	alias, err := mcp.ListStickNotes(ctx, db, "ada-id", 10)
	if err != nil || len(alias.Records) != 1 || alias.Records[0].ID != created[5].ID {
		t.Fatalf("alias list = %+v err=%v", alias, err)
	}
	beaAlias, err := mcp.ListStickNotes(ctx, db, "bea-id", 10)
	if err != nil || len(beaAlias.Records) != 0 {
		t.Fatalf("bea alias = %+v", beaAlias)
	}

	omitted, err := mcp.CreateNote(ctx, db, "ada-id", mcp.CreateNoteArgs{Title: "plain note", Body: "no kind"})
	if err != nil || omitted.Kind != mcp.KindNote || omitted.ID <= 0 {
		t.Fatalf("omitted kind = %+v err=%v", omitted, err)
	}

	first, err := mcp.CreateNote(ctx, db, "ada-id", mcp.CreateNoteArgs{Title: "once", Body: "only", IdempotencyKey: "k1"})
	if err != nil {
		t.Fatal(err)
	}
	again, err := mcp.CreateNote(ctx, db, "ada-id", mcp.CreateNoteArgs{Title: "twice", Body: "other", IdempotencyKey: "k1"})
	if err != nil || again.ID != first.ID {
		t.Fatalf("idempotent = %+v want %d err=%v", again, first.ID, err)
	}

	note := created[0]
	if _, err := mcp.UpdateNote(ctx, db, "ada-id", note.RecordID, "", "changed-body", nil, 99); !errors.Is(err, mcp.ErrConflict) {
		t.Fatalf("stale = %v", err)
	}
	got, err := mcp.GetNote(ctx, db, "ada-id", note.RecordID)
	if err != nil || strings.Contains(got.Body, "changed-body") {
		t.Fatalf("conflict wrote: %+v err=%v", got, err)
	}
	updated, err := mcp.UpdateNote(ctx, db, "ada-id", note.RecordID, "", "changed-body", []string{"Dock"}, 0)
	if err != nil || updated.Version != 2 || !containsTag(updated.Tags, "dock") {
		t.Fatalf("update = %+v err=%v", updated, err)
	}

	if err := mcp.ArchiveNote(ctx, db, "bea-id", note.RecordID); !errors.Is(err, mcp.ErrNotFound) {
		t.Fatalf("bea archive = %v", err)
	}
	if err := mcp.ArchiveNote(ctx, db, "ada-id", note.RecordID); err != nil {
		t.Fatal(err)
	}
	if _, err := mcp.GetNote(ctx, db, "ada-id", note.RecordID); !errors.Is(err, mcp.ErrNotFound) {
		t.Fatalf("archived get = %v", err)
	}
	after, err := mcp.ListNotes(ctx, db, "ada-id", mcp.KindNote, 50)
	if err != nil {
		t.Fatal(err)
	}
	for _, rec := range after.Notes {
		if rec.RecordID == note.RecordID {
			t.Fatalf("list kept archived: %+v", rec)
		}
	}

	bea, err := mcp.CreateNote(ctx, db, "bea-id", mcp.CreateNoteArgs{Title: "Bea secret", Body: "west gate"})
	if err != nil {
		t.Fatal(err)
	}
	hits, err := mcp.SearchMyNotes(ctx, db, "ada-id", "west gate", 10)
	if err != nil {
		t.Fatal(err)
	}
	blob := noteBlob(hits.Notes)
	if strings.Contains(blob, "Bea secret") || strings.Contains(blob, bea.Title) {
		t.Fatalf("search leaked bea: %s", blob)
	}
	junk, err := mcp.SearchMyNotes(ctx, db, "ada-id", `"" OR * NEAR(`, 10)
	if err != nil {
		t.Fatalf("junk query: %v", err)
	}
	if len(junk.Notes) != 0 {
		t.Fatalf("junk hits = %+v", junk)
	}
	_ = writer
}

func TestReadOnlyPATCannotCreate(t *testing.T) {
	path, writer := seedTwoUsers(t)
	before := noteCount(t, writer)
	id := mcp.Identity{UserID: "ada-id", Scopes: []string{auth.ScopeNotesRead}}
	if err := mcp.RequireScope(id, auth.ScopeNotesWrite); !errors.Is(err, mcp.ErrForbiddenScope) {
		t.Fatalf("scope = %v", err)
	}
	notes, err := mcp.OpenReadWrite(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = notes.Close() })
	if err := mcp.RequireScope(id, auth.ScopeNotesWrite); err != nil {
		if noteCount(t, writer) != before {
			t.Fatal("scope check wrote a row")
		}
		return
	}
	t.Fatal("read-only identity was allowed to write")
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func containsTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func noteBlob(notes []mcp.NoteRecord) string {
	var b strings.Builder
	for _, n := range notes {
		b.WriteString(n.Title)
		b.WriteByte('\n')
		b.WriteString(n.Body)
		b.WriteByte('\n')
	}
	return b.String()
}
