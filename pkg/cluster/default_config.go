//go:build !spec_embed_config
// +build !spec_embed_config

package cluster

import "embed"

var configFS embed.FS
