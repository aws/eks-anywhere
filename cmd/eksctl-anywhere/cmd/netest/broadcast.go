package netest

// TestBroadcaster broadcasts names of tests as they are executed.
type TestBroadcaster interface {
	Broadcast(name string)
}
