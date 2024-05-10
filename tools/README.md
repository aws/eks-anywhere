# Intelligent Support Bundle Analyzer

Support bundles are commonly attached to tickets or provided by customers during debugging session for engineers to debug issues with the cluster. These support bundles often have too much information to go through to narrow down and investigate what is actually happening on the cluster. This support-bundle analyzer would be able to take in a support bundle and run analyzers and be able to produce a list of must know metadata about the cluster like version, replicas, provider, machine status, etc as well as list out potential problems with the clusters by analyzing the logs and status.

The app is built using Fyne, a Go-based UI toolkit and app API.

### Building the application

Currently the support-bundle analyzer app has been built and tested only on MacOS. You will need Go 1.21 to build and run the app.

The first step is to install the Fyne Go module and command-line utility on your system. This can be done using the following commands:
```console
$ go get fyne.io/fyne/v2@latest
$ go install fyne.io/fyne/v2/cmd/fyne@latest
```

You can check your Fyne installation with a simple version command:
```console
$ fyne --version
fyne version v2.4.2 
```

You can now build and package the application using the `fyne package` command:
```console
$ fyne package -os darwin -icon ./eks-anywhere-transparent.png --name support-bundle-ui --use-raw-icon=true
```

This will create a `support-bundle-ui.app` folder in the current working directory. If you open the directory on Finder, the application `support-bundle-ui` will show up with the provided icon. You can double-click the app to launch it.

### Using the application

The app has 3 main buttons on its home page - the Folder button (:file:), the Exit button and the Reset button. Click the Open button to launch a file dialog and select the support bundle you would like to analyze.

The app will then display the results of the analysis, providing you with useful information such as cluster metadata, status of different controllers running on the cluster, CRDs and logs.

If you want to reset the application, you can click on the Reset button to remove all the information on screen and present you with the homepage once again.

If you are done using the analyzer, you can click the Exit button to close the application.

# Future scope

* Add more analyzers
* Potentially integrate with EKS-A either as a running web service, or another sub-command in the CLI