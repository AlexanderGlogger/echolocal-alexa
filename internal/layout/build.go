package layout

import "fmt"

// Set with -ldflags at build time.
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func VersionString() string {
	return fmt.Sprintf("%s (%s, %s)", Version, GitCommit, BuildDate)
}
