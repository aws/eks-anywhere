//go:build !files_embed_fs
// +build !files_embed_fs

package files

import "embed"

// embedFS is used to conditionally embed files in the binary. Only one
// of embed_config.go and empty_embed.go will be used, following
// `files_embed_fs` build tag.
var embedFS embed.FS
