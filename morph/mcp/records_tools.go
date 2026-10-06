package mcp

import (
	"context"
	"database/sql"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolCount is whoami, the notes tools, and create, list, and get for five modules.
const ToolCount = 19

type limitInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"Maximum rows. Default 50, capped at 100."`
}

type idInput struct {
	ID int `json:"id" jsonschema:"Record id."`
}

type stickInput struct {
	Title       string `json:"title" jsonschema:"Stick note title. Required."`
	Description string `json:"description,omitempty" jsonschema:"Stick note text."`
	StartAt     string `json:"start_at,omitempty" jsonschema:"RFC3339 start. Default is the current UTC hour."`
	EndAt       string `json:"end_at,omitempty" jsonschema:"RFC3339 end. Default is one hour after start."`
}

type textInput struct {
	Title   string `json:"title,omitempty" jsonschema:"Title."`
	Content string `json:"content,omitempty" jsonschema:"Body text."`
	Idea    string `json:"idea,omitempty" jsonschema:"Story idea."`
	Prompt  string `json:"prompt,omitempty" jsonschema:"Research prompt."`
}

type eventInput struct {
	Title    string `json:"title,omitempty" jsonschema:"Event title."`
	Time     string `json:"time,omitempty" jsonschema:"RFC3339 time."`
	Detail   string `json:"detail,omitempty" jsonschema:"Event detail."`
	Reporter string `json:"reporter,omitempty" jsonschema:"Reporter name."`
	ID       string `json:"id,omitempty" jsonschema:"Event id."`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum rows. Default 50, capped at 100."`
}

func registerModuleTools(server *sdkmcp.Server, id Identity, tasks *sql.DB, notes *noteStore, recheck func() error, readOnly func(name, title, description string) *sdkmcp.Tool) {
	sdkmcp.AddTool(server, readOnly("list_stick_notes", "List stick notes", "List stick notes on the shared board, newest first."), listStick(tasks, recheck))
	sdkmcp.AddTool(server, readOnly("get_stick_note", "Get stick note", "Read one stick note by id. A missing id is not found."), getStick(tasks, recheck))
	sdkmcp.AddTool(server, writeTool("create_stick_note", "Create stick note", "Create a stick note on the shared board. The stored title starts with [morph-mcp]. Does not create a Notes and TODOs row."), createStick(notes, recheck))

	sdkmcp.AddTool(server, readOnly("list_timelines", "List timelines", "List the signed-in user's timelines."), listOwnedTool(id, tasks, recheck, "timeline"))
	sdkmcp.AddTool(server, readOnly("get_timeline", "Get timeline", "Read one of the signed-in user's timelines. Another user's id is not found."), getOwnedTool(id, tasks, recheck, "timeline"))
	sdkmcp.AddTool(server, writeTool("create_timeline", "Create timeline", "Store a timeline for the signed-in user from title and content. Does not run the generator and does not create a Notes and TODOs row."), createTimelineTool(id, notes, recheck))

	sdkmcp.AddTool(server, readOnly("list_stories", "List stories", "List the signed-in user's stories."), listOwnedTool(id, tasks, recheck, "big_note"))
	sdkmcp.AddTool(server, readOnly("get_story", "Get story", "Read one of the signed-in user's stories. Another user's id is not found."), getOwnedTool(id, tasks, recheck, "big_note"))
	sdkmcp.AddTool(server, writeTool("create_story", "Create story", "Store a story for the signed-in user. Does not run the generator and does not create a Notes and TODOs row."), createStoryTool(id, notes, recheck))

	sdkmcp.AddTool(server, readOnly("list_research", "List research", "List the signed-in user's research."), listOwnedTool(id, tasks, recheck, "research"))
	sdkmcp.AddTool(server, readOnly("get_research", "Get research", "Read one of the signed-in user's research items. Another user's id is not found."), getOwnedTool(id, tasks, recheck, "research"))
	sdkmcp.AddTool(server, writeTool("create_research", "Create research", "Store a research item for the signed-in user. Does not run the generator and does not create a Notes and TODOs row."), createResearchTool(id, notes, recheck))

	sdkmcp.AddTool(server, readOnly("list_event_logs", "List event logs", "List event logs through the Morph events API. Does not open Badger."), listEvents(recheck))
	sdkmcp.AddTool(server, readOnly("get_event_log", "Get event log", "Read one event log through the Morph events API."), getEvent(recheck))
	sdkmcp.AddTool(server, writeTool("create_event_log", "Create event log", "Create an event log through the Morph events API. Does not open Badger and does not create a Notes and TODOs row."), createEvent(recheck))
}

