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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DatabaseAccountSpec defines the desired state of DatabaseAccount
type DatabaseAccountSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SubscriptionID string `json:"subscriptionID,omitempty"`
	ResourceGroup  string `json:"resourceGroup,omitempty"`
	Kind           string `json:"kind,omitempty"`
	Location       string `json:"location,omitempty"`
}

// DatabaseAccountStatus defines the observed state of DatabaseAccount
type DatabaseAccountStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ProvisioningState string `json:"provisioningState,omitempty"`
	ConnectionString  string `json:"connectionString,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DatabaseAccount is the Schema for the databaseaccounts API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ResourceGroup",type="string",JSONPath=".spec.resourceGroup",description=""
// +kubebuilder:printcolumn:name="Location",type="string",JSONPath=".spec.location",description=""
// +kubebuilder:printcolumn:name="Kind",type="string",JSONPath=".spec.kind",description=""
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.provisioningState",description=""
type DatabaseAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseAccountSpec   `json:"spec,omitempty"`
	Status DatabaseAccountStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DatabaseAccountList contains a list of DatabaseAccount
type DatabaseAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseAccount{}, &DatabaseAccountList{})
}
