package update

type Channel int

const (
	// GitHub's "latest" excludes prereleases.
	Stable Channel = iota

	Dev
)

const releases = "https://github.com/ygelfand/echolocal/releases"

func (c Channel) Label() string {
	switch c {
	case Dev:
		return "dev"
	}
	return "stable"
}

// The GitHub REST API allows 60 unauthenticated requests an hour; asset downloads do not count.
func (c Channel) URL() string {
	switch c {
	case Dev:
		return releases + "/download/dev/manifest.json"
	}
	return releases + "/latest/download/manifest.json"
}

func Channels() []Channel { return []Channel{Stable, Dev} }
