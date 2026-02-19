package main

import (
	"github.com/templatr/templatr-setup/cmd"
)

// Version info — injected by GoReleaser via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	cmd.SetWebAssets(WebAssets)
	cmd.Execute()
}
