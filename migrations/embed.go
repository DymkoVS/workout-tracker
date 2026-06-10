// Package migrations bundles the SQL migration files into the binary so the
// runtime migration runner (internal/db.Migrate) can apply them without
// depending on the files being present on disk next to the server.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
