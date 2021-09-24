# EKS Anywhere Documentation

EKS Documentation lives in this folder.
It uses [`hugo`](https://gohugo.io/) for static site generation with the [docsy](https://docsy.dev) for a theme.

## Local development

To run the local documentation server, you can use the provided make targets to create a container with the required dependencies.

```bash
make container-build
make container-serve
```

Open http://127.0.0.1:1313 to see the local site.
With the serve container running, you can now edit the documentation in your Git clone and changes will be rebuilt automatically.

## Production site

The production website is hosted on Amplify.
To deploy the docs to a personal Amplify app, you need to first create an app with a branch.
Export your `${AWS_PROFILE}`, `${AMPLIFY_APP_ID}`, and `${AMPLIFY_APP_BRANCH}` (default: main).

Then run
```bash
make deploy
```

It will build the site, create a zip, and deploy it to your Amplify app.

If you want to connect a custom domain, you can do that manually in Amplify/Route53, or you can look at the CDK deployment infrastructure.

### Website versions

Each website version has a unique subdomain URL (e.g., v1.anywhere.eks.amazonaws.com) so users can view different versions of the documentation.
