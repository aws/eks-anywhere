package workflow

import "context"

// targetCluster is used to store and retrieve a target cluster kubeconfig.
var targetCluster contextKey = "targetCluster"

// WithTargetCluster returns a context based on ctx containing the target cluster kubeconfig.
func WithTargetCluster(ctx context.Context, kubeconfig string) context.Context {
	return context.WithValue(ctx, targetCluster, kubeconfig)
}

// TargetCluster retrieves the target cluster configured in ctx or returns an empty string.
func TargetCluster(ctx context.Context) string {
	return ctx.Value(targetCluster).(string)
}

// contextKey is used to create collisionless context keys.
type contextKey string

func (c contextKey) String() string {
	return "workflow context value " + string(c)
}
