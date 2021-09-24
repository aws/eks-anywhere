//go:build spec_embed_config
// +build spec_embed_config

package cluster

import "embed"

//go:embed config
var configFS embed.FS
