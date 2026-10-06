package mcp_test

import (
	"context"
	"strings"
	"testing"

	"idongivaflyinfa/mcp"
)

func TestCreateStickNoteAndOwnedRecords(t *testing.T) {
	path, writer := seedTwoUsers(t)
	ctx := context.Background()
	notes, err := mcp.OpenReadWrite(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = notes.Close() })

	beforeNotes := noteCount(t, writer)
	stick, err := mcp.CreateStickNote(ctx, notes, "Dock check", "west gate", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if stick.ID <= 0 || !strings.HasPrefix(stick.Title, "[morph-mcp]") || stick.StartAt == "" || stick.EndAt == "" {
		t.Fatalf("stick = %+v", stick)
	}
	if noteCount(t, writer) != beforeNotes {
		t.Fatal("stick note wrote a Notes and TODOs row")
	}
	listed, err := mcp.ListStickNotes(ctx, notes, 10)
	if err != nil || len(listed.Records) == 0 || listed.Records[0].ID != stick.ID {
		t.Fatalf("list = %+v err=%v", listed, err)
	}

	story, err := mcp.CreateStory(ctx, notes, "ada-id", "Night shift", "handover", "keys are in the box")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mcp.GetOwned(ctx, notes, "bea-id", "big_note", story.ID); err != mcp.ErrNotFound {
		t.Fatalf("bea get = %v", err)
	}
	got, err := mcp.GetOwned(ctx, notes, "ada-id", "big_note", story.ID)
	if err != nil || !strings.Contains(got.Body, "keys are in the box") {
		t.Fatalf("ada story = %+v err=%v", got, err)
	}

	research, err := mcp.CreateResearch(ctx, notes, "ada-id", "tide delays", "north dock")
	if err != nil {
		t.Fatal(err)
	}
	var status string
	if err := notes.QueryRow(`SELECT status FROM research WHERE id = ?`, research.ID).Scan(&status); err != nil || status != "complete" {
		t.Fatalf("status = %q err=%v", status, err)
	}
}
