package realtime

import (
    "encoding/json"
    "sync"
)

// Event represents a message delivered to listeners on a story channel.
type Event struct {
    StoryID string      `json:"storyId"`
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}

// Hub fan-outs events to interested subscribers per story.
type Hub struct {
    mu           sync.RWMutex
    subscribers  map[string]map[chan Event]struct{}
}

// NewHub constructs an empty broadcaster.
func NewHub() *Hub {
    return &Hub{subscribers: make(map[string]map[chan Event]struct{})}
}

// Publish sends the event to all subscribers, dropping messages on slow listeners.
func (h *Hub) Publish(event Event) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    subs := h.subscribers[event.StoryID]
    for ch := range subs {
        select {
        case ch <- event:
        default:
            // Drop the message if the receiver is not keeping up to keep the hub snappy.
        }
    }
}

// Subscribe attaches a new channel to the story ID, returning a cancel func to release resources.
func (h *Hub) Subscribe(storyID string) (<-chan Event, func()) {
    ch := make(chan Event, 8)
    h.mu.Lock()
    defer h.mu.Unlock()
    if _, ok := h.subscribers[storyID]; !ok {
        h.subscribers[storyID] = make(map[chan Event]struct{})
    }
    h.subscribers[storyID][ch] = struct{}{}

    cancel := func() {
        h.mu.Lock()
        defer h.mu.Unlock()
        if subs, ok := h.subscribers[storyID]; ok {
            delete(subs, ch)
            if len(subs) == 0 {
                delete(h.subscribers, storyID)
            }
        }
        close(ch)
    }
    return ch, cancel
}

// Marshal prepares the event payload for SSE delivery.
func (e Event) Marshal() ([]byte, error) {
    return json.Marshal(e)
}
