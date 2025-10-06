package server

import (
    "net/http"
    "time"

    realtimepkg "github.com/example/multistory/internal/realtime"
    storypkg "github.com/example/multistory/internal/story"
)

// New constructs an *http.Server configured with sensible defaults ready to serve requests.
func New(cfg Config, svc storypkg.Service, hub *realtimepkg.Hub) *http.Server {
    handler := newRouter(cfg, svc, hub)
    return &http.Server{
        Addr:              cfg.httpAddr(),
        Handler:           handler,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       10 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       60 * time.Second,
    }
}
