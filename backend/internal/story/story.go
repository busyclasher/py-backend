package story

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrNotFound is returned when a story cannot be located in the repository.
	ErrNotFound = errors.New("story: not found")
)

// Visibility controls who can view or edit a story.
type Visibility string

const (
	VisibilityPrivate      Visibility = "private"
	VisibilityOrganization Visibility = "organization"
	VisibilityPublic       Visibility = "public"
)

// BlockType signals how a block should be rendered.
type BlockType string

const (
	BlockMarkdown BlockType = "markdown"
	BlockCode     BlockType = "code"
	BlockViz      BlockType = "visualization"
)

// Block represents a notebook cell or narrative section.
type Block struct {
	ID        string    `json:"id"`
	Type      BlockType `json:"type"`
	Language  string    `json:"language,omitempty"`
	Source    string    `json:"source"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Outputs   []Output  `json:"outputs"`
}

// Output contains rendered content associated with a block execution.
type Output struct {
	Kind     string `json:"kind"`
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

// Comment captures discussion anchored to a story or block.
type Comment struct {
	ID        string    `json:"id"`
	StoryID   string    `json:"storyId"`
	BlockID   string    `json:"blockId,omitempty"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// Story is the core collaborative artifact.
type Story struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Owners      []string   `json:"owners"`
	Visibility  Visibility `json:"visibility"`
	RevisionID  string     `json:"revisionId"`
	Blocks      []Block    `json:"blocks"`
	Comments    []Comment  `json:"comments"`
	Tags        []string   `json:"tags"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// Revision records the state of a story at a single point in time.
type Revision struct {
	ID        string    `json:"id"`
	StoryID   string    `json:"storyId"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
	Blocks    []Block   `json:"blocks"`
}

// ExecutionResult represents the outcome of re-running a story.
type ExecutionResult struct {
	StoryID    string    `json:"storyId"`
	Revision   string    `json:"revision"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt"`
	Status     string    `json:"status"`
	Blocks     []Block   `json:"blocks"`
	Logs       []string  `json:"logs"`
}

// ExecutionRequest captures inputs required to execute a story.
type ExecutionRequest struct {
	Story Story
	Actor string
}

// Runner abstracts execution backends used by the story service.
type Runner interface {
	Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error)
}

// Repository describes persistence operations for stories.
type Repository interface {
	Create(ctx context.Context, story Story) error
	Update(ctx context.Context, story Story) error
	Get(ctx context.Context, id string) (Story, error)
	List(ctx context.Context) ([]Story, error)
	AppendRevision(ctx context.Context, revision Revision) error
	ListRevisions(ctx context.Context, storyID string) ([]Revision, error)
	AppendComment(ctx context.Context, comment Comment) error
}

// Filter is used when searching for stories.
type Filter struct {
	Owner string
	Tag   string
	Query string
}

// Service exposes high-level story workflows.
type Service interface {
	CreateStory(ctx context.Context, input CreateStoryInput) (Story, error)
	ListStories(ctx context.Context, filter Filter) ([]Story, error)
	GetStory(ctx context.Context, id string) (Story, error)
	AppendBlock(ctx context.Context, id string, input BlockInput) (Story, error)
	RecordComment(ctx context.Context, id string, input CommentInput) (Story, error)
	ExecuteStory(ctx context.Context, id string, actor string) (ExecutionResult, error)
}

// CreateStoryInput captures the payload for a new story.
type CreateStoryInput struct {
	Title       string
	Description string
	Owners      []string
	Visibility  Visibility
	Tags        []string
	Blocks      []BlockInput
}

// BlockInput defines the data required to insert a new block.
type BlockInput struct {
	Type     BlockType
	Language string
	Source   string
	Position int
}

// CommentInput collects authoring information for a comment.
type CommentInput struct {
	Author  string
	Body    string
	BlockID string
}
