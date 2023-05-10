package version

var gitVersion string

type Info struct {
	GitVersion string
}

var Get = func() Info {
	return Info{
		GitVersion: gitVersion,
	}
}
