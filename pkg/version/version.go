package version

var gitVersion string

type Info struct {
	GitVersion string
}

// Get is a function that retrieves the version Info.
var Get = func() Info {
	return Info{
		GitVersion: gitVersion,
	}
}
