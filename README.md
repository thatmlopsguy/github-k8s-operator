# github-k8s-operator

## Bootstrapping the Operator

To create this operator using Operator SDK (Go version):

1. Install Operator SDK ([installation guide](https://sdk.operatorframework.io/docs/installation/))
1. Initialize the project:

```sh
operator-sdk init --domain github.com --repo github.com/thatmlopsguy/github-k8s-operator --project-name github-operator
```

1. Create the API and controller for the GitHub Repository resource:

```sh
operator-sdk create api --group github --version v1alpha1 --kind Repository --resource --controller
```

1. Generate CRD manifests:

```sh
make manifests
```

1. (Optional) Install the CRD to your cluster:

```sh
make install
```

You can now implement your controller logic in `internal/controller/repository_controller.go` and define the CRD spec in `api/v1alpha1/repository_types.go`.

## Troubleshooting

If you encounter errors when running `make install` (e.g., connection timeouts to the Kubernetes API server), verify your cluster is running and your kubeconfig is set correctly:

```sh
kubectl config get-contexts
kubectl cluster-info
```

You may need to check your network connection or update your kubeconfig file.
