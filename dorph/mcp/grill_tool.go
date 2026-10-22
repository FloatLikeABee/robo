package mcp

import (
	"context"
	"errors"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/robo/dorphai"

	"idongivaflyinfa/grill"
)

const grillAppURI = "ui://dorph/grill"

func registerGrillTools(server *sdkmcp.Server, id Identity, recheck func() error, complete func(context.Context, string) (string, error)) {
	if complete == nil {
		client := dorphai.NewClient(dorphai.LoadFromEnv())
		complete = func(ctx context.Context, prompt string) (string, error) {
			if !client.Configured() {
				return "", errors.New("AI service not configured")
			}
			return client.ChatCompletion(ctx, []dorphai.Message{{Role: "user", Content: prompt}})
		}
	}
	store := grill.NewStore()
	closedWorld := false
	additive := false
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        "note_grill",
		Description: "Interview the signed-in user about a title, then draft the body. action is start, answer, skip, or finish. start needs kind (note, todo, big_note, timeline, research, stick, or task) and title. answer, skip, and finish need id. Does not insert a row. Another user's grill id is not found.",
		Meta:        grillAppToolMeta(),
		Annotations: &sdkmcp.ToolAnnotations{
			Title:           "Note grill",
			ReadOnlyHint:    false,
			DestructiveHint: &additive,
			OpenWorldHint:   &closedWorld,
		},
	}, noteGrillTool(id, store, complete, recheck))
	server.AddResource(grillAppResource(), readGrillApp)
}

func grillAppToolMeta() sdkmcp.Meta {
	return sdkmcp.Meta{
		"ui": map[string]any{
			"resourceUri": grillAppURI,
			"visibility":  []string{"model", "app"},
		},
		"ui/resourceUri": grillAppURI,
	}
}

func grillAppResource() *sdkmcp.Resource {
	return &sdkmcp.Resource{
		URI:         grillAppURI,
		Name:        "dorph-grill",
		Title:       "AI assist",
		Description: "Interview the signed-in user, then draft a note.",
		MIMEType:    notesAppMIME,
		Meta:        sdkmcp.Meta{"ui": notesAppUIMeta()},
	}
}

func readGrillApp(context.Context, *sdkmcp.ReadResourceRequest) (*sdkmcp.ReadResourceResult, error) {
	return &sdkmcp.ReadResourceResult{
		Contents: []*sdkmcp.ResourceContents{{
			URI:      grillAppURI,
			MIMEType: notesAppMIME,
			Text:     grill.AppHTML,
			Meta:     sdkmcp.Meta{"ui": notesAppUIMeta()},
		}},
	}, nil
}

type noteGrillInput struct {
	Action      string `json:"action" jsonschema:"start, answer, skip, or finish"`
	ID          string `json:"id,omitempty" jsonschema:"Grill id returned by start."`
	Kind        string `json:"kind,omitempty" jsonschema:"note, todo, big_note, timeline, research, stick, or task. Used by start."`
	Title       string `json:"title,omitempty" jsonschema:"Title. Required for start."`
	Description string `json:"description,omitempty" jsonschema:"Optional context or existing body."`
	Answer      string `json:"answer,omitempty" jsonschema:"The user's answer for action answer."`
}

func noteGrillTool(id Identity, store *grill.Store, complete func(context.Context, string) (string, error), recheck func() error) func(context.Context, *sdkmcp.CallToolRequest, noteGrillInput) (*sdkmcp.CallToolResult, grill.View, error) {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, in noteGrillInput) (*sdkmcp.CallToolResult, grill.View, error) {
		if err := callRecheck(recheck); err != nil {
			return nil, grill.View{}, err
		}
		view, err := store.Act(ctx, id.UserID, grill.Input{
			Action:      in.Action,
			ID:          in.ID,
			Kind:        in.Kind,
			Title:       in.Title,
			Description: in.Description,
			Answer:      in.Answer,
		}, complete)
		if err != nil {
			return nil, grill.View{}, err
		}
		return nil, view, nil
	}
}
