/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v74/github"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	githubv1alpha1 "github.com/thatmlopsguy/github-k8s-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// GitHubTokenSecretKey is the key for the GitHub token in the secret
	GitHubTokenSecretKey = "token"
	// GitHubTokenSecretName is the default name for the GitHub token secret
	GitHubTokenSecretName = "github-token"
	// GitHubTokenSecretNamespace is the namespace for the GitHub token secret
	GitHubTokenSecretNamespace = "default"
	// RepositoryFinalizer is the finalizer for repository resources
	RepositoryFinalizer = "github.thatmlopsguy.com/repository-finalizer"
)

// RepositoryReconciler reconciles a GitHubRepo object
type RepositoryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=github.thatmlopsguy.com,resources=githubrepoes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=github.thatmlopsguy.com,resources=githubrepoes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=github.thatmlopsguy.com,resources=githubrepoes/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the GitHubRepo instance
	var gitHubRepo githubv1alpha1.GitHubRepo
	if err := r.Get(ctx, req.NamespacedName, &gitHubRepo); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			log.Info("GitHubRepo resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get GitHubRepo")
		return ctrl.Result{}, err
	}

	// Initialize GitHub client
	githubClient, err := r.getGitHubClient(ctx, gitHubRepo.Namespace)
	if err != nil {
		log.Error(err, "Failed to initialize GitHub client")
		return r.updateStatus(ctx, &gitHubRepo, false, "", fmt.Sprintf("Failed to initialize GitHub client: %v", err), 0)
	}

	// Handle deletion
	if gitHubRepo.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, githubClient, &gitHubRepo)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(&gitHubRepo, RepositoryFinalizer) {
		controllerutil.AddFinalizer(&gitHubRepo, RepositoryFinalizer)
		return ctrl.Result{}, r.Update(ctx, &gitHubRepo)
	}

	// Check if repository already exists on GitHub
	existingRepo, err := r.getGitHubRepository(ctx, githubClient, &gitHubRepo)
	if err != nil {
		log.Error(err, "Failed to check if repository exists")
		return r.updateStatus(ctx, &gitHubRepo, false, "", fmt.Sprintf("Failed to check repository: %v", err), 0)
	}

	if existingRepo != nil {
		// Repository exists, update status
		log.Info("Repository already exists on GitHub", "url", existingRepo.GetHTMLURL())
		return r.updateStatus(ctx, &gitHubRepo, true, existingRepo.GetHTMLURL(), "Repository exists", existingRepo.GetID())
	}

	// Create new repository
	log.Info("Creating new GitHub repository", "name", gitHubRepo.Spec.Name)
	newRepo, err := r.createGitHubRepository(ctx, githubClient, &gitHubRepo)
	if err != nil {
		log.Error(err, "Failed to create repository")
		return r.updateStatus(ctx, &gitHubRepo, false, "", fmt.Sprintf("Failed to create repository: %v", err), 0)
	}

	log.Info("Successfully created GitHub repository", "url", newRepo.GetHTMLURL())
	return r.updateStatus(ctx, &gitHubRepo, true, newRepo.GetHTMLURL(), "Repository created successfully", newRepo.GetID())
}

// getGitHubClient creates a GitHub client using the token from the secret
func (r *RepositoryReconciler) getGitHubClient(ctx context.Context, namespace string) (*github.Client, error) {
	// Get the GitHub token from the secret
	secret := &corev1.Secret{}
	secretName := types.NamespacedName{
		Name:      GitHubTokenSecretName,
		Namespace: namespace,
	}

	if err := r.Get(ctx, secretName, secret); err != nil {
		return nil, fmt.Errorf("failed to get GitHub token secret: %w", err)
	}

	token, ok := secret.Data[GitHubTokenSecretKey]
	if !ok {
		return nil, fmt.Errorf("GitHub token not found in secret")
	}

	// Create GitHub client with token
	client := &http.Client{}
	githubClient := github.NewClient(client).WithAuthToken(string(token))

	return githubClient, nil
}

// getGitHubRepository checks if a repository exists on GitHub
func (r *RepositoryReconciler) getGitHubRepository(ctx context.Context, client *github.Client, gitHubRepo *githubv1alpha1.GitHubRepo) (*github.Repository, error) {
	owner := gitHubRepo.Spec.Owner
	if owner == "" {
		// If no owner specified, use the authenticated user
		user, _, err := client.Users.Get(ctx, "")
		if err != nil {
			return nil, fmt.Errorf("failed to get authenticated user: %w", err)
		}
		owner = user.GetLogin()
	}

	repo, resp, err := client.Repositories.Get(ctx, owner, gitHubRepo.Spec.Name)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			// Repository doesn't exist
			return nil, nil
		}
		return nil, fmt.Errorf("failed to check repository: %w", err)
	}

	return repo, nil
}

// createGitHubRepository creates a new repository on GitHub
func (r *RepositoryReconciler) createGitHubRepository(ctx context.Context, client *github.Client, gitHubRepo *githubv1alpha1.GitHubRepo) (*github.Repository, error) {
	repo := &github.Repository{
		Name:        github.String(gitHubRepo.Spec.Name),
		Description: github.String(gitHubRepo.Spec.Description),
		Private:     github.Bool(gitHubRepo.Spec.Private),
	}

	owner := gitHubRepo.Spec.Owner
	createdRepo, _, err := client.Repositories.Create(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return createdRepo, nil
}

// handleDeletion handles the deletion of a GitHubRepo resource
func (r *RepositoryReconciler) handleDeletion(ctx context.Context, client *github.Client, gitHubRepo *githubv1alpha1.GitHubRepo) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if controllerutil.ContainsFinalizer(gitHubRepo, RepositoryFinalizer) {
		// Check if repository exists and delete it
		existingRepo, err := r.getGitHubRepository(ctx, client, gitHubRepo)
		if err != nil {
			log.Error(err, "Failed to check repository during deletion")
			// Continue with finalizer removal even if we can't check/delete
		} else if existingRepo != nil {
			// Delete the repository
			owner := gitHubRepo.Spec.Owner
			if owner == "" {
				user, _, err := client.Users.Get(ctx, "")
				if err != nil {
					log.Error(err, "Failed to get authenticated user during deletion")
				} else {
					owner = user.GetLogin()
				}
			}

			if owner != "" {
				_, err = client.Repositories.Delete(ctx, owner, gitHubRepo.Spec.Name)
				if err != nil {
					log.Error(err, "Failed to delete repository from GitHub")
					// Continue with finalizer removal even if deletion failed
				} else {
					log.Info("Successfully deleted repository from GitHub")
				}
			}
		}

		// Remove finalizer
		controllerutil.RemoveFinalizer(gitHubRepo, RepositoryFinalizer)
		return ctrl.Result{}, r.Update(ctx, gitHubRepo)
	}

	return ctrl.Result{}, nil
}

// updateStatus updates the status of the GitHubRepo resource
func (r *RepositoryReconciler) updateStatus(ctx context.Context, gitHubRepo *githubv1alpha1.GitHubRepo, ready bool, url, message string, githubID int64) (ctrl.Result, error) {
	gitHubRepo.Status.Ready = ready
	gitHubRepo.Status.URL = url
	gitHubRepo.Status.Message = message
	gitHubRepo.Status.GitHubID = githubID

	if err := r.Status().Update(ctx, gitHubRepo); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update status: %w", err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&githubv1alpha1.GitHubRepo{}).
		Named("repository").
		Complete(r)
}
