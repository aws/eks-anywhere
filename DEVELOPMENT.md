## Development

### Generate/Update Mocks

#### Prerequisites

1. We need mockgen installed, `go get github.com/golang/mock/mockgen`
2. Verify the binary was installed, `$GOPATH/bin/mockgen`

#### Generate/Update

To generate or update Mocks for an interface in a package, use the following 
command:

```bash
make mocks
```

If you are mocking a new interface for testing, be sure to update the Makefile 
target `mocks` to include the new interface.

```Makefile
.PHONY: mocks
mocks:
    mockgen -destination=pkg/<package path>/mocks/<package name>.go -package=mocks "github.com/aws/eks-anywhere/<package>" <Interface1,Interface2>
```
