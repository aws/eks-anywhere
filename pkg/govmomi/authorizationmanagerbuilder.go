package govmomi

import (
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
)

type vMOMIAuthorizationManagerBuilder struct{}

func (*vMOMIAuthorizationManagerBuilder) Build(c *vim25.Client) *object.AuthorizationManager {
	return object.NewAuthorizationManager(c)
}
