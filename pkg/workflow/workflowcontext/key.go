package workflowcontext

// contextKey is used to create collisionless context keys.
type contextKey string

func (c contextKey) String() string {
	return string(c)
}
