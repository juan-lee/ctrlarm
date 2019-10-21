/*

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManagedClusterSpec defines the desired state of ManagedCluster
type ManagedClusterSpec struct {
	AzureMeta `json:",inline"`

	// Name defines the name of the azure kubernetes cluster resource.
	Name string `json:"name,omitempty"`

	// NodePools defines the node pools in an azure kubernetes cluster resource.
	NodePools []NodePool `json:"nodePools"`

	// CredentialsRef is a reference to the azure kubernetes cluster credentials.
	CredentialsRef corev1.LocalObjectReference `json:"credentialsRef,omitempty"`
}

// ManagedClusterStatus defines the observed state of ManagedCluster
type ManagedClusterStatus struct {
}

// +kubebuilder:object:root=true

// ManagedCluster is the Schema for the managedclusters API
type ManagedCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedClusterSpec   `json:"spec,omitempty"`
	Status ManagedClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ManagedClusterList contains a list of ManagedCluster
type ManagedClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedCluster `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&ManagedCluster{}, &ManagedClusterList{})
}