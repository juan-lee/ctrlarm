/*
Copyright 2019 (c) Microsoft and contributors. All rights reserved.

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

// NodeSpec represents the configuration of an agent VM
type NodeSpec struct {
	VMSize       string `json:"vmSize,omitempty"`
	OsDiskSizeGB int32  `json:"osDiskSizeGB,omitempty"`
	VnetSubnetID string `json:"vnetSubnetID,omitempty"`
	MaxPods      int32  `json:"maxPods,omitempty"`
}

// NodePoolSpec represents the configuration of an agent VM pool
type NodePoolSpec struct {
	Name     string   `json:"name,omitempty"`
	Size     int32    `json:"size,omitempty"`
	NodeSpec NodeSpec `json:"nodeSpec,omitempty"`
}

// AddonSpec represents the configuration of a managed cluster addon
type AddonSpec struct {
	Enabled bool              `json:"enabled,omitempty"`
	Config  map[string]string `json:"config"`
}

// NetworkSpec represents the managed cluster's network configuration
type NetworkSpec struct {
	Plugin           string `json:"plugin,omitempty"`
	Policy           string `json:"policy,omitempty"`
	PodCidr          string `json:"podCidr,omitempty"`
	ServiceCidr      string `json:"serviceCidr,omitempty"`
	DNSServiceIP     string `json:"dnsServiceIP,omitempty"`
	DockerBridgeCidr string `json:"dockerBridgeCidr,omitempty"`
}

// IdentitySpec represents the managed cluster's identity
type IdentitySpec struct {
	ClientID   string `json:"clientID,omitempty"`
	SecretName string `json:"secretName,omitempty"`
}

// LinuxSpec represents the linux configuration of agent VMs
type LinuxSpec struct {
	AdminUsername  string   `json:"adminUsername,omitempty"`
	AuthorizedKeys []string `json:"authorizedKeys,omitempty"`
}

// ManagedClusterSpec defines the desired state of an AKS ManagedCluster
type ManagedClusterSpec struct {
	SubscriptionID    string               `json:"subscriptionID,omitempty"`
	ResourceGroup     string               `json:"resourceGroup,omitempty"`
	Location          string               `json:"location,omitempty"`
	KubernetesVersion string               `json:"kubernetesVersion,omitempty"`
	DNSPrefix         string               `json:"dnsPrefix,omitempty"`
	Rbac              bool                 `json:"rbac,omitempty"`
	Addons            map[string]AddonSpec `json:"addons,omitempty"`
	Network           NetworkSpec          `json:"network,omitempty"`
	NodePools         []NodePoolSpec       `json:"nodePools,omitempty"`
	Linux             LinuxSpec            `json:"linux,omitempty"`
	Identity          IdentitySpec         `json:"identity,omitempty"`
	Tags              map[string]string    `json:"tags,omitempty"`
}

// ManagedClusterStatus defines the observed state of an AKS ManagedCluster
type ManagedClusterStatus struct {
	ProvisioningState string `json:"provisioningState,omitempty"`
	ResourceID        string `json:"resourceID,omitempty"`
	Fqdn              string `json:"fqdn,omitempty"`
	NodeResourceGroup string `json:"nodeResourceGroup,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManagedCluster is the Schema for the managedclusters API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ResourceGroup",type="string",JSONPath=".spec.resourceGroup",description=""
// +kubebuilder:printcolumn:name="Location",type="string",JSONPath=".spec.location",description=""
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.kubernetesVersion",description=""
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.provisioningState",description=""
type ManagedCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedClusterSpec   `json:"spec,omitempty"`
	Status ManagedClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManagedClusterList contains a list of ManagedCluster
type ManagedClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedCluster{}, &ManagedClusterList{})
}
