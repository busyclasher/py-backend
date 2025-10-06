package server

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "strings"
    "time"

    realtimepkg "github.com/example/multistory/internal/realtime"
    storypkg "github.com/example/multistory/internal/story"
)

type handler struct {
    stories storypkg.Service
    hub     *realtimepkg.Hub
}

func newRouter(cfg Config, svc storypkg.Service, hub *realtimepkg.Hub) http.Handler {
    h := handler{stories: svc, hub: hub}
    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", h.health)
    mux.HandleFunc("/api/stories", h.handleStories)
    mux.HandleFunc("/api/stories/", h.handleStoryByID)

    return withLogging(withCORS(cfg.AllowedOrigins, mux))
}

func (h handler) health(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h handler) handleStories(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        h.listStories(w, r)
    case http.MethodPost:
        h.createStory(w, r)
    case http.MethodOptions:
        w.WriteHeader(http.StatusNoContent)
    default:
        writeError(w, http.StatusMethodNotAllowed, "method not allowed")
    }
}

func (h handler) handleStoryByID(w http.ResponseWriter, r *http.Request) {
    if !strings.HasPrefix(r.URL.Path, "/api/stories/") {
        writeError(w, http.StatusNotFound, "not found")
        return
    }
    id := strings.TrimPrefix(r.URL.Path, "/api/stories/")
    switch {
    case strings.HasSuffix(id, "/blocks"):
        storyID := strings.TrimSuffix(id, "/blocks")
        if idx := strings.Index(storyID, "/"); idx != -1 {
            writeError(w, http.StatusNotFound, "invalid path")
            return
        }
        h.appendBlock(w, r, storyID)
        return
    case strings.HasSuffix(id, "/comments"):
        storyID := strings.TrimSuffix(id, "/comments")
        if idx := strings.Index(storyID, "/"); idx != -1 {
            writeError(w, http.StatusNotFound, "invalid path")
            return
        }
        h.createComment(w, r, storyID)
        return
    case strings.HasSuffix(id, "/execute"):
        storyID := strings.TrimSuffix(id, "/execute")
        if idx := strings.Index(storyID, "/"); idx != -1 {
            writeError(w, http.StatusNotFound, "invalid path")
            return
        }
        h.executeStory(w, r, storyID)
        return
    case strings.HasSuffix(id, "/events"):
        storyID := strings.TrimSuffix(id, "/events")
        if idx := strings.Index(storyID, "/"); idx != -1 {
            writeError(w, http.StatusNotFound, "invalid path")
            return
        }
        h.streamEvents(w, r, storyID)
        return
    }

    switch r.Method {
    case http.MethodGet:
        h.getStory(w, r, id)
    case http.MethodOptions:
        w.WriteHeader(http.StatusNoContent)
    default:
        writeError(w, http.StatusMethodNotAllowed, "method not allowed")
    }
}

func (h handler) listStories(w http.ResponseWriter, r *http.Request) {
    filter := storypkg.Filter{
        Owner: r.URL.Query().Get("owner"),
        Tag:   r.URL.Query().Get("tag"),
        Query: r.URL.Query().Get("q"),
    }
    stories, err := h.stories.ListStories(r.Context(), filter)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, stories)
}