func writeTool(name, title, description string) *sdkmcp.Tool {
	closedWorld := false
	additive := false
	return &sdkmcp.Tool{
		Name:        name,
		Description: description,
		Annotations: &sdkmcp.ToolAnnotations{
			Title:           title,
			ReadOnlyHint:    false,
			DestructiveHint: &additive,
			OpenWorldHint:   &closedWorld,
		},
	}
}

func listStick(tasks *sql.DB, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, limitInput) (*sdkmcp.CallToolResult, ModuleList, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in limitInput) (*sdkmcp.CallToolResult, ModuleList, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, ModuleList{}, err
		}
		out, err := ListStickNotes(ctx, tasks, in.Limit)
		return nil, out, err
	}
}

func getStick(tasks *sql.DB, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, idInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in idInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, ModuleRecord{}, err
		}
		out, err := GetStickNote(ctx, tasks, in.ID)
		return nil, out, err
	}
}

func createStick(notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, stickInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in stickInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, ModuleRecord{}, err
		}
		db, err := notes.writer()
		if err != nil {
			return nil, ModuleRecord{}, err
		}
		out, err := CreateStickNote(ctx, db, in.Title, in.Description, in.StartAt, in.EndAt)
		return nil, out, err
	}
}

func listOwnedTool(id Identity, tasks *sql.DB, recheck func() error, table string) func(context.Context, *sdkmcp.CallToolRequest, limitInput) (*sdkmcp.CallToolResult, ModuleList, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in limitInput) (*sdkmcp.CallToolResult, ModuleList, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, ModuleList{}, err
		}
		out, err := ListOwned(ctx, tasks, id.UserID, table, in.Limit)
		return nil, out, err
	}
}

func getOwnedTool(id Identity, tasks *sql.DB, recheck func() error, table string) func(context.Context, *sdkmcp.CallToolRequest, idInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in idInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, ModuleRecord{}, err
		}
		out, err := GetOwned(ctx, tasks, id.UserID, table, in.ID)
		return nil, out, err
	}
}

func createTimelineTool(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, textInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in textInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
		return writeOwned(ctx, notes, recheck, func(db *sql.DB) (ModuleRecord, error) {
			return CreateTimeline(ctx, db, id.UserID, in.Title, in.Content)
		})
	}
}

func createStoryTool(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, textInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in textInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
		return writeOwned(ctx, notes, recheck, func(db *sql.DB) (ModuleRecord, error) {
			return CreateStory(ctx, db, id.UserID, in.Title, in.Idea, in.Content)
		})
	}
}

func createResearchTool(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, textInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in textInput) (*sdkmcp.CallToolResult, ModuleRecord, error) {
		return writeOwned(ctx, notes, recheck, func(db *sql.DB) (ModuleRecord, error) {
			return CreateResearch(ctx, db, id.UserID, in.Prompt, in.Content)
		})
	}
}

func writeOwned(ctx context.Context, notes *noteStore, recheck func() error, insert func(*sql.DB) (ModuleRecord, error)) (*sdkmcp.CallToolResult, ModuleRecord, error) {
	if err := callRecheck(recheck); err != nil {
		return nil, ModuleRecord{}, err
	}
	db, err := notes.writer()
	if err != nil {
		return nil, ModuleRecord{}, err
	}
	out, err := insert(db)
	return nil, out, err
}

func listEvents(recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, eventInput) (*sdkmcp.CallToolResult, EventLogResult, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in eventInput) (*sdkmcp.CallToolResult, EventLogResult, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, EventLogResult{}, err
		}
		out, err := ListEventLogs(ctx, in.Limit)
		return nil, out, err
	}
}

func getEvent(recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, eventInput) (*sdkmcp.CallToolResult, EventLogResult, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in eventInput) (*sdkmcp.CallToolResult, EventLogResult, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, EventLogResult{}, err
		}
		out, err := GetEventLog(ctx, in.ID)
		return nil, out, err
	}
}

func createEvent(recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, eventInput) (*sdkmcp.CallToolResult, EventLogResult, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in eventInput) (*sdkmcp.CallToolResult, EventLogResult, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, EventLogResult{}, err
		}
		out, err := CreateEventLog(ctx, in.Title, in.Time, in.Detail, in.Reporter)
		return nil, out, err
	}
}
