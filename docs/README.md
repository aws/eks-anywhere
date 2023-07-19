# EKS Anywhere Documentation

EKS Documentation lives in this folder.
It uses [`hugo`](https://gohugo.io/) for static site generation with the [docsy](https://docsy.dev) for a theme.

## Local development

To run the local documentation server you can use the provided make targets to create a container with the required dependencies.

```bash
make container-build
make container-serve
```

Open http://127.0.0.1:1313 to see the local site.
With the serve container running you can now edit the documentation in your git clone and changes will be rebuilt automatically.

To serve documentation more permanently, detach the container using the `DETACH=true` option to the `container-serve` recipe.

```bash
make container-build
make container-serve DETACH=true
```

You may notice an update to the `/docs/themes/docsy` submodule which is the result of a patch. The submodule patches should not be committed with your docs changes. Before committing your changes, reset the submodule using `make submodule-reset` (this will impact any containers serving the site locally).

## Public development

If you want to make a version of the docs site you can share with someone else you will need to follow these steps.

1. Create an Amplify app in your AWS account
1. Create a "main" branch in your Amplify app
1. Deploy the app using your `$AMPLIFY_APP_ID`

```
export AWS_PROFILE=<YOUR AWS ACCOUNT INFORMATION>
export AMPLIFY_APP_ID=$(aws amplify create-app --name eksa-docs --query 'app.appId' --output text)
aws amplify create-branch --app-id $AMPLIFY_APP_ID --branch-name main --stage PRODUCTION
cd docs
make deploy
# Get your docs URL
echo "https://main.$(aws amplify get-app --app-id $AMPLIFY_APP_ID --query 'app.defaultDomain' --output text)"
```

## Production site

The production website is hosted on Amplify.
To deploy the docs to a personal amplify app you need to first create an app with a branch.
Export your `${AWS_PROFILE}`, `${AMPLIFY_APP_ID}`, and `${AMPLIFY_APP_BRANCH}` (default: main).

Then run
```bash
make deploy
```
It will build the site, create a zip, and deploy it to your Amplify app.

If you want to connect a custom domain you can do that manually in Amplify/Route53 or you can look at the CDK deployment infrastructure.

### Website versions

Each website version has a unique subdomain url (eg v1.anywhere.eks.amazonaws.com) so users can view different versions of the documentation.
