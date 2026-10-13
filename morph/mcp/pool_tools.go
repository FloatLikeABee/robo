package mcp

import (
	"context"
	"database/sql"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"idongivaflyinfa/auth"
)

// PoolToolCount is preview/share/unshare/list plus the five pool index tools.
const PoolToolCount = 9

func registerPoolTools(server *sdkmcp.Server, id Identity, tasks *sql.DB, notes *noteStore, recheck func() error, readOnly func(name, title, description string) *sdkmcp.Tool) {
	sdkmcp.AddTool(server, writeTool("preview_share", "Preview share", "Scan a note and return a snapshot plus a confirm_token. Requires an explicit yes before share_note."), previewShareTool(id, notes, recheck))
	sdkmcp.AddTool(server, writeTool("share_note", "Share note", "Publish the previewed snapshot. confirm_token is required. Secrets or personal data block the share."), shareNoteTool(id, notes, recheck))
	sdkmcp.AddTool(server, writeTool("unshare_note", "Unshare note", "Revoke the signed-in user's live share. Another user's share id is not found."), unshareNoteTool(id, notes, recheck))
	sdkmcp.AddTool(server, readOnly("list_my_shares", "List my shares", "List the signed-in user's shares. Another user's shares are not returned."), listMySharesTool(id, tasks, recheck))
	sdkmcp.AddTool(server, readOnly("pool_topics", "Pool topics", "List curated pool topics and live share counts. There is no public pool page."), poolTopicsTool(id, tasks, recheck))
	sdkmcp.AddTool(server, readOnly("pool_search", "Pool search", "Search live visibility=pool shares. Hits include an alias, never an email."), poolSearchTool(id, tasks, recheck))
	sdkmcp.AddTool(server, readOnly("pool_get", "Pool get", "Read one live share by id. Content is wrapped as untrusted. A revoked or hidden share is not found."), poolGetTool(id, tasks, recheck))
	sdkmcp.AddTool(server, writeTool("pool_save_copy", "Save pool copy", "Save a private copy of a live share with attribution to the author's alias and source share id."), poolSaveCopyTool(id, notes, recheck))
	sdkmcp.AddTool(server, writeTool("pool_report", "Report pool share", "Report a live share. Three distinct reporters hide it."), poolReportTool(id, notes, recheck))
}

type previewShareInput struct {
	ID          string   `json:"id" jsonschema:"kind:rowid such as note:12"`
	Visibility  string   `json:"visibility,omitempty" jsonschema:"pool or link. Default pool."`
	Topics      []string `json:"topics,omitempty" jsonschema:"Up to 5 curated topic slugs."`
	Tags        []string `json:"tags,omitempty"`
	License     string   `json:"license,omitempty" jsonschema:"Ignored. License is CC BY 4.0."`
	Attribution string   `json:"attribution,omitempty" jsonschema:"alias or anonymous."`
}

type shareNoteInput struct {
	ConfirmToken string `json:"confirm_token" jsonschema:"Token from preview_share. Expires in about five minutes."`
}

type shareIDInput struct {
	ShareID string `json:"share_id"`
}

type poolSearchInput struct {
	Query string `json:"query,omitempty" jsonschema:"Search text. Only letters, digits, and underscores are kept."`
	Kind  string `json:"kind,omitempty"`
	Topic string `json:"topic,omitempty"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum rows. Default 25, capped at 25."`
}

type poolSaveInput struct {
	ShareID string `json:"share_id"`
	Kind    string `json:"kind,omitempty" jsonschema:"Private copy kind. Default note."`
}

type poolReportInput struct {
	ShareID string `json:"share_id"`
	Reason  string `json:"reason,omitempty"`
}

type shareList struct {
	Shares []ShareRecord `json:"shares"`
}

type topicList struct {
	Topics []PoolTopic `json:"topics"`
}

type poolSearchResult struct {
	Hits []PoolHit `json:"hits"`
}

func previewShareTool(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, previewShareInput) (*sdkmcp.CallToolResult, SharePreview, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in previewShareInput) (*sdkmcp.CallToolResult, SharePreview, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, SharePreview{}, err
		}
		if err := RequireScope(id, auth.ScopePoolPublish); err != nil {
			return nil, SharePreview{}, err
		}
		db, err := notes.writer()
		if err != nil {
			return nil, SharePreview{}, err
		}
		out, err := PreviewShare(ctx, db, id.UserID, in.ID, in.Visibility, in.Attribution, in.Topics, in.Tags)
		return nil, out, err
	}
}

