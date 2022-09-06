package govmomi

import (
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25"
)

type vMOMIFinderBuilder struct{}

func NewVMOMIFinderBuilder() *vMOMIFinderBuilder {
	return &vMOMIFinderBuilder{}
}

func (*vMOMIFinderBuilder) Build(client *vim25.Client, all ...bool) VMOMIFinder {
	return find.NewFinder(client, all...)
}
