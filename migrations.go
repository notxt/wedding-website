package wedding

import "embed"

//go:embed all:migrations
var Migrations embed.FS
