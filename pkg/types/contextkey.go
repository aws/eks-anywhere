package types

// EKSACliContextKey is defined to avoid conflict with other packages.
type EKSACliContextKey string

// InsecureRegistry can be used to bypass https registry certification check when push/pull images or artifacts.
var InsecureRegistry = EKSACliContextKey("insecure-registry")
