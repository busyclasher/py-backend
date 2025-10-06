package id

import (
    "crypto/rand"
    "encoding/base32"
)

var encoding = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.NoPadding)

// New returns a URL-friendly random identifier suitable for demo data.
func New() string {
    buf := make([]byte, 10)
    if _, err := rand.Read(buf); err != nil {
        panic(err)
    }
    return encoding.EncodeToString(buf)
}
