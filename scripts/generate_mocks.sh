#!/bin/bash
# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

cmds=(
    '${GOPATH}/bin/mockgen -destination=pkg/providers/mocks/providers.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers" Provider,DatacenterConfig,MachineConfig'
	'${GOPATH}/bin/mockgen -destination=pkg/executables/mocks/executables.go -package=mocks "github.com/aws/eks-anywhere/pkg/executables" Executable'
	'${GOPATH}/bin/mockgen -destination=pkg/providers/docker/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/docker" ProviderClient,ProviderKubectlClient'
	'${GOPATH}/bin/mockgen -destination=pkg/providers/tinkerbell/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/tinkerbell" ProviderKubectlClient,ProviderTinkClient,ProviderPbnjClient'
	'${GOPATH}/bin/mockgen -destination=pkg/providers/cloudstack/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/cloudstack" ProviderCmkClient,ProviderKubectlClient'
	'${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/providers/vsphere" ProviderGovcClient,ProviderKubectlClient,ClusterResourceSetManager'
	'${GOPATH}/bin/mockgen -destination=pkg/filewriter/mocks/filewriter.go -package=mocks "github.com/aws/eks-anywhere/pkg/filewriter" FileWriter'
	'${GOPATH}/bin/mockgen -destination=pkg/clustermanager/mocks/client_and_networking.go -package=mocks "github.com/aws/eks-anywhere/pkg/clustermanager" ClusterClient,Networking,AwsIamAuth'
	'${GOPATH}/bin/mockgen -destination=pkg/addonmanager/addonclients/mocks/fluxaddonclient.go -package=mocks "github.com/aws/eks-anywhere/pkg/addonmanager/addonclients" Flux'
	'${GOPATH}/bin/mockgen -destination=pkg/task/mocks/task.go -package=mocks "github.com/aws/eks-anywhere/pkg/task" Task'
	'${GOPATH}/bin/mockgen -destination=pkg/bootstrapper/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/bootstrapper" ClusterClient'
	'${GOPATH}/bin/mockgen -destination=pkg/cluster/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/cluster" ClusterClient'
	'${GOPATH}/bin/mockgen -destination=pkg/workflows/interfaces/mocks/clients.go -package=mocks "github.com/aws/eks-anywhere/pkg/workflows/interfaces" Bootstrapper,ClusterManager,AddonManager,Validator,CAPIManager'
	'${GOPATH}/bin/mockgen -destination=pkg/git/providers/github/mocks/github.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/providers/github" GitProviderClient,GithubProviderClient'
	'${GOPATH}/bin/mockgen -destination=pkg/git/mocks/git.go -package=mocks "github.com/aws/eks-anywhere/pkg/git" Provider'
	'${GOPATH}/bin/mockgen -destination=pkg/git/gogithub/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/gogithub" Client'
	'${GOPATH}/bin/mockgen -destination=pkg/git/gogit/mocks/client.go -package=mocks "github.com/aws/eks-anywhere/pkg/git/gogit" GoGitClient'
	'${GOPATH}/bin/mockgen -destination=pkg/validations/mocks/docker.go -package=mocks "github.com/aws/eks-anywhere/pkg/validations" DockerExecutable'
	'${GOPATH}/bin/mockgen -destination=controllers/controllers/resource/mocks/resource.go -package=mocks "github.com/aws/eks-anywhere/controllers/controllers/resource" ResourceFetcher,ResourceUpdater'
	'${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/internal/templates/mocks/govc.go -package=mocks -source "pkg/providers/vsphere/internal/templates/factory.go" GovcClient'
	'${GOPATH}/bin/mockgen -destination=pkg/providers/vsphere/internal/tags/mocks/govc.go -package=mocks -source "pkg/providers/vsphere/internal/tags/factory.go" GovcClient'
	'${GOPATH}/bin/mockgen -destination=pkg/validations/mocks/kubectl.go -package=mocks -source "pkg/validations/kubectl.go" KubectlClient'
	'${GOPATH}/bin/mockgen -destination=pkg/validations/mocks/tls.go -package=mocks -source "pkg/validations/tls.go" TlsValidator'
	'${GOPATH}/bin/mockgen -destination=pkg/diagnostics/interfaces/mocks/diagnostics.go -package=mocks -source "pkg/diagnostics/interfaces.go" DiagnosticBundle,AnalyzerFactory,CollectorFactory,BundleClient'
	'${GOPATH}/bin/mockgen -destination=pkg/clusterapi/mocks/capiclient.go -package=mocks -source "pkg/clusterapi/manager.go" CAPIClient,KubectlClient'
	'${GOPATH}/bin/mockgen -destination=pkg/clusterapi/mocks/client.go -package=mocks -source "pkg/clusterapi/resourceset_manager.go" Client'
	'${GOPATH}/bin/mockgen -destination=pkg/crypto/mocks/crypto.go -package=mocks -source "pkg/crypto/certificategen.go" CertificateGenerator'
	'${GOPATH}/bin/mockgen -destination=pkg/networking/cilium/mocks/clients.go -package=mocks -source "pkg/networking/cilium/client.go"'
	'${GOPATH}/bin/mockgen -destination=pkg/networking/cilium/mocks/helm.go -package=mocks -source "pkg/networking/cilium/templater.go"'
	'${GOPATH}/bin/mockgen -destination=pkg/networking/cilium/mocks/upgrader.go -package=mocks -source "pkg/networking/cilium/upgrader.go"'
	'${GOPATH}/bin/mockgen -destination=pkg/networking/kindnetd/mocks/client.go -package=mocks -source "pkg/networking/kindnetd/upgrader.go"'
	'${GOPATH}/bin/mockgen -destination=pkg/networking/cilium/mocks/cilium.go -package=mocks -source "pkg/networking/cilium/cilium.go"'
	'${GOPATH}/bin/mockgen -destination=pkg/networkutils/mocks/client.go -package=mocks -source "pkg/networkutils/netclient.go" NetClient'
)

# Toggle concurent mock generation.
CONCURRENT=true

# Create a directory we can store all output in to be collected at the end in-case something goes wrong.
temp_dir="$PWD/mock_generation_tmp"
mkdir -p $temp_dir

if [ ${CONCURRENT} == true ]; then
	echo "Executing ${#cmds[@]} mock generation commands. This may take a while ..."
fi

# Iterate over all commands launching them as background jobs and writing output to separate files.
output_collection=()
for cmd in "${cmds[@]}"; do
	stdout="${temp_dir}/$(echo $cmd | sha256sum | cut -d" " -f1)"
	if [ ${CONCURRENT} == true ]; then
		eval "${cmd}" > "${stdout}" &
	else
		echo "${cmd}"
		eval "${cmd}" > "${stdout}"
	fi
	output_collection+=("${stdout}")
done

# Wait for all background jobs launched by this shell to finish.
wait
	
# Collect output from background processes incase something went wrong.
for output_file in "${output_collection[@]}"; do
	cat "${output_file}"
done

rm -rf $temp_dir
