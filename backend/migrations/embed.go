// Package migrations embute os arquivos SQL versionados no binário.
// Convenção: NNNN_descricao.up.sql (aplicado) e NNNN_descricao.down.sql (rollback manual).
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
