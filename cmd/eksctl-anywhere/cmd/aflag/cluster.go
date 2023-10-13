package aflag

// ClusterConfig is the path to a cluster specification YAML.
var ClusterConfig = Flag[string]{
	Name:  "filename",
	Short: "f",
	Usage: "Path that contains a cluster configuration",
}

// BundleOverride is a path to a bundles manifest YAML that will be used in-place of the cluster
// specifications bundle override.
var BundleOverride = Flag[string]{
	Name:  "bundles-override",
	Usage: "A path to a custom bundles manifest",
}
