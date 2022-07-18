# Curated Packages Customized Configuration

## Problem

Currently, the eks anywhere CLI provides `install` command. This command's aim is to provide the ability for users to install a curated package in the cluster. Currently, this command does not support the ability to override any configurations that would enable the user to customize the curated package.

## Goals and Objectives

As a curated packages user, I want to:

* Provide custom configurations to override default configurations from the CLI

## Possible Solutions

### Provide separate key value pairs from the CLI

The first solution would be to provide separate key value pairs through the cli similar to helm (https://all.docs.genesys.com/PrivateEdition/Current/PEGuide/HelmOverrides)

*Sample*

```bash
$ eksctl anywhere install package harbor --set key=key --set key2=key2
```

### Provide consolidated key value pairs from the CLI

Similar to the first solution, the solution would be to consolidate all the key value pairs into one but separated by a delimiter, similar to cobra (https://github.com/spf13/pflag/pull/133)

*Sample*

```bash
$ eksctl anywhere install package harbor --config=key1=key1, key2=key2
```

### Provide key value pair through a file and pass the file to the CLI

This solution is different than both the previous solution given that it takes a file for the configuration

*Sample*

```bash
$ cat config.txt
secretKey=use-a-secret-key
harborAdminPassword=test
$ eksctl anywhere install package harbor --config ./config.txt
```

## Proposed Solution
Since the amount of configurations required for curated packages is a limited set, we will be implementing [Provide separate key value pairs from the CLI](#Provide-separate-key-value-pairs-from-the-CLI) option.

## Customized Configuration Discovery

In order for the user to identify what configurations are available, the CLI needs to provide a mechanism to identify the configurations available for a package. Similar to helm's show values, we would provide a capability that is inline with this format.

The command would look like below

```bash 
$ eksctl anywhere install package harbor --source cluster --show-options
```