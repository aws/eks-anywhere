# E2E tests

## How to run tests
```sh
make e2e
./bin/e2e.test -test.v -test.run [test name regex]
```
or
```sh
# Create a .env file at the root of the repo with all the required envs vars listed below
#
# The makefile will include the .env file and export all the vars to the environment for you
#
# By default the local-e2e target will run TestDockerKubernetes121SimpleFlow. You can either 
#   override LOCAL_E2E_TESTS in your .env file or pass it on the cli every time (i.e LOCAL_E2E_TESTS=TestDockerKubernetes121SimpleFlow)
make local-e2e
```
or
```sh
go test -tags e2e -run [test name regex]
```

# VSphere tests requisites
The following env variables need to be set:

```sh
T_VSPHERE_DATACENTER
T_VSPHERE_DATASTORE
T_VSPHERE_FOLDER
T_VSPHERE_NETWORK
T_VSPHERE_RESOURCE_POOL
T_VSPHERE_SERVER
T_VSPHERE_SSH_AUTHORIZED_KEY
T_VSPHERE_TEMPLATE_UBUNTU_1_19
T_VSPHERE_TEMPLATE_UBUNTU_1_20
T_VSPHERE_TEMPLATE_UBUNTU_1_21
T_VSPHERE_TEMPLATE_UBUNTU_1_22
T_VSPHERE_TLS_INSECURE
T_VSPHERE_TLS_THUMBPRINT
VSPHERE_USERNAME
VSPHERE_PASSWORD
T_VSPHERE_CIDR
GOVC_URL
T_CLUSTER_IP_POOL # comma-separated list of CP ip addresses
```

# OIDC tests requisites
The following env variables need to be set:

```sh
T_OIDC_ISSUER_URL
T_OIDC_CLIENT_ID
T_OIDC_KID
T_OIDC_KEY_FILE # private rsa key to sign jwt tokens
```

# GitOps tests requisites
The following env variables need to be set:

```sh
T_GIT_REPOSITORY
T_GITHUB_USER
GITHUB_TOKEN
```
The [oidc](https://github.com/aws/eks-anywhere/blob/main/internal/pkg/oidc/server.go) and [e2e](https://github.com/aws/eks-anywhere/blob/main/internal/test/e2e/oidc.go) packages can be used to create a minimal compliant OIDC server in S3 

# Proxy test requisites
The following env variables need to be set:

```sh
T_HTTP_PROXY
T_HTTPS_PROXY
T_NO_PROXY
```

# Registry test requisites
The following env variables need to be set:

```sh
T_REGISTRY_MIRROR_ENDPOINT
T_REGISTRY_MIRROR_PORT
T_REGISTRY_MIRROR_CA_CERT
T_REGISTRY_MIRROR_USERNAME
T_REGISTRY_MIRROR_PASSWORD
```

# Adding new tests
When adding new tests to run in our postsubmit environment we need to bump up the total number of EC2s we create for the tests.

The value is controlled by the `INTEGRATION_TEST_MAX_EC2_COUNT` env variable in the [test-eks-a-cli.yaml](https://github.com/aws/eks-anywhere/blob/main/cmd/integration_test/build/buildspecs/test-eks-a-cli.yml) buildspec.

```
env:
  variables:
    INTEGRATION_TEST_MAX_EC2_COUNT: <COUNT>
```