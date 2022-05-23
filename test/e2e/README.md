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

### Using bundle overrides
In order to use bundle overrides, take your bundle overrides yaml file and move it to `ROOT_DIR/bin/local-bundle-release.yaml`.
You will also need to set the environment variable `T_BUNDLES_OVERRIDE=true`

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
T_VSPHERE_TEMPLATE_UBUNTU_1_23
T_VSPHERE_TLS_INSECURE
T_VSPHERE_TLS_THUMBPRINT
VSPHERE_USERNAME
VSPHERE_PASSWORD
T_VSPHERE_CIDR
GOVC_URL
T_VSPHERE_CLUSTER_IP_POOL # comma-separated list of CP ip addresses
```

# Tinkerbell tests requisites
The following env variables need to be set:

```sh
T_TINKERBELL_IP
T_TINKERBELL_CERT_URL=http://${T_TINKERBELL_IP}:42114/cert
T_TINKERBELL_HEGEL_URL=http://${T_TINKERBELL_IP}:50061
T_TINKERBELL_GRPC_AUTHORITY=${T_TINKERBELL_IP}:42113
T_TINKERBELL_PBNJ_GRPC_AUTHORITY=${T_TINKERBELL_IP}:50051
T_TINKERBELL_IMAGE_UBUNTU_1_20
T_TINKERBELL_IMAGE_UBUNTU_1_21
T_TINKERBELL_NETWORK_CIDR
T_TINKERBELL_INVENTORY_CSV # path to hardware-inventory.csv file
T_TINKERBELL_SSH_AUTHORIZED_KEY # ssh public key for connectioning to machines
```
## Tinkerbell hardware-inventory.csv example
```csv
guid,ip_address,gateway,nameservers,netmask,mac,hostname,vendor,bmc_ip,bmc_username,bmc_password
bb341bc6-546f-4b38-s584-bb4f0e5f8934,10.24.32.110,10.24.32.1,8.8.8.8,255.255.255.0,3c:ec:ef:6e:a4:82,eksa-node01,supermicro,10.24.32.10,admin,password
cc5619b8-a894-4db0-bf1a-fd04d5964d54,10.24.32.111,10.24.32.1,8.8.8.8,,255.255.255.0,3c:ec:ef:6e:a5:7c,eksa-node02,supermicro,10.24.32.11,admin,password
```

# CloudStack tests requisites

The following env variables need to be set:
```
T_CLOUDSTACK_DOMAIN
T_CLOUDSTACK_ZONE
T_CLOUDSTACK_ACCOUNT
T_CLOUDSTACK_NETWORK
T_CLOUDSTACK_MANAGEMENT_SERVER
T_CLOUDSTACK_SSH_AUTHORIZED_KEY
T_CLOUDSTACK_TEMPLATE_REDHAT_1_20
T_CLOUDSTACK_TEMPLATE_REDHAT_1_21
T_CLOUDSTACK_COMPUTE_OFFERING_LARGE
T_CLOUDSTACK_COMPUTE_OFFERING_LARGER
T_CLOUDSTACK_POD_CIDR
T_CLOUDSTACK_SERVICE_CIDR
T_CLOUDSTACK_CLUSTER_IP_POOL # Comma separated list of control plane IP's

EKSA_CLOUDSTACK_B64ENCODED_SECRET
CLOUDSTACK_PROVIDER=true (while cloudstack provider is under development)
```

# Snow tests requisites
The following env variables need to be set:

```sh
T_SNOW_AMIID_UBUNTU_1_20
T_SNOW_AMIID_UBUNTU_1_21
T_SNOW_CONTROL_PLANE_CIDR
T_SNOW_POD_CIDR
EKSA_AWS_CREDENTIALS_FILE
EKSA_AWS_CA_BUNDLES_FILE
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