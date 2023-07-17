# E2E tests

## How to run tests
```bash
make build-all-test-binaries
./bin/e2e.test -test.v -test.run [test name regex]
```
or
```bash
# Create a .env file at the root of the repo with all the required envs vars listed below
#
# The makefile will include the .env file and export all the vars to the environment for you
#
# By default the local-e2e target will run TestDockerKubernetes125SimpleFlow. You can either 
#   override LOCAL_E2E_TESTS in your .env file or pass it on the cli every time (i.e LOCAL_E2E_TESTS=TestDockerKubernetes125SimpleFlow)
make local-e2e
```
or
```bash
go test -tags "e2e all_providers" -run [test name regex]
```

### Configuring KubeVersion and OS version for the tests

Each provider file in the [E2E framework folder](../framework) implements the `WithKubeVersionAndOS()` that takes in a Kubernetes version and an OS value as arguments, which instructs the tests to use an template/image corresponding to that specific Kubernetes version and OS. The OS passed as input can be any of the ones in the [OS versions file](../framework/os_versions.go).

For example, here is how you would write vSphere Kubernetes 1.27 simple flow test that uses an Ubuntu 22.04 template.
```Go
func TestVSphereKubernetes127Ubuntu2204SimpleFlow(t *testing.T) {
	provider := framework.NewVSphere(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
	).WithClusterConfig(
		provider.WithKubeVersionAndOS(v1alpha1.Kube127, framework.Ubuntu2204, nil),
	)
	runSimpleFlowWithoutClusterConfigGeneration(test)
}
```

Currently, Ubuntu is the only OS that we build and test multiple versions for. In the future, when we introduce multiple version support for another OS, you can define it in the [OS versions file](../framework/os_versions.go) and use it in the tests like above.

### Using bundle overrides
In order to use bundle overrides, take your bundle overrides yaml file and move it to `ROOT_DIR/bin/local-bundle-release.yaml`.
You will also need to set the environment variable `T_BUNDLES_OVERRIDE=true`

### Cleaning up VM's after a test run
In order to clean up VM's after a test runs automatically, set `T_CLEANUP_VMS=true`

## VSphere tests requisites
The following env variables need to be set:

```sh
T_VSPHERE_DATACENTER
T_VSPHERE_DATASTORE
T_VSPHERE_FOLDER
T_VSPHERE_NETWORK
T_VSPHERE_RESOURCE_POOL
T_VSPHERE_SERVER
T_VSPHERE_SSH_AUTHORIZED_KEY
T_VSPHERE_TEMPLATE_UBUNTU_1_21
T_VSPHERE_TEMPLATE_UBUNTU_1_22
T_VSPHERE_TEMPLATE_UBUNTU_1_23
T_VSPHERE_TEMPLATE_UBUNTU_1_24
T_VSPHERE_TEMPLATE_UBUNTU_1_25
T_VSPHERE_TLS_INSECURE
T_VSPHERE_TLS_THUMBPRINT
VSPHERE_USERNAME
VSPHERE_PASSWORD
T_VSPHERE_CIDR
GOVC_URL
T_VSPHERE_CLUSTER_IP_POOL # comma-separated list of CP ip addresses
EKSA_AWS_ACCESS_KEY_ID # For Packages tests
EKSA_AWS_SECRET_ACCESS_KEY # For Packages tests
EKSA_AWS_REGION:# For Packages tests
```

### Tests upgrading from old release

If you are running tests that create clusters using an old release and upgrade to the new one (eg. `TestVSphereKubernetes123BottlerocketUpgradeFromLatestMinorRelease`), you will need extra variables for the templates.

The format is: `T_VSPHERE_TEMPLATE_{OS}_{EKS-D VERSION}`. For example, for Ubuntu, kubernetes 1.23 and release v0.11.0, which uses eks-d release `kubernetes-1-23-eks-4`: `T_VSPHERE_TEMPLATE_UBUNTU_KUBERNETES_1_23_EKS_4`.

# Nutanix tests requisites
 The following env variables need to be set:

 ```sh
 EKSA_NUTANIX_USERNAME
 EKSA_NUTANIX_PASSWORD

 T_NUTANIX_ENDPOINT
 T_NUTANIX_PORT
 T_NUTANIX_INSECURE
 T_NUTANIX_ADDITIONAL_TRUST_BUNDLE # This should be set to the base64 encoded CA cert used for Nutanix Prism Central
 T_NUTANIX_MACHINE_BOOT_TYPE
 T_NUTANIX_MACHINE_MEMORY_SIZE
 T_NUTANIX_SYSTEMDISK_SIZE
 T_NUTANIX_MACHINE_VCPU_PER_SOCKET
 T_NUTANIX_MACHINE_VCPU_SOCKET
 T_NUTANIX_PRISM_ELEMENT_CLUSTER_NAME
 T_NUTANIX_SUBNET_NAME
 T_NUTANIX_SSH_AUTHORIZED_KEY
 T_NUTANIX_CONTROL_PLANE_ENDPOINT_IP
 T_NUTANIX_POD_CIDR
 T_NUTANIX_SERVICE_CIDR
 T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_22
 T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_23
 T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_24
 T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_25
 T_NUTANIX_TEMPLATE_NAME_UBUNTU_1_26
 ```
## Tinkerbell tests requisites
The following env variables need to be set:

