package bootstrap

import "fmt"

// ErrClusterExists indicates a bootstrap cluster with the same name already exists so could not be
// created.
type ErrClusterExists struct {
	ClusterName string
}

func (e ErrClusterExists) Error() string {
	return fmt.Sprintf("cluster %v already exists", e.ClusterName)
}

// ErrUnexpectedState indicates a bootstrap cluster that we tried to delete has unexpected
// state.
type ErrUnexpectedState struct {
	ClusterName string

	// (chrisdoherty4) Consider including a list of resources/state that contributed to this error
	//  as consumers may want to know and/or output more verbose logging.
}

func (e ErrUnexpectedState) Error() string {
	return fmt.Sprintf("cluster %v has state", e.ClusterName)
}
