package platform

import "os"

// Env retrieves environment variable or returns default.
func Env(key, def string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return def
}