func (h handler) createStory(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        Title       string                   `json:"title"`
        Description string                   `json:"description"`
        Owners      []string                 `json:"owners"`
        Visibility  storypkg.Visibility      `json:"visibility"`
        Tags        []string                 `json:"tags"`
        Blocks      []storypkg.BlockInput    `json:"blocks"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        writeError(w, http.StatusBadRequest, "invalid json payload")
        return
    }
    created, err := h.stories.CreateStory(r.Context(), storypkg.CreateStoryInput{
        Title:       payload.Title,
        Description: payload.Description,
        Owners:      payload.Owners,
        Visibility:  payload.Visibility,
        Tags:        payload.Tags,
        Blocks:      payload.Blocks,
    })
    if err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, created)
}

func (h handler) getStory(w http.ResponseWriter, r *http.Request, id string) {
    story, err := h.stories.GetStory(r.Context(), id)
    if err != nil {
        if err == storypkg.ErrNotFound {
            writeError(w, http.StatusNotFound, "story not found")
            return
        }
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, story)
}

func (h handler) appendBlock(w http.ResponseWriter, r *http.Request, id string) {
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }
    if r.Method != http.MethodPost {
        writeError(w, http.StatusMethodNotAllowed, "method not allowed")
        return
    }
    var payload storypkg.BlockInput
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        writeError(w, http.StatusBadRequest, "invalid json payload")
        return
    }
    updated, err := h.stories.AppendBlock(r.Context(), id, payload)
    if err != nil {
        if err == storypkg.ErrNotFound {
            writeError(w, http.StatusNotFound, "story not found")
            return
        }
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, updated)
}

func (h handler) createComment(w http.ResponseWriter, r *http.Request, id string) {
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }
    if r.Method != http.MethodPost {
        writeError(w, http.StatusMethodNotAllowed, "method not allowed")
        return
    }
    var payload struct {
        Author  string `json:"author"`
        Body    string `json:"body"`
        BlockID string `json:"blockId"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        writeError(w, http.StatusBadRequest, "invalid json payload")
        return
    }
    updated, err := h.stories.RecordComment(r.Context(), id, storypkg.CommentInput{
        Author:  payload.Author,
        Body:    payload.Body,
        BlockID: payload.BlockID,
    })
    if err != nil {
        if err == storypkg.ErrNotFound {
            writeError(w, http.StatusNotFound, "story not found")
            return
        }
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, updated)
}

func (h handler) executeStory(w http.ResponseWriter, r *http.Request, id string) {
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }
    if r.Method != http.MethodPost {
        writeError(w, http.StatusMethodNotAllowed, "method not allowed")
        return
    }
    var payload struct {
        Actor string `json:"actor"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        writeError(w, http.StatusBadRequest, "invalid json payload")
        return
    }
    result, err := h.stories.ExecuteStory(r.Context(), id, payload.Actor)
    if err != nil {
        if err == storypkg.ErrNotFound {
            writeError(w, http.StatusNotFound, "story not found")
            return
        }
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    writeJSON(w, http.StatusOK, result)
}

func (h handler) streamEvents(w http.ResponseWriter, r *http.Request, id string) {
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }
    if r.Method != http.MethodGet {
        writeError(w, http.StatusMethodNotAllowed, "method not allowed")
        return
    }
    flusher, ok := w.(http.Flusher)
    if !ok {
        writeError(w, http.StatusInternalServerError, "streaming unsupported")
        return
    }
    ctx, cancel := context.WithCancel(r.Context())
    defer cancel()
    ch, unsubscribe := h.hub.Subscribe(id)
    defer unsubscribe()

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    notify := r.Context().Done()
    for {
        select {
        case <-ctx.Done():
            return
        case <-notify:
            return
        case event := <-ch:
            payload, err := event.Marshal()
            if err != nil {
                log.Printf("sse marshal error: %v", err)
                continue
            }
            if _, err := w.Write([]byte("event: " + event.Type + "\n")); err != nil {
                return
            }
            if _, err := w.Write([]byte("data: " + string(payload) + "\n\n")); err != nil {
                return
            }
            flusher.Flush()
        }
    }
}

func withLogging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
    })
}

func withCORS(allowed []string, next http.Handler) http.Handler {
    allowedOrigins := make(map[string]struct{}, len(allowed))
    for _, origin := range allowed {
        allowedOrigins[origin] = struct{}{}
    }
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        if origin != "" {
            if len(allowedOrigins) == 0 {
                w.Header().Set("Access-Control-Allow-Origin", origin)
            } else if _, ok := allowedOrigins[origin]; ok {
                w.Header().Set("Access-Control-Allow-Origin", origin)
            }
            w.Header().Set("Vary", "Origin")
        }
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}
