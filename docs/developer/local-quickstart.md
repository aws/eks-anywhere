# Quickstart Guide For Local Development

## Important Make commands:

Run from project root.

Build an executable:
```
make eks-a
```

Use the executable from project root with:
```
./bin/eksctl-anywhere <command>
```

Lint the project:
```
make lint
```

Note that to correct linting issues you will need to install [gofumpt](https://github.com/mvdan/gofumpt) and [gci](https://github.com/daixiang0/gci).

Invoke gofumpt:
```
gofumpt -w <target files or directory>
```

Invoke gci:
```
gci write  -s 'standard,default,prefix(github.com/aws/eks-anywhere)' <target files or directory>
```

Run project unit tests:
```
make unit-test
```


## Debugging
If using VSCode, you can use the following as a `launch.json` configuration:
```
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "create-cluster-debug",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/eksctl-anywhere/main.go",
            "cwd": "${workspaceFolder}",
            "env": {},
            "envFile": "${workspaceFolder}/.env",
            "buildFlags": "-ldflags='-X github.com/aws/eks-anywhere/pkg/version.gitVersion=v0.0.0-dev -X github.com/aws/eks-anywhere/pkg/cluster.releasesManifestURL=https://dev-release-prod-pdx.s3.us-west-2.amazonaws.com/eks-a-release.yaml'",
            "args": ["create", "cluster", "-f", "eksa-mgmt-cluster.yaml"],
        }
    ]
}
```

## Building e2e Tests

Building e2e tests requires adding build tags. You can read more about what they are and how they work here.

If using vscode, add this to your `.vscode/settings.json` file:
```
{
    ...
    "go.buildFlags": [
        "-tags=e2e,all_providers"
    ],
    ...
}
```