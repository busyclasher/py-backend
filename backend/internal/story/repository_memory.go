package story

import (
    "context"
    "sort"
    "sync"
)

type memoryRepository struct {
    mu        sync.RWMutex
    stories   map[string]Story
    revisions map[string][]Revision
}

// NewMemoryRepository returns an in-memory store suitable for prototypes.
func NewMemoryRepository() Repository {
    return &memoryRepository{
        stories:   make(map[string]Story),
        revisions: make(map[string][]Revision),
    }
}

func (m *memoryRepository) Create(_ context.Context, story Story) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.stories[story.ID] = cloneStory(story)
    return nil
}

func (m *memoryRepository) Update(_ context.Context, story Story) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if _, ok := m.stories[story.ID]; !ok {
        return ErrNotFound
    }
    m.stories[story.ID] = cloneStory(story)
    return nil
}

func (m *memoryRepository) Get(_ context.Context, id string) (Story, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    story, ok := m.stories[id]
    if !ok {
        return Story{}, ErrNotFound
    }
    return cloneStory(story), nil
}

func (m *memoryRepository) List(_ context.Context) ([]Story, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    stories := make([]Story, 0, len(m.stories))
    for _, s := range m.stories {
        stories = append(stories, cloneStory(s))
    }
    sort.Slice(stories, func(i, j int) bool {
        return stories[i].CreatedAt.After(stories[j].CreatedAt)
    })
    return stories, nil
}

func (m *memoryRepository) AppendRevision(_ context.Context, revision Revision) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if _, ok := m.stories[revision.StoryID]; !ok {
        return ErrNotFound
    }
    m.revisions[revision.StoryID] = append(m.revisions[revision.StoryID], revision)
    story := cloneStory(m.stories[revision.StoryID])
    story.RevisionID = revision.ID
    story.Blocks = revision.Blocks
    m.stories[story.ID] = story
    return nil
}

func (m *memoryRepository) ListRevisions(_ context.Context, storyID string) ([]Revision, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    revs := m.revisions[storyID]
    clones := make([]Revision, len(revs))
    copy(clones, revs)
    return clones, nil
}

func (m *memoryRepository) AppendComment(_ context.Context, comment Comment) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    story, ok := m.stories[comment.StoryID]
    if !ok {
        return ErrNotFound
    }
    story.Comments = append(story.Comments, comment)
    m.stories[story.ID] = story
    return nil
}

func cloneStory(s Story) Story {
    clone := s
    clone.Blocks = append([]Block(nil), s.Blocks...)
    clone.Comments = append([]Comment(nil), s.Comments...)
    clone.Owners = append([]string(nil), s.Owners...)
    clone.Tags = append([]string(nil), s.Tags...)
    return clone
}
