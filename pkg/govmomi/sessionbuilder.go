package govmomi

import (
	"context"
	"net/url"

	"github.com/vmware/govmomi"
)

type vMOMISessionBuilder struct{}

func NewvMOMISessionBuilder() *vMOMIClientBuilder {
	return &vMOMIClientBuilder{}
}

func (*vMOMISessionBuilder) Build(ctx context.Context, u *url.URL, insecure bool) (*govmomi.Client, error) {
	return govmomi.NewClient(ctx, u, insecure)
}