```sh
T_TINKERBELL_IP
T_TINKERBELL_IMAGE_UBUNTU_1_21
T_TINKERBELL_IMAGE_UBUNTU_1_22
T_TINKERBELL_IMAGE_UBUNTU_1_23
T_TINKERBELL_CP_NETWORK_CIDR
T_TINKERBELL_INVENTORY_CSV # path to hardware-inventory.csv file
T_TINKERBELL_SSH_AUTHORIZED_KEY # ssh public key for connectioning to machines
```
### Tinkerbell hardware-inventory.csv example
```csv
guid,ip_address,gateway,nameservers,netmask,mac,hostname,vendor,bmc_ip,bmc_username,bmc_password,labels,disk
bb341bc6-546f-4b38-s584-bb4f0e5f8934,10.24.32.110,10.24.32.1,8.8.8.8,255.255.255.0,3c:ec:ef:6e:a4:82,eksa-node01,supermicro,10.24.32.10,admin,password,type=cp,/dev/sda
cc5619b8-a894-4db0-bf1a-fd04d5964d54,10.24.32.111,10.24.32.1,8.8.8.8,,255.255.255.0,3c:ec:ef:6e:a5:7c,eksa-node02,supermicro,10.24.32.11,admin,password,type=worker,/dev/sda
```

## CloudStack tests requisites

The following env variables need to be set:
```
T_CLOUDSTACK_DOMAIN
T_CLOUDSTACK_ZONE
T_CLOUDSTACK_ZONE_2
T_CLOUDSTACK_CREDENTIALS
T_CLOUDSTACK_CREDENTIALS_2
T_CLOUDSTACK_ACCOUNT
T_CLOUDSTACK_NETWORK
T_CLOUDSTACK_NETWORK_2
T_CLOUDSTACK_MANAGEMENT_SERVER
T_CLOUDSTACK_MANAGEMENT_SERVER_2
T_CLOUDSTACK_SSH_AUTHORIZED_KEY
T_CLOUDSTACK_TEMPLATE_REDHAT_1_22
T_CLOUDSTACK_TEMPLATE_REDHAT_1_23
T_CLOUDSTACK_COMPUTE_OFFERING_LARGE
T_CLOUDSTACK_COMPUTE_OFFERING_LARGER
T_CLOUDSTACK_POD_CIDR
T_CLOUDSTACK_SERVICE_CIDR
T_CLOUDSTACK_CLUSTER_IP_POOL # Comma separated list of control plane IP's

EKSA_CLOUDSTACK_B64ENCODED_SECRET
```

## Snow tests requisites

The following env variables need to be set (required):

```sh
T_SNOW_CONTROL_PLANE_CIDR
T_SNOW_POD_CIDR
T_SNOW_DEVICES
EKSA_AWS_CREDENTIALS_FILE
EKSA_AWS_CA_BUNDLES_FILE
```

> **NOTE**: `T_SNOW_DEVICES` should be a comma-separated list of device IPs.

Optional env variables for specific tests:

```sh
T_SNOW_AMIID_UBUNTU_1_21
T_SNOW_AMIID_UBUNTU_1_22
T_SNOW_AMIID_UBUNTU_1_23
T_SNOW_AMIID_UBUNTU_1_24
T_SNOW_AMIID_UBUNTU_1_25
T_SNOW_AMIID_BOTTLEROCKET_1_21
T_SNOW_AMIID_BOTTLEROCKET_1_22
T_SNOW_AMIID_BOTTLEROCKET_1_23
T_SNOW_AMIID_BOTTLEROCKET_1_24
T_SNOW_AMIID_BOTTLEROCKET_1_25
T_SNOW_IPPOOL_IPSTART
T_SNOW_IPPOOL_IPEND
T_SNOW_IPPOOL_GATEWAY
T_SNOW_IPPOOL_SUBNET
```

> **NOTE**: 
  * Env vars with prefix `T_SNOW_AMIID_` are the optional AMI ids based on OS family and K8s version that will be used to create node instances. If not specified or left empty, CAPAS will use its AMI lookup logic and try to find a valid AMI in the device based on the OS family and K8s version.
  * Env vars with prefix `T_SNOW_IPPOOL_` are required when running Snow test with static ip. The values will be used to generate the `SnowIPPool` object.

## OIDC tests requisites
The following env variables need to be set:

```sh
T_OIDC_ISSUER_URL
T_OIDC_CLIENT_ID
T_OIDC_KID
T_OIDC_KEY_FILE # private rsa key to sign jwt tokens
```

## GitOps tests requisites
The following env variables need to be set:

```sh
T_GIT_REPOSITORY
T_GITHUB_USER
GITHUB_TOKEN
```
The [oidc](https://github.com/aws/eks-anywhere/blob/main/internal/pkg/oidc/server.go) and [e2e](https://github.com/aws/eks-anywhere/blob/main/internal/test/e2e/oidc.go) packages can be used to create a minimal compliant OIDC server in S3 

## Proxy test requisites
The following env variables need to be set:

For VSphere proxy:
```sh
T_HTTP_PROXY_VSPHERE
T_HTTPS_PROXY_VSPHERE
T_NO_PROXY_VSPHERE
```

For CloudStack proxy:
```sh
T_HTTP_PROXY_CLOUDSTACK
T_HTTPS_PROXY_CLOUDSTACK
T_NO_PROXY_CLOUDSTACK
```

## Registry test requisites
The following env variables need to be set:

```sh
T_REGISTRY_MIRROR_ENDPOINT
T_REGISTRY_MIRROR_PORT
T_REGISTRY_MIRROR_CA_CERT
T_REGISTRY_MIRROR_USERNAME
T_REGISTRY_MIRROR_PASSWORD
```

## Adding new tests
When adding new tests to run in our postsubmit environment we need to bump up the total number of EC2s we create for the tests.

The value is controlled by the `INTEGRATION_TEST_MAX_EC2_COUNT` env variable in each provider's buildspec in the [integration test buildspecs folder](https://github.com/aws/eks-anywhere/blob/main/cmd/integration_test/build/buildspecs).

```
env:
  variables:
    INTEGRATION_TEST_MAX_EC2_COUNT: <COUNT>
```
