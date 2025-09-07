# GitHub Kubernetes Operator

A Kubernetes operator for managing GitHub repositories using the GitHub API.

This operator allows you to create, manage, and delete GitHub repositories through Kubernetes Custom Resource Definitions (CRDs).

## Quick Start

### Prerequisites

1. **GitHub Personal Access Token**

2. **Kubernetes Cluster**: A running Kubernetes cluster (local or remote)

### Installation

1. **Clone and build the operator**:

   ```bash
   git clone https://github.com/thatmlopsguy/github-k8s-operator
   cd github-k8s-operator

   # Build the Docker image
   make docker-build

   # For kind clusters, load the image
   kind load docker-image thatmlopsguy/github-operator:latest --name github-operator-test-e2e
   ```

2. **Create GitHub token secret**:

   ```bash
   # Replace 'your_github_token_here' with your actual GitHub token
   kubectl create secret generic github-token \
     --from-literal=token=your_github_token_here \
     --namespace=default
   ```

3. **Install CRDs and deploy the operator**:

   ```bash
   # Install CRDs
   make install

   # Deploy the operator
   make deploy
   ```

4. **Verify the deployment**:

   ```bash
   kubectl get pods -n github-operator-system
   kubectl logs -n github-operator-system deployment/github-operator-controller-manager
   ```

### Usage

Create a GitHub repository by applying a GitHubRepo resource:

```yaml
apiVersion: github.thatmlopsguy.com/v1alpha1
kind: GitHubRepo
metadata:
  name: test-public-repo
  namespace: default
spec:
  name: test-public-repo
  description: "Testing public repository creation"
  private: false
  license: mit
  gitignore: Go
```

Apply it:

```bash
kubectl apply -f your-repo-config.yaml
```

Check the status:

```bash
kubectl get githubrepo test-public-repo -o yaml
```

## Configuration

### CRD Specification

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Name of the GitHub repository |
| `description` | string | No | Repository description |
| `private` | bool | No | Set to `true` for private repositories, `false` for public (default: false) |
| `gitignore` | string | No | Gitignore template to apply (e.g., Go, Python, Node, etc.) |
| `license` | string | No | License template to apply (e.g., mit, apache-2.0, gpl-3.0, etc.) |
| `owner` | string | No | GitHub user or organization that will own the repository (defaults to authenticated user) |

### Status Fields

The operator reports status in the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `ready` | bool | Whether the repository was successfully created |
| `url` | string | GitHub repository URL |
| `message` | string | Status message with details |
| `githubId` | int64 | GitHub repository ID |

### Secret Configuration

The controller looks for a secret named `github-token` in the same namespace as the GitHubRepo resource:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-token
  namespace: default
type: Opaque
data:
  token: <base64-encoded-github-token>
```

**Important**: Ensure your token is properly base64 encoded without trailing newlines:

```bash
echo -n "your_github_token" | base64
```

## Limitations

### Repository Updates

The GitHub operator supports **repository creation and limited updates** to existing repositories:

- For supported fields (visibility, description), the controller will detect changes and update the GitHub repository accordingly
- For unsupported fields (license, gitignore), changes are only applied during repository creation

## Development

### Bootstrapping the Operator

To recreate this operator using Operator SDK:

1. Install Operator SDK ([installation guide](https://sdk.operatorframework.io/docs/installation/))

2. Initialize the project:

   ```bash
   operator-sdk init --domain github.com --repo github.com/thatmlopsguy/github-k8s-operator --project-name github-operator
   ```

3. Create the API and controller:

   ```bash
   operator-sdk create api --group github --version v1alpha1 --kind Repository --resource --controller
   ```

4. Add go-github dependency:

   ```bash
   go get github.com/google/go-github/v74
   ```

### Building and Testing

```bash
# Generate code and manifests
make generate
make manifests

# Build and test
make build
make test

# Build Docker image
make docker-build

# Deploy for testing
make deploy
```

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.
