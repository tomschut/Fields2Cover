// Copyright (C) 2026 Wageningen University — BSD-3-Clause
//
// Package static holds the compiled React SPA assets. The "dist"
// directory is populated by the Docker build-frontend stage before
// go build runs. In a clean checkout without a frontend build, the
// committed placeholder files in dist/ satisfy the embed directive.

package static

import "embed"

//go:embed dist
var FS embed.FS
