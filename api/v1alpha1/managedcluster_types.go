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

const ManagedClusterFinalizer = "managedcluster.azure.jpang.dev"

// ManagedClusterSpec defines the desired state of ManagedCluster
type ManagedClusterSpec struct {
	AzureMeta `json:",inline"`

	// Name defines the name of the azure kubernetes cluster resource.
	Name string `json:"name,omitempty"`

	// Version defines the kubernetes version of the cluster.
	Version string `json:"version,omitempty"`

	// NodePools defines the node pools in a azure kubernetes cluster resource.
	NodePools []NodePool `json:"nodePools"`

	// AuthorizedIPRanges gives a list of CIDR addresses that should be whitelisted.
	AuthorizedIPRanges []string `json:"authorizedIPRanges,omitempty"`

	// CredentialsRef is a reference to the azure kubernetes cluster credentials.
	CredentialsRef corev1.SecretReference `json:"credentialsRef,omitempty"`
}

// ManagedClusterStatus defines the observed state of ManagedCluster
type ManagedClusterStatus struct {
	// ID represents the cluster resource id.
	ID string `json:"id,omitempty"`

	// State represents the provisioning state of the cluster resource.
	State string `json:"state,omitempty"`

	// FQDN represents the cluster api server endpoint.
	FQDN string `json:"fqdn,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=managedclusters,shortName=mc,scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="FQDN",type="string",JSONPath=".status.fqdn",description="API endpoint FQDN"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="Provisioning state of the cluster resource"

// ManagedCluster is the Schema for the managedclusters API
type ManagedCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedClusterSpec   `json:"spec,omitempty"`
	Status ManagedClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ManagedClusterList contains a list of ManagedCluster
type ManagedClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedCluster `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&ManagedCluster{}, &ManagedClusterList{})
}

func (in *ManagedCluster) HasFinalizer() bool {
	for _, f := range in.ObjectMeta.Finalizers {
		if f == ManagedClusterFinalizer {
			return true
		}
	}
	return false
}

func (in *ManagedCluster) AddFinalizer() {
	if !in.HasFinalizer() {
		in.ObjectMeta.Finalizers = append(in.ObjectMeta.Finalizers, ManagedClusterFinalizer)
	}
}

func (in *ManagedCluster) RemoveFinalizer() {
	var result []string
	for _, f := range in.ObjectMeta.Finalizers {
		if f == ManagedClusterFinalizer {
			continue
		}
		result = append(result, f)
	}
	in.ObjectMeta.Finalizers = result
}
