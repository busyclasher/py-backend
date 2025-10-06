package story

import (
    "context"
    "strings"
    "time"

    "github.com/example/multistory/internal/executor"
    "github.com/example/multistory/internal/realtime"
    "github.com/example/multistory/pkg/id"
)

type service struct {
    repo   Repository
    runner executor.Runner
    hub    *realtime.Hub
    now    func() time.Time
}

// NewService wires dependencies for high-level operations on stories.
func NewService(repo Repository, runner executor.Runner, hub *realtime.Hub) Service {
    return &service{
        repo:   repo,
        runner: runner,
        hub:    hub,
        now:    func() time.Time { return time.Now().UTC() },
    }
}

func (s *service) CreateStory(ctx context.Context, input CreateStoryInput) (Story, error) {
    story := Story{
        ID:          id.New(),
        Title:       input.Title,
        Description: input.Description,
        Owners:      append([]string(nil), input.Owners...),
        Visibility:  input.Visibility,
        Tags:        append([]string(nil), input.Tags...),
        CreatedAt:   s.now(),
        UpdatedAt:   s.now(),
    }
    for idx, blockInput := range input.Blocks {
        story.Blocks = append(story.Blocks, s.newBlock(blockInput, idx))
    }
    story.RevisionID = id.New()
    if err := s.repo.Create(ctx, story); err != nil {
        return Story{}, err
    }
    if err := s.repo.AppendRevision(ctx, Revision{
        ID:        story.RevisionID,
        StoryID:   story.ID,
        Author:    firstOrDefault(input.Owners),
        Message:   "Initial draft",
        CreatedAt: s.now(),
        Blocks:    append([]Block(nil), story.Blocks...),
    }); err != nil {
        return Story{}, err
    }
    s.hub.Publish(realtime.Event{StoryID: story.ID, Type: "story.created", Payload: story})
    return story, nil
}

func (s *service) ListStories(ctx context.Context, filter Filter) ([]Story, error) {
    stories, err := s.repo.List(ctx)
    if err != nil {
        return nil, err
    }
    if filter == (Filter{}) {
        return stories, nil
    }
    filtered := stories[:0]
    for _, story := range stories {
        if filter.Owner != "" && !contains(story.Owners, filter.Owner) {
            continue
        }
        if filter.Tag != "" && !contains(story.Tags, filter.Tag) {
            continue
        }
        if filter.Query != "" && !strings.Contains(strings.ToLower(story.Title), strings.ToLower(filter.Query)) {
            continue
        }
        filtered = append(filtered, story)
    }
    return append([]Story(nil), filtered...), nil
}

func (s *service) GetStory(ctx context.Context, id string) (Story, error) {
    return s.repo.Get(ctx, id)
}

func (s *service) AppendBlock(ctx context.Context, storyID string, input BlockInput) (Story, error) {
    story, err := s.repo.Get(ctx, storyID)
    if err != nil {
        return Story{}, err
    }
    block := s.newBlock(input, len(story.Blocks))
    if input.Position > 0 && input.Position <= len(story.Blocks) {
        pos := input.Position - 1
        story.Blocks = append(story.Blocks[:pos], append([]Block{block}, story.Blocks[pos:]...)...)
    } else {
        story.Blocks = append(story.Blocks, block)
    }
    for idx := range story.Blocks {
        story.Blocks[idx].Position = idx
        story.Blocks[idx].UpdatedAt = s.now()
    }
    story.UpdatedAt = s.now()
    if err := s.repo.Update(ctx, story); err != nil {
        return Story{}, err
    }
    s.hub.Publish(realtime.Event{StoryID: story.ID, Type: "story.updated", Payload: story})
    return story, nil
}

func (s *service) RecordComment(ctx context.Context, storyID string, input CommentInput) (Story, error) {
    story, err := s.repo.Get(ctx, storyID)
    if err != nil {
        return Story{}, err
    }
    comment := Comment{
        ID:        id.New(),
        StoryID:   storyID,
        BlockID:   input.BlockID,
        Author:    input.Author,
        Body:      input.Body,
        CreatedAt: s.now(),
    }
    if err := s.repo.AppendComment(ctx, comment); err != nil {
        return Story{}, err
    }
    story.Comments = append(story.Comments, comment)
    s.hub.Publish(realtime.Event{StoryID: story.ID, Type: "comment.created", Payload: comment})
    return story, nil
}

func (s *service) ExecuteStory(ctx context.Context, id string, actor string) (ExecutionResult, error) {
    story, err := s.repo.Get(ctx, id)
    if err != nil {
        return ExecutionResult{}, err
    }
    result, err := s.runner.Execute(ctx, executor.RunRequest{Story: story, Actor: actor})
    if err != nil {
        return ExecutionResult{}, err
    }
    revision := Revision{
        ID:        result.Revision,
        StoryID:   story.ID,
        Author:    actor,
        Message:   "Automated execution",
        CreatedAt: result.FinishedAt,
        Blocks:    result.Blocks,
    }
    if err := s.repo.AppendRevision(ctx, revision); err != nil {
        return ExecutionResult{}, err
    }
    s.hub.Publish(realtime.Event{StoryID: story.ID, Type: "story.executed", Payload: result})
    return result, nil
}

func (s *service) newBlock(input BlockInput, position int) Block {
    now := s.now()
    return Block{
        ID:        id.New(),
        Type:      input.Type,
        Language:  input.Language,
        Source:    input.Source,
        Position:  position,
        CreatedAt: now,
        UpdatedAt: now,
        Outputs:   nil,
    }
}

func contains(values []string, target string) bool {
    for _, value := range values {
        if value == target {
            return true
        }
    }
    return false
}

func firstOrDefault(values []string) string {
    if len(values) == 0 {
        return "system"
    }
    return values[0]
}
