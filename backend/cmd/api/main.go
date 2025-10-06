package main

import (
    "context"
    "log"
    "net/http"
    "os/signal"
    "syscall"
    "time"

    "github.com/example/multistory/internal/executor"
    "github.com/example/multistory/internal/platform"
    "github.com/example/multistory/internal/realtime"
    "github.com/example/multistory/internal/server"
    "github.com/example/multistory/internal/story"
)

func main() {
    cfg := server.Config{
        Addr: platform.Env("PORT", "8080"),
        AllowedOrigins: []string{
            "http://localhost:3000",
            "http://localhost:8501",
        },
    }

    repo := story.NewMemoryRepository()
    hub := realtime.NewHub()
    runner := executor.NewStub()
    svc := story.NewService(repo, runner, hub)

    srv := server.New(cfg, svc, hub)

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    go func() {
        log.Printf("http server listening on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("http server error: %v", err)
        }
    }()

    <-ctx.Done()
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Printf("graceful shutdown error: %v", err)
    }

    log.Println("server stopped")
}
