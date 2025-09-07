# GitHub Kubernetes Operator

A Kubernetes operator for managing GitHub repositories using the GitHub API. This operator allows you to create, manage, and delete GitHub repositories through Kubernetes Custom Resource Definitions (CRDs).

## Features

- ✅ **Repository Creation**: Automatically creates repositories on GitHub based on CRD specifications
- ✅ **Repository Deletion**: Handles cleanup and repository deletion when CRDs are removed
- ✅ **Status Tracking**: Reports repository creation status, URL, and GitHub ID
- ✅ **Flexible Configuration**: Supports public/private repositories with descriptions
- ✅ **Owner Management**: Supports specifying repository owner (defaults to authenticated user)
- ✅ **Finalizer Support**: Proper cleanup using Kubernetes finalizers
- ✅ **Error Handling**: Comprehensive error handling with detailed status updates

## Architecture

- **CRD**: `github.thatmlopsguy.com/v1alpha1` with `GitHubRepo` kind
- **Controller**: Full reconciliation loop with create, update, delete operations
- **GitHub API Integration**: Using go-github v74 library
- **Authentication**: GitHub Personal Access Token via Kubernetes Secret
- **RBAC**: Proper permissions for secrets and CRD management

## Quick Start

### Prerequisites

1. **GitHub Personal Access Token**: Create a token with the following scopes:
   - `repo` (for full control of private repositories)
   - `public_repo` (for access to public repositories)

2. **Kubernetes Cluster**: A running Kubernetes cluster (local or remote)

### Installation

1. **Clone and build the operator**:
   ```bash
   git clone https://github.com/thatmlopsguy/github-k8s-operator
   cd github-k8s-operator
   
   # Build the Docker image
   make docker-build IMG=thatmlopsguy/github-operator:latest
   
   # For kind clusters, load the image
   kind load docker-image thatmlopsguy/github-operator:latest --name your-cluster-name
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

## Testing Results

The operator has been successfully tested with the following scenarios:

### ✅ Successful Test Cases

1. **Public Repository Creation**:
   - Created `test-public-repo` with `private: false`
   - Status: `ready: true`, proper URL and GitHub ID returned

2. **Private Repository Creation**:
   - Created `test-private-repo` with `private: true`
   - Status: `ready: true`, proper URL and GitHub ID returned

3. **Repository Deletion**:
   - Finalizer-based cleanup working correctly
   - Repositories deleted from GitHub when CRD is deleted
   - Proper logging: "Successfully deleted repository from GitHub"

4. **Existing Repository Detection**:
   - Controller detects existing repositories
   - Status: "Repository exists" with proper GitHub ID and URL

5. **Error Recovery**:
   - Fixed GitHub token formatting issues
   - Handled organization vs user account creation properly

## Troubleshooting

### Common Issues and Solutions

#### 1. Image Pull Errors in Kind Clusters

**Problem**: `ErrImagePull` or `ImagePullBackOff` errors

**Solution**: Ensure the image is loaded into kind and set `imagePullPolicy: Never`:
```bash
# Load image into kind
kind load docker-image thatmlopsguy/github-operator:latest --name your-cluster-name

# Verify image is loaded
docker exec -it your-cluster-control-plane crictl images | grep github-operator
```

#### 2. GitHub Token Authentication Errors

**Problem**: `net/http: invalid header field value for "Authorization"`

**Cause**: Trailing newline characters in base64-encoded token

**Solution**: Ensure proper base64 encoding without newlines:
```bash
# Wrong (includes newline)
echo "your_token" | base64

# Correct (no newline)
echo -n "your_token" | base64
```

#### 3. Organization vs User Repository Creation

**Problem**: `404 Not Found` when creating repositories with owner specified

**Cause**: When `owner` is specified, GitHub API treats it as an organization

**Solution**: 
- For personal repositories: omit the `owner` field (defaults to authenticated user)
- For organization repositories: ensure the authenticated user has permissions in that organization

#### 4. Controller Not Starting

**Problem**: Controller pods not starting or crashlooping

**Solutions**:
1. Check RBAC permissions:
   ```bash
   kubectl get clusterrole github-operator-manager-role -o yaml
   ```
2. Verify secret exists and is readable:
   ```bash
   kubectl get secret github-token -o yaml
   ```
3. Check controller logs:
   ```bash
   kubectl logs -n github-operator-system deployment/github-operator-controller-manager
   ```

#### 5. Repository Creation Failures

**Problem**: Repository creation fails with API errors

**Common Causes and Solutions**:
- **Token permissions**: Ensure token has `repo` and `public_repo` scopes
- **Repository exists**: Controller will detect and update status accordingly
- **Rate limiting**: GitHub API rate limits apply (check logs for rate limit errors)
- **Network connectivity**: Ensure cluster can reach `api.github.com`

### Debugging Commands

```bash
# Check operator status
kubectl get pods -n github-operator-system
kubectl logs -n github-operator-system deployment/github-operator-controller-manager --tail=50

# Check GitHubRepo resources
kubectl get githubrepo -A
kubectl describe githubrepo your-repo-name

# Verify secret
kubectl get secret github-token -o yaml

# Check CRD installation
kubectl get crd githubrepoes.github.thatmlopsguy.com

# Force reconciliation (add annotation to trigger)
kubectl annotate githubrepo your-repo-name reconcile-trigger=$(date +%s)
```

### Log Messages to Look For

**Successful Operations**:
- `"Creating new GitHub repository"`
- `"Successfully created GitHub repository"`
- `"Repository already exists on GitHub"`
- `"Successfully deleted repository from GitHub"`

**Error Conditions**:
- `"Failed to get GitHub token secret"`
- `"Failed to create repository"`
- `"Failed to check repository"`
- `"net/http: invalid header field value"`

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
make docker-build IMG=your-registry/github-operator:tag

# Deploy for testing
make deploy IMG=your-registry/github-operator:tag
```

## Security Considerations

- Store GitHub tokens securely in Kubernetes secrets
- Use tokens with minimal required permissions
- Consider using GitHub Apps for production deployments for better security
- Rotate tokens regularly
- Monitor operator logs for authentication issues
- Use namespace isolation for multi-tenant environments

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.
