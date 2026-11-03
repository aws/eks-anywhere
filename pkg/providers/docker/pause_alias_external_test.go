package docker_test

const testPauseImageAliasCommand = `set -eu
images=$(kubeadm config images list --kubernetes-version 'v1.23.12-eks-1-23-6' --image-repository 'public.ecr.aws/eks-distro/kubernetes')
pause_image=$(printf '%s\n' "$images" | grep '/pause:')
ctr --namespace k8s.io images tag --force 'registry.example.test/kubernetes/pause:v1.23.12-example' "$pause_image"
`
