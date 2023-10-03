---
toc_hide: true
---
1. Download the EKS Anywhere artifacts that contain the list and locations of the EKS Anywhere dependencies. A compressed file `eks-anywhere-downloads.tar.gz` will be downloaded. You can use the `eksctl anywhere download artifacts --dry-run` command to see the list of artifacts it will download.
   ```bash
   eksctl anywhere download artifacts
   ```
   
1. Decompress the `eks-anywhere-downloads.tar.gz` file using the following command. This will create an `eks-anywhere-downloads` folder.
   ```bash
   tar -xvf eks-anywhere-downloads.tar.gz
   ```

1. Download the EKS Anywhere image dependencies to the Admin machine. This command may take several minutes (10+) to complete. To monitor the progress of the command, you can run with the `-v 6` command line argument, which will show details of the images that are being pulled. Docker must be running for the following command to succeed.
   ```bash
   eksctl anywhere download images -o images.tar
   ```

1. Set up a local registry mirror to host the downloaded EKS Anywhere images and configure your Admin machine with the certificates and authentication information if your registry requires it. For details, refer to the [Registry Mirror Configuration documentation.]({{< relref "../../getting-started/optional/registrymirror/#configure-local-registry-mirror" >}})

1. Import images to the local registry mirror using the following command. Set `REGISTRY_MIRROR_URL` to the url of the local registry mirror you created in the previous step. This command may take several minutes to complete. To monitor the progress of the command, you can run with the `-v 6` command line argument.  
   ```bash
   export REGISTRY_MIRROR_URL=<registryurl>
   ```
   ```bash
   eksctl anywhere import images -i images.tar -r ${REGISTRY_MIRROR_URL} \
      --bundles ./eks-anywhere-downloads/bundle-release.yaml
   ```

1. Optionally import curated packages to your registry mirror. The curated packages images are copied from Amazon ECR to your local registry mirror in a single step, as opposed to separate download and import steps. For post-cluster creation steps, reference the [Curated Packages documentation.]({{< relref "../../packages/prereq/#prepare-for-using-curated-packages-for-airgapped-environments" >}})
   
   <details>
      <summary>Expand for curated packages instructions</summary>

      If your EKS Anywhere cluster is running in an airgapped environment, and you set up a local registry mirror, you can copy curated packages from Amazon ECR to your local registry mirror with the following command.
      
      Set `$KUBEVERSION` to be equal to the `spec.kubernetesVersion` of your EKS Anywhere cluster specification.
      
      The `copy packages` command uses the credentials in your docker config file. So you must `docker login` to the source registries and the destination registry before running the command.
      
      ```bash
      eksctl anywhere copy packages \
        ${REGISTRY_MIRROR_URL}/curated-packages \
        --kube-version $KUBEVERSION \
        --src-chart-registry public.ecr.aws/eks-anywhere \
        --src-image-registry 783794618700.dkr.ecr.us-west-2.amazonaws.com
      ```
   </details>
