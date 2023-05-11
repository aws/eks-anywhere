package version

var gitVersion string

type Info struct {
	GitVersion string
}

func Get() Info {
	return Info{
		GitVersion: gitVersion,
	}
}
