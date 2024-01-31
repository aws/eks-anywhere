---
title: "Verify Cluster Images"
linkTitle: "Verify Cluster Images"
weight: 90
date: 2017-01-05
description: >
  How to verify cluster images 
---

### Verify Cluster Images

Starting from v0.19.0 release, all the images used in cluster operations are signed using AWS Signer and Notation CLI. Anyone can verify signatures associated with the container images EKS Anywhere uses. Signatures are valid for two years. To verify container images, one would have to perform the following steps:

1. Install and configure the latest version of the AWS CLI. For more information, see [Installing or updating the latest version of the AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) in the AWS Command Line Interface User Guide.

2. Download the container-signing tools from [this page](https://docs.aws.amazon.com/signer/latest/developerguide/image-signing-prerequisites.html) by following step 3.

3. Make sure notation CLI is configured along with the AWS Signer plugin and AWS CLI. Create a JSON file with the following trust policy. 

   ```bash
   {
      "version":"1.0",
      "trustPolicies":[
         {
            "name":"aws-signer-tp",
            "registryScopes":[
               "*"
            ],
            "signatureVerification":{
               "level":"strict"
            },
            "trustStores":[
               "signingAuthority:aws-signer-ts"
            ],
            "trustedIdentities":[
               "arn:aws:signer:us-west-2:857151390494:/signing-profiles/notationimageSigningProfileECR_rGorpoAE4o0o"
            ]
         }
      ]
   }
   ```

4. Import above trust policy using:
   ```bash
   notation policy import <json-file>
   ```

5. Get the bundle of the version for which you want to verify an image:
   ```bash
   export EKSA_RELEASE_VERSION=v0.19.0
   BUNDLE_MANIFEST_URL=$(curl -sL https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml | yq ".spec.releases[] | select(.version==\"$EKSA_RELEASE_VERSION\").bundleManifestUrl")
   ```

6. Get the image you want to verify by downloading BUNDLE_MANIFEST_URL file.
7. Run the following command to verify the image signature:
   ```bash
   notation verify <image-uri>@<sha256-image-digest>
   ```