func shareNoteTool(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, shareNoteInput) (*sdkmcp.CallToolResult, ShareRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in shareNoteInput) (*sdkmcp.CallToolResult, ShareRecord, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, ShareRecord{}, err
		}
		if err := RequireScope(id, auth.ScopePoolPublish); err != nil {
			return nil, ShareRecord{}, err
		}
		db, err := notes.writer()
		if err != nil {
			return nil, ShareRecord{}, err
		}
		out, err := ShareNote(ctx, db, id.UserID, in.ConfirmToken)
		return nil, out, err
	}
}

func unshareNoteTool(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, shareIDInput) (*sdkmcp.CallToolResult, ShareRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in shareIDInput) (*sdkmcp.CallToolResult, ShareRecord, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, ShareRecord{}, err
		}
		if err := RequireScope(id, auth.ScopePoolPublish); err != nil {
			return nil, ShareRecord{}, err
		}
		db, err := notes.writer()
		if err != nil {
			return nil, ShareRecord{}, err
		}
		out, err := UnshareNote(ctx, db, id.UserID, in.ShareID)
		return nil, out, err
	}
}

func listMySharesTool(id Identity, tasks *sql.DB, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, struct{}) (*sdkmcp.CallToolResult, shareList, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ struct{}) (*sdkmcp.CallToolResult, shareList, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, shareList{}, err
		}
		if err := RequireScope(id, auth.ScopePoolPublish); err != nil {
			return nil, shareList{}, err
		}
		out, err := ListMyShares(ctx, tasks, id.UserID)
		if out == nil {
			out = []ShareRecord{}
		}
		return nil, shareList{Shares: out}, err
	}
}

func poolTopicsTool(id Identity, tasks *sql.DB, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, struct{}) (*sdkmcp.CallToolResult, topicList, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ struct{}) (*sdkmcp.CallToolResult, topicList, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, topicList{}, err
		}
		if err := RequireScope(id, auth.ScopePoolRead); err != nil {
			return nil, topicList{}, err
		}
		out, err := ListPoolTopics(ctx, tasks)
		if out == nil {
			out = []PoolTopic{}
		}
		return nil, topicList{Topics: out}, err
	}
}

func poolSearchTool(id Identity, tasks *sql.DB, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, poolSearchInput) (*sdkmcp.CallToolResult, poolSearchResult, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in poolSearchInput) (*sdkmcp.CallToolResult, poolSearchResult, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, poolSearchResult{}, err
		}
		if err := RequireScope(id, auth.ScopePoolRead); err != nil {
			return nil, poolSearchResult{}, err
		}
		out, err := SearchPool(ctx, tasks, id.UserID, in.Query, in.Kind, in.Topic, in.Limit)
		if out == nil {
			out = []PoolHit{}
		}
		return nil, poolSearchResult{Hits: out}, err
	}
}

func poolGetTool(id Identity, tasks *sql.DB, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, shareIDInput) (*sdkmcp.CallToolResult, PoolGet, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in shareIDInput) (*sdkmcp.CallToolResult, PoolGet, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, PoolGet{}, err
		}
		if err := RequireScope(id, auth.ScopePoolRead); err != nil {
			return nil, PoolGet{}, err
		}
		out, err := GetPoolShare(ctx, tasks, id.UserID, in.ShareID)
		return nil, out, err
	}
}

func poolSaveCopyTool(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, poolSaveInput) (*sdkmcp.CallToolResult, NoteRecord, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in poolSaveInput) (*sdkmcp.CallToolResult, NoteRecord, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, NoteRecord{}, err
		}
		if err := RequireScope(id, auth.ScopePoolRead); err != nil {
			return nil, NoteRecord{}, err
		}
		if err := RequireScope(id, auth.ScopeNotesWrite); err != nil {
			return nil, NoteRecord{}, err
		}
		db, err := notes.writer()
		if err != nil {
			return nil, NoteRecord{}, err
		}
		out, err := SavePoolCopy(ctx, db, id.UserID, in.ShareID, in.Kind)
		return nil, out, err
	}
}

func poolReportTool(id Identity, notes *noteStore, recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, poolReportInput) (*sdkmcp.CallToolResult, map[string]any, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in poolReportInput) (*sdkmcp.CallToolResult, map[string]any, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, nil, err
		}
		if err := RequireScope(id, auth.ScopePoolRead); err != nil {
			return nil, nil, err
		}
		db, err := notes.writer()
		if err != nil {
			return nil, nil, err
		}
		rid, err := ReportPoolShare(ctx, db, id.UserID, in.ShareID, in.Reason)
		if err != nil {
			return nil, nil, err
		}
		return nil, map[string]any{"report_id": rid}, nil
	}
}
