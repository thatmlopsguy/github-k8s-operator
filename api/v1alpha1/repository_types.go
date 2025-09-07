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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RepositorySpec defines the desired state of GitHubRepo.
type RepositorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name is the name of the GitHub repository
	Name string `json:"name"`

	// Description is the description of the GitHub repository
	Description string `json:"description,omitempty"`

	// Private indicates whether the repository should be private or public
	Private bool `json:"private,omitempty"`

	// Owner is the GitHub user or organization that will own the repository
	// If not specified, the repository will be created under the authenticated user
	Owner string `json:"owner,omitempty"`
}

// RepositoryStatus defines the observed state of GitHubRepo.
type RepositoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Ready indicates if the repository has been successfully created
	Ready bool `json:"ready,omitempty"`

	// URL is the GitHub repository URL
	URL string `json:"url,omitempty"`

	// Message provides additional information about the repository status
	Message string `json:"message,omitempty"`

	// GitHubID is the GitHub repository ID
	GitHubID int64 `json:"githubId,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GitHubRepo is the Schema for the repositories API.
type GitHubRepo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec,omitempty"`
	Status RepositoryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GitHubRepoList contains a list of GitHubRepo.
type GitHubRepoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitHubRepo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitHubRepo{}, &GitHubRepoList{})
}
