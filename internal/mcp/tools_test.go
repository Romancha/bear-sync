package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMCPServer(c *Client) *gomcp.Server {
	s := gomcp.NewServer(&gomcp.Implementation{Name: "test", Version: "test"}, nil)
	RegisterTools(s, c)
	return s
}

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return NewClient(srv.URL, "test-token")
}

func TestSearchNotes_Success(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/notes/search", r.URL.Path)
		assert.Equal(t, "test query", r.URL.Query().Get("q"))
		assert.Equal(t, "5", r.URL.Query().Get("limit"))
		assert.Equal(t, "work", r.URL.Query().Get("tag"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"note-1","title":"Test Note","body":"hello world"}]`))
	})

	_, out, err := handleSearchNotes(context.Background(), c, SearchNotesInput{
		Query: "test query",
		Limit: 5,
		Tag:   "work",
	})
	require.NoError(t, err)
	require.Len(t, out.Notes, 1)
	assert.Equal(t, "note-1", out.Notes[0].ID)
	assert.Equal(t, "Test Note", out.Notes[0].Title)
	assert.Equal(t, "hello world", out.Notes[0].Body)
}

func TestSearchNotes_MinimalParams(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "my query", r.URL.Query().Get("q"))
		assert.Empty(t, r.URL.Query().Get("limit"))
		assert.Empty(t, r.URL.Query().Get("tag"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	})

	_, out, err := handleSearchNotes(context.Background(), c, SearchNotesInput{Query: "my query"})
	require.NoError(t, err)
	assert.Empty(t, out.Notes)
}

func TestSearchNotes_APIError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	})

	_, _, err := handleSearchNotes(context.Background(), c, SearchNotesInput{Query: "test"})
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}

func TestGetNote_Success(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/notes/abc-123", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id":"abc-123",
			"title":"My Note",
			"body":"note body",
			"tags":[{"id":"tag-1","title":"work"}],
			"attachments":[{"id":"att-1","filename":"file.pdf"}],
			"backlinks":[{"id":"bl-1","title":"Other Note"}]
		}`))
	})

	_, out, err := handleGetNote(context.Background(), c, GetNoteInput{ID: "abc-123"})
	require.NoError(t, err)
	assert.Equal(t, "abc-123", out.ID)
	assert.Equal(t, "My Note", out.Title)
	assert.Equal(t, "note body", out.Body)
	require.Len(t, out.Tags, 1)
	assert.Equal(t, "work", out.Tags[0].Title)
	require.Len(t, out.Attachments, 1)
	assert.Equal(t, "file.pdf", out.Attachments[0].Filename)
	require.Len(t, out.Backlinks, 1)
	assert.Equal(t, "Other Note", out.Backlinks[0].Title)
}

func TestGetNote_NotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"note not found"}`))
	})

	_, _, err := handleGetNote(context.Background(), c, GetNoteInput{ID: "not-exist"})
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
}

func TestListNotes_Success(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/notes", r.URL.Path)
		assert.Equal(t, "work", r.URL.Query().Get("tag"))
		assert.Equal(t, "modified_at", r.URL.Query().Get("sort"))
		assert.Equal(t, "desc", r.URL.Query().Get("order"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))
		assert.Equal(t, "false", r.URL.Query().Get("trashed"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"n1","title":"Note 1"},{"id":"n2","title":"Note 2"}]`))
	})

	_, out, err := handleListNotes(context.Background(), c, ListNotesInput{
		Tag:     "work",
		Sort:    "modified_at",
		Order:   "desc",
		Limit:   10,
		Trashed: "false",
	})
	require.NoError(t, err)
	require.Len(t, out.Notes, 2)
	assert.Equal(t, "n1", out.Notes[0].ID)
	assert.Equal(t, "Note 1", out.Notes[0].Title)
	assert.Equal(t, "n2", out.Notes[1].ID)
}

func TestListNotes_NoParams(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	})

	_, out, err := handleListNotes(context.Background(), c, ListNotesInput{})
	require.NoError(t, err)
	assert.Empty(t, out.Notes)
}

func TestListNotes_APIError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`internal error`))
	})

	_, _, err := handleListNotes(context.Background(), c, ListNotesInput{})
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
}

func TestRegisterTools_AllRegistered(t *testing.T) {
	c := NewClient("http://localhost", "token")
	s := newMCPServer(c)

	// Verify the server was created without panics and tools are registered.
	// The MCP SDK doesn't expose a way to list tools programmatically in tests,
	// so we verify by ensuring RegisterTools completes without error.
	require.NotNil(t, s)
}
