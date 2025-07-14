package version

import (
	"fmt"
	"runtime"
)

// Version information
var (
	Version   = "1.0.0"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// Info contains version information
type Info struct {
	Version   string
	GitCommit string
	BuildTime string
	GoVersion string
	Platform  string
}

// GetVersion returns version information
func GetVersion() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("rhesis %s (%s) built with %s on %s at %s",
		i.Version, i.GitCommit, i.GoVersion, i.Platform, i.BuildTime)
}
