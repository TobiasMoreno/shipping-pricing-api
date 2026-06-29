// Package migrations embeds the SQL migration files so they can be applied from
// the binary and from tests without depending on the working directory.
package migrations

import "embed"

// FS holds the embedded .sql migration files.
//
//go:embed *.sql
var FS embed.FS
