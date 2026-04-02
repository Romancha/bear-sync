package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTools registers all Salmon MCP tools on the given server.
func RegisterTools(s *mcp.Server, c *Client) {
	registerSearchNotes(s, c)
	registerGetNote(s, c)
	registerListNotes(s, c)
}

func registerSearchNotes(s *mcp.Server, c *Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_notes",
		Description: "Full-text search across Bear notes (titles and bodies)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input SearchNotesInput) (*mcp.CallToolResult, SearchNotesOutput, error) {
		return handleSearchNotes(ctx, c, input)
	})
}

func registerGetNote(s *mcp.Server, c *Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_note",
		Description: "Get a single Bear note by ID with full body, tags, attachments, and backlinks",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input GetNoteInput) (*mcp.CallToolResult, GetNoteOutput, error) {
		return handleGetNote(ctx, c, input)
	})
}

func registerListNotes(s *mcp.Server, c *Client) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_notes",
		Description: "List Bear notes (without body). Supports filtering by tag, sorting, and pagination",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input ListNotesInput) (*mcp.CallToolResult, ListNotesOutput, error) {
		return handleListNotes(ctx, c, input)
	})
}

func handleSearchNotes(ctx context.Context, c *Client, input SearchNotesInput) (*mcp.CallToolResult, SearchNotesOutput, error) {
	q := url.Values{}
	q.Set("q", input.Query)
	if input.Limit > 0 {
		q.Set("limit", strconv.Itoa(input.Limit))
	}
	if input.Tag != "" {
		q.Set("tag", input.Tag)
	}

	data, err := c.get(ctx, "/api/notes/search", q)
	if err != nil {
		return nil, SearchNotesOutput{}, err
	}

	var out SearchNotesOutput
	if err := json.Unmarshal(data, &out.Notes); err != nil {
		return nil, SearchNotesOutput{}, fmt.Errorf("parsing search results: %w", err)
	}

	return nil, out, nil
}

func handleGetNote(ctx context.Context, c *Client, input GetNoteInput) (*mcp.CallToolResult, GetNoteOutput, error) {
	data, err := c.get(ctx, "/api/notes/"+input.ID, nil)
	if err != nil {
		return nil, GetNoteOutput{}, err
	}

	var out GetNoteOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, GetNoteOutput{}, fmt.Errorf("parsing note: %w", err)
	}

	return nil, out, nil
}

func handleListNotes(ctx context.Context, c *Client, input ListNotesInput) (*mcp.CallToolResult, ListNotesOutput, error) {
	q := url.Values{}
	if input.Tag != "" {
		q.Set("tag", input.Tag)
	}
	if input.Sort != "" {
		q.Set("sort", input.Sort)
	}
	if input.Order != "" {
		q.Set("order", input.Order)
	}
	if input.Limit > 0 {
		q.Set("limit", strconv.Itoa(input.Limit))
	}
	if input.Trashed != "" {
		q.Set("trashed", input.Trashed)
	}

	data, err := c.get(ctx, "/api/notes", q)
	if err != nil {
		return nil, ListNotesOutput{}, err
	}

	var out ListNotesOutput
	if err := json.Unmarshal(data, &out.Notes); err != nil {
		return nil, ListNotesOutput{}, fmt.Errorf("parsing notes list: %w", err)
	}

	return nil, out, nil
}
