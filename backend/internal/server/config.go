package server

// Config collects runtime settings for the HTTP server.
type Config struct {
    Addr           string
    AllowedOrigins []string
}

func (c Config) httpAddr() string {
    if c.Addr == "" {
        return ":8080"
    }
    if c.Addr[0] == ':' {
        return c.Addr
    }
    return ":" + c.Addr
}
