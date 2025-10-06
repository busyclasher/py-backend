package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/example/multistory/internal/story"
)

// Stub is a lightweight Runner used for local development.
type Stub struct{}

var _ story.Runner = (*Stub)(nil)

// NewStub creates a Runner that simulates execution.
func NewStub() *Stub {
	return &Stub{}
}

// Execute pretends to run the story and echoes deterministic output so the UI has data to render.
func (s *Stub) Execute(ctx context.Context, req story.ExecutionRequest) (story.ExecutionResult, error) {
	started := time.Now().UTC()
	select {
	case <-ctx.Done():
		return story.ExecutionResult{}, ctx.Err()
	case <-time.After(150 * time.Millisecond):
	}

	blocks := make([]story.Block, len(req.Story.Blocks))
	for i, block := range req.Story.Blocks {
		b := block
		b.Outputs = []story.Output{
			{
				Kind:     "text",
				MimeType: "text/plain",
				Data:     fmt.Sprintf("Simulated output for block %s", block.ID),
			},
		}
		blocks[i] = b
	}

	return story.ExecutionResult{
		StoryID:    req.Story.ID,
		Revision:   fmt.Sprintf("sim-%d", started.UnixNano()),
		StartedAt:  started,
		FinishedAt: time.Now().UTC(),
		Status:     "completed",
		Blocks:     blocks,
		Logs: []string{
			"Execution routed to stub runner",
			fmt.Sprintf("Actor: %s", req.Actor),
			fmt.Sprintf("Blocks executed: %d", len(blocks)),
		},
	}, nil
}
