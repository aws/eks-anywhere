---
title: "Contributing to documentation"
weight: 20
description: >
  Guidelines for contributing to EKS Anywhere documentation
---
EKS Anywhere documentation uses the [Hugo](https://gohugo.io/categories/fundamentals) site generator and the [Docsy](https://www.docsy.dev/docs/) theme. To get started contributing:

* View the published EKS Anywhere [user documentation](https://anywhere.eks.amazonaws.com/docs/).
* Fork and clone the [eks-anywhere](https://github.com/aws/eks-anywhere) project.
* See [EKS Anywhere Documentation](https://github.com/aws/eks-anywhere/tree/main/docs#eks-anywhere-documentation) to set up your own docs test site.
* See the [General Guidelines](https://github.com/aws/eks-anywhere/blob/main/docs/content/en/docs/community/contributing.md) for contributing to the EKS Anywhere project
* Create EKS Anywhere documentation [Issues](https://github.com/aws/eks-anywhere/issues) and [Pull Requests](https://github.com/aws/eks-anywhere/pulls).

## Style issues

* **EKS Anywhere**: Always refer to EKS Anywhere as EKS Anywhere and *NOT* EKS-A or EKS-Anywhere.
* **Line breaks**: Put each sentence on its own line and don’t do a line break in the middle of a sentence. 
  We are using a modified [Semantic Line Breaking](https://sembr.org/) in that we are requiring a break at the end of every sentence, but not at commas or other semantic boundaries.
* **Headings**: Use sentence case in headings. So do “Cluster specification reference” and not “Cluster Specification Reference”
* **Cross references**: To cross reference to another doc in the EKS Anywhere docs set, use relref in the link so that Hugo will test it and fail the build for links not found. Also, use relative paths to point to other content in the docs set. Here is an example of a cross reference (code and results):
   ```
     See the [troubleshooting section]({ {< relref "../troubleshooting" >} } ) page.
   ```
     See the [troubleshooting section]({{< relref "../troubleshooting" >}}) page.

* **Notes, Warnings, etc.**: You can use this form for notes:

    <b><tt>\{\{% alert title="Note" color="primary" %\}\}

    <b><tt><put note here, multiple paragraphs are allowed></b></tt>

    \{\{% /alert %\}\}</b></tt>

    {{% alert title="Note" color="primary" %}}
    <put note here, multiple paragraphs are allowed>
    {{% /alert %}}

* **Embedding content**: If you want to read in content from a separate file, you can use the following format.
  Do this if you think the content might be useful in multiple pages:

  <b><tt>\{\{% content "./newfile.md" %\}\}</b></tt>

* **General style issues**: Unless otherwise instructed, follow the [Kubernetes Documentation Style Guide](https://kubernetes.io/docs/contribute/style/style-guide/) for formatting and presentation guidance.

* **Creating images**: Screen shots and other images should be published in PNG format.
  For complex diagrams, create them in SVG format, then export the image to PNG and store both in the EKS Anywhere GitHub site’s [docs/static/images](https://github.com/aws/eks-anywhere/tree/main/docs/static/images) directory.
  To include an image in a docs page, use the following format:

  <b><tt>\!\[Create workload clusters\]\(/images/eks-a_cluster_management.png\)</b></tt>

You can use tools such as [draw.io](https://diagrams.net) or image creation programs, such as inkscape, to create SVG diagrams.
Use a white background for images and have cropping focus on the content you are highlighing (in other words, don't show a whole web page if you are only interested in one field).

## Where to put content

* **Yaml examples**: Put full yaml file examples into the EKS Anywhere GitHub site’s [docs/static/manifests](https://github.com/aws/eks-anywhere/tree/main/docs/static/manifests) directory.
  In kubectl examples, you can point to those files using: `https://anywhere.eks.amazonaws.com/manifests/whatever.yaml`
* **Generic instructions for creating a cluster** should go into the [getting started]({{< relref "/docs/getting-started/chooseprovider" >}}) documentation for the appropriate provider.
* **Instructions that are specific to an EKS Anywhere provider** should go into the appropriate provider section. 

## Contributing docs for third-party solutions

To contribute documentation describing how to use third-party software products or projects with EKS Anywhere, follow these guidelines.

### Docs for third-party software in EKS Anywhere

Documentation PRs for EKS Anywhere that describe third-party software that is included in EKS Anywhere are acceptable, provided they meet the quality standards described in the Tips described below. This includes:

* Software bundled with EKS Anywhere (for example, [Cilium docs]({{< relref "../clustermgmt/networking/networking-and-security/" >}}))
* Supported platforms on which EKS Anywhere runs (for example, [VMware vSphere]({{< relref "../getting-started/vsphere/" >}}))
* Curated software that is packaged by the EKS Anywhere project to run EKS Anywhere. This includes documentation for Harbor local registry, Ingress controller, and Prometheus, Grafana, and Fluentd monitoring and logging.

### Docs for third-party software NOT in EKS Anywhere

Documentation for software that is not part of EKS Anywhere software can still be added to EKS Anywhere docs by meeting one of the following criteria:

* **Partners**: Documentation PRs for software from vendors listed on the [EKS Anywhere Partner page]({{< relref "../concepts/eksafeatures/" >}})) can be considered to add to the EKS Anywhere docs.
  Links point to partners from the [Compare EKS Anywhere to EKS](https://anywhere.eks.amazonaws.com/docs/concepts/eksafeatures/) page and other content can be added to EKS Anywhere documentation for features from those partners.
  Contact the AWS container partner team if you are interested in becoming a partner: aws-container-partners@amazon.com

### Tips for contributing third-party docs

The Kubernetes docs project itself describes a similar approach to docs covering third-party software in the [How Docs Handle Third Party and Dual Sourced Content](https://kubernetes.io/blog/2020/05/third-party-dual-sourced-content/) blog.
In line with these general guidelines, we recommend that even acceptable third-party docs contributions to EKS Anywhere:

* **Not be dual-sourced**: The project does not allow content that is already published somewhere else.
  You can provide links to that content, if it is relevant. Heavily rewriting such content to be EKS Anywhere-specific might be acceptable.
* **Not be marketing oriented**. The content shouldn’t sell a third-party products or make vague claims of quality.
* **Not outside the scope of EKS Anywhere**:  Just because some projects or products of a partner are appropriate for EKS Anywhere docs, it doesn’t mean that any project or product by that partner can be documented in EKS Anywhere.
* **Stick to the facts**:  So, for example, docs about third-party software could say: “To set up load balancer ABC, do XYZ” or “Make these modifications to improve speed and efficiency.” It should not make blanket statements like: “ABC load balancer is the best one in the industry.”
* **EKS features**: Features that relate to EKS which runs in AWS or requires an AWS account should link to [the official documentation](https://docs.aws.amazon.com/eks/) as much as possible.
