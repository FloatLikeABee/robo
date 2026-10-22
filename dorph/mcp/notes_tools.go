package mcp

import (
	"context"
	"database/sql"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"idongivaflyinfa/auth"
)

func registerUnifiedNoteTools(server *sdkmcp.Server, id Identity, tasks *sql.DB, notes *noteStore, recheck func() error, readOnly func(name, title, description string) *sdkmcp.Tool) {
	get := readOnly("get_note", "Get note", "Read one of the signed-in user's notes by kind:rowid. Another user's id is not found.")
	get.Meta = notesAppToolMeta()
	sdkmcp.AddTool(server, get, getNoteTool(id, tasks, recheck))

	list := readOnly("list_notes", "List notes", "List the signed-in user's notes. Optional kind is note, todo, big_note, timeline, research, stick, or task.")
	list.Meta = notesAppToolMeta()
	sdkmcp.AddTool(server, list, listNotesTool(id, tasks, recheck))

	sdkmcp.AddTool(server, writeTool("update_note", "Update note", "Patch the signed-in user's note. expected_version must match when sent. A mismatch is conflict."), updateNoteTool(id, notes, recheck))
	sdkmcp.AddTool(server, writeTool("append_to_note", "Append to note", "Append text to the signed-in user's note body."), appendNoteTool(id, notes, recheck))
	sdkmcp.AddTool(server, writeTool("archive_note", "Archive note", "Soft-delete the signed-in user's note. Later get, list, and search treat it as not found."), archiveNoteTool(id, notes, recheck))

	search := readOnly("search_my_notes", "Search my notes", "Search the signed-in user's unarchived notes. Another user's text is not returned.")
	search.Meta = notesAppToolMeta()
	sdkmcp.AddTool(server, search, searchNotesTool(id, tasks, recheck))
}

type noteIDInput struct {
	ID string `json:"id" jsonschema:"kind:rowid such as note:12"`
}

type listNotesInput struct {
	Kind  string `json:"kind,omitempty" jsonschema:"Optional kind filter."`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum rows. Default 50, capped at 100."`
}

type updateNoteInput struct {
	ID              string   `json:"id" jsonschema:"kind:rowid"`
	Title           string   `json:"title,omitempty" jsonschema:"Replacement title."`
	Body            string   `json:"body,omitempty" jsonschema:"Replacement body."`
	Tags            []string `json:"tags,omitempty" jsonschema:"Replacement tags."`
	ExpectedVersion int      `json:"expected_version,omitempty" jsonschema:"Current version. 0 means omitted."`
}

type appendNoteInput struct {
	ID              string `json:"id" jsonschema:"kind:rowid"`
	Body            string `json:"body" jsonschema:"Text to append."`
	ExpectedVersion int    `json:"expected_version,omitempty" jsonschema:"Current version. 0 means omitted."`
}

type searchNotesInput struct {
	Query string `json:"query" jsonschema:"Search text. Only letters, digits, and underscores are kept."`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum rows. Default 50, capped at 100."`
}

func getNoteTool(id Identity, tasks *sql.DB, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, noteIDInput) (*sdkmcp.CallToolResult, NoteRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in noteIDInput) (*sdkmcp.CallToolResult, NoteRecord, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, NoteRecord{}, err
		}
		if err := RequireScope(id, auth.ScopeNotesRead); err != nil {
			return nil, NoteRecord{}, err
		}
		out, err := GetNote(ctx, tasks, id.UserID, in.ID)
		return nil, out, err
	}
}

func listNotesTool(id Identity, tasks *sql.DB, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, listNotesInput) (*sdkmcp.CallToolResult, NoteList, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in listNotesInput) (*sdkmcp.CallToolResult, NoteList, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, NoteList{}, err
		}
		if err := RequireScope(id, auth.ScopeNotesRead); err != nil {
			return nil, NoteList{}, err
		}
		out, err := ListNotes(ctx, tasks, id.UserID, in.Kind, in.Limit)
		return nil, out, err
	}
}

func updateNoteTool(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, updateNoteInput) (*sdkmcp.CallToolResult, NoteRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in updateNoteInput) (*sdkmcp.CallToolResult, NoteRecord, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, NoteRecord{}, err
		}
		if err := RequireScope(id, auth.ScopeNotesWrite); err != nil {
			return nil, NoteRecord{}, err
		}
		db, err := notes.writer()
		if err != nil {
			return nil, NoteRecord{}, err
		}
		out, err := UpdateNote(ctx, db, id.UserID, in.ID, in.Title, in.Body, in.Tags, in.ExpectedVersion)
		return nil, out, err
	}
}

func appendNoteTool(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, appendNoteInput) (*sdkmcp.CallToolResult, NoteRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in appendNoteInput) (*sdkmcp.CallToolResult, NoteRecord, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, NoteRecord{}, err
		}
		if err := RequireScope(id, auth.ScopeNotesWrite); err != nil {
			return nil, NoteRecord{}, err
		}
		db, err := notes.writer()
		if err != nil {
			return nil, NoteRecord{}, err
		}
		out, err := AppendNote(ctx, db, id.UserID, in.ID, in.Body, in.ExpectedVersion)
		return nil, out, err
	}
}

func archiveNoteTool(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, noteIDInput) (*sdkmcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in noteIDInput) (*sdkmcp.CallToolResult, map[string]any, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, nil, err
		}
		if err := RequireScope(id, auth.ScopeNotesWrite); err != nil {
			return nil, nil, err
		}
		db, err := notes.writer()
		if err != nil {
			return nil, nil, err
		}
		if err := ArchiveNote(ctx, db, id.UserID, in.ID); err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"archived": true, "id": in.ID}, nil
	}
}

func searchNotesTool(id Identity, tasks *sql.DB, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, searchNotesInput) (*sdkmcp.CallToolResult, NoteList, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in searchNotesInput) (*sdkmcp.CallToolResult, NoteList, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, NoteList{}, err
		}
		if err := RequireScope(id, auth.ScopeNotesRead); err != nil {
			return nil, NoteList{}, err
		}
		out, err := SearchMyNotes(ctx, tasks, id.UserID, in.Query, in.Limit)
		return nil, out, err
	}
}
