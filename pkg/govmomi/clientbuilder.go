package govmomi

import (
	"context"
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

type VSphereClient interface {
	Username() string
	GetPrivsOnEntity(ctx context.Context, path string, objType string, username string) ([]string, error)
}

type VMOMIFinderBuilder interface {
	Build(arg0 *vim25.Client, arg1 ...bool) VMOMIFinder
}

type VMOMISessionBuilder interface {
	Build(ctx context.Context, u *url.URL, insecure bool) (*govmomi.Client, error)
}

type VMOMIAuthorizationManagerBuilder interface {
	Build(c *vim25.Client) *object.AuthorizationManager
}

type vMOMIClientBuilder struct {
	vfb VMOMIFinderBuilder
	gcb VMOMISessionBuilder
	amb VMOMIAuthorizationManagerBuilder
}

func NewVMOMIClientBuilder() *vMOMIClientBuilder {
	return &vMOMIClientBuilder{vfb: &vMOMIFinderBuilder{}, gcb: &vMOMISessionBuilder{}, amb: &vMOMIAuthorizationManagerBuilder{}}
}

func NewVMOMIClientBuilderOverride(vfb VMOMIFinderBuilder, gcb VMOMISessionBuilder, amb VMOMIAuthorizationManagerBuilder) *vMOMIClientBuilder {
	return &vMOMIClientBuilder{vfb: vfb, gcb: gcb, amb: amb}
}

func (vcb *vMOMIClientBuilder) Build(ctx context.Context, host string, username string, password string, insecure bool, datacenter string) (VSphereClient, error) {
	u, err := soap.ParseURL(host)
	u.User = url.UserPassword(username, password)
	if err != nil {
		return nil, err
	}

	// start gvmc
	gvmc, err := vcb.gcb.Build(ctx, u, insecure)
	if err != nil {
		return nil, err
	}

	f := vcb.vfb.Build(gvmc.Client, true)

	dc, err := f.Datacenter(ctx, datacenter)
	if err != nil {
		return nil, err
	}

	f.SetDatacenter(dc)

	am := vcb.amb.Build(gvmc.Client)

	return &VMOMIClient{gvmc, f, username, am}, nil
}
