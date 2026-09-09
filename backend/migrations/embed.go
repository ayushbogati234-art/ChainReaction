package migrations

import "embed"

// Files is embedded so migrations work in the minimal API/worker images.
//
//go:embed *.sql
var Files embed.FS
