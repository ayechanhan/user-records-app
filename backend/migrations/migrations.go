// Package migrations embeds the SQL migration files so the server binary
// can apply them at startup without depending on the filesystem layout.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
