package mcp

import "github.com/romancha/salmon/internal/models"

// --- Search Notes ---

// SearchNotesInput is the input for the search_notes tool.
type SearchNotesInput struct {
	Query string `json:"query" jsonschema:"full-text search query"`
	Limit int    `json:"limit,omitempty" jsonschema:"max results (default 20, max 200)"`
	Tag   string `json:"tag,omitempty" jsonschema:"filter by tag name"`
}

// SearchNotesOutput is the output for the search_notes tool.
type SearchNotesOutput struct {
	Notes []models.Note `json:"notes"`
}

// --- Get Note ---

// GetNoteInput is the input for the get_note tool.
type GetNoteInput struct {
	ID string `json:"id" jsonschema:"note ID (hub UUID)"`
}

// GetNoteOutput is the output for the get_note tool.
type GetNoteOutput struct {
	models.Note
}

// --- List Notes ---

// ListNotesInput is the input for the list_notes tool.
type ListNotesInput struct {
	Tag     string `json:"tag,omitempty" jsonschema:"filter by tag name"`
	Sort    string `json:"sort,omitempty" jsonschema:"sort column: modified_at, created_at, or title"`
	Order   string `json:"order,omitempty" jsonschema:"sort order: asc or desc"`
	Limit   int    `json:"limit,omitempty" jsonschema:"max results (max 200)"`
	Trashed string `json:"trashed,omitempty" jsonschema:"filter by trashed status: true or false"`
}

// ListNotesOutput is the output for the list_notes tool.
type ListNotesOutput struct {
	Notes []models.Note `json:"notes"`
}

// --- List Tags ---

// ListTagsInput is the input for the list_tags tool.
type ListTagsInput struct{}

// ListTagsOutput is the output for the list_tags tool.
type ListTagsOutput struct {
	Tags []models.Tag `json:"tags"`
}

// --- Get Attachment ---

// GetAttachmentInput is the input for the get_attachment tool.
type GetAttachmentInput struct {
	ID string `json:"id" jsonschema:"attachment ID (hub UUID)"`
}

// GetAttachmentOutput is the output for the get_attachment tool.
type GetAttachmentOutput struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Base64      string `json:"base64"`
}

// --- Sync Status ---

// SyncStatusInput is the input for the sync_status tool.
type SyncStatusInput struct{}

// SyncStatusOutput is the output for the sync_status tool.
type SyncStatusOutput struct {
	LastSyncAt          string   `json:"last_sync_at"`
	LastPushAt          string   `json:"last_push_at"`
	QueueSize           int      `json:"queue_size"`
	InitialSyncComplete string   `json:"initial_sync_complete"`
	ConflictCount       int      `json:"conflict_count"`
	ConflictNoteIDs     []string `json:"conflict_note_ids,omitempty"`
}

// --- List Backlinks ---

// ListBacklinksInput is the input for the list_backlinks tool.
type ListBacklinksInput struct {
	NoteID string `json:"note_id" jsonschema:"note ID to get backlinks for"`
}

// ListBacklinksOutput is the output for the list_backlinks tool.
type ListBacklinksOutput struct {
	Backlinks []models.Backlink `json:"backlinks"`
}

// --- Create Note ---

// CreateNoteInput is the input for the create_note tool.
type CreateNoteInput struct {
	Title string   `json:"title" jsonschema:"note title"`
	Body  string   `json:"body,omitempty" jsonschema:"note body (Markdown)"`
	Tags  []string `json:"tags,omitempty" jsonschema:"tags to assign to the note"`
}

// CreateNoteOutput is the output for the create_note tool.
type CreateNoteOutput struct {
	models.Note
}

// --- Update Note ---

// UpdateNoteInput is the input for the update_note tool.
type UpdateNoteInput struct {
	ID    string `json:"id" jsonschema:"note ID (hub UUID)"`
	Title string `json:"title,omitempty" jsonschema:"new title (optional)"`
	Body  string `json:"body" jsonschema:"new body (required)"`
}

// UpdateNoteOutput is the output for the update_note tool.
type UpdateNoteOutput struct {
	models.Note
}

// --- Trash Note ---

// TrashNoteInput is the input for the trash_note tool.
type TrashNoteInput struct {
	ID string `json:"id" jsonschema:"note ID (hub UUID)"`
}

// TrashNoteOutput is the output for the trash_note tool.
type TrashNoteOutput struct {
	models.Note
}

// --- Archive Note ---

// ArchiveNoteInput is the input for the archive_note tool.
type ArchiveNoteInput struct {
	ID string `json:"id" jsonschema:"note ID (hub UUID)"`
}

// ArchiveNoteOutput is the output for the archive_note tool.
type ArchiveNoteOutput struct {
	models.Note
}

// --- Add Tag ---

// AddTagInput is the input for the add_tag tool.
type AddTagInput struct {
	NoteID string `json:"note_id" jsonschema:"note ID to add the tag to"`
	Tag    string `json:"tag" jsonschema:"tag name to add"`
}

// AddTagOutput is the output for the add_tag tool.
type AddTagOutput struct {
	models.WriteQueueItem
}

// --- Rename Tag ---

// RenameTagInput is the input for the rename_tag tool.
type RenameTagInput struct {
	ID      string `json:"id" jsonschema:"tag ID (hub UUID)"`
	NewName string `json:"new_name" jsonschema:"new tag name"`
}

// RenameTagOutput is the output for the rename_tag tool.
type RenameTagOutput struct {
	models.WriteQueueItem
}

// --- Delete Tag ---

// DeleteTagInput is the input for the delete_tag tool.
type DeleteTagInput struct {
	ID string `json:"id" jsonschema:"tag ID (hub UUID)"`
}

// DeleteTagOutput is the output for the delete_tag tool.
type DeleteTagOutput struct {
	models.WriteQueueItem
}
