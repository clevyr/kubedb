# Kubedb

Kubedb is a command to interact with databases running in Kubernetes.
It supports dumping, restoring, and dropping into a database shell.
Optional flags are available to set running database parameters
(see each command's help entry for more details).
If no flags are given, kubedb will inspect the pod configuration and attempt
to configure itself via the Kubernetes EnvVar API.

## Installation

To install kubedb, run the following command:

```shell
brew install clevyr/tap/kubedb
```

Kubedb requires an existing Kubeconfig. See below for details. 

## Usage

[View the generated docs for usage information.](docs/kubedb.md)

### Connecting to GKE

1. To connect to a Kubernetes cluster running in GKE,
   ensure you have the `gcloud` command installed. 
   If you have it then skip to step 2.  
   Otherwise, you can either [take a look at GCP's install doc](https://cloud.google.com/sdk/docs/install), 
   or run:

   ```shell
   brew install google-cloud-sdk
   gcloud init
   ```

2. Then to generate a Kubeconfig, run:

   ```shell
   gcloud container clusters get-credentials --project=PROJECT CLUSTER_NAME
   ```
   
3. If you don’t encounter any errors then you should be connected and ready to work with databases!
   To verify, type in the following command and press the tab key twice:

   ```shell
   kubedb exec -n <TAB><TAB>
   ```

   All of your current namespaces should show up in your shell.
   Many of the kubedb flags support tab completion.
