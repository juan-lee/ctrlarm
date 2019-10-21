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

// AzureMeta represents metadata that all azure resources must have.
type AzureMeta struct {
	// SubscriptionID is the subscription id for an azure resource.
	// +kubebuilder:validation:Pattern=`^[0-9A-Fa-f]{8}(?:-[0-9A-Fa-f]{4}){3}-[0-9A-Fa-f]{12}$`
	SubscriptionID string `json:"subscriptionID,omitempty"`

	// ResourceGroup is the resource group name for an azure resource.
	// +kubebuilder:validation:Pattern=`^[-\w\._\(\)]+$`
	ResourceGroup string `json:"resourceGroup,omitempty"`

	// Location is the region where the azure resource resides.
	Location string `json:"location,omitempty"`
}

// NodePool defines a node pool for an azure cluster resource.
type NodePool struct {
	// Name of the node pool.
	Name string `json:"name,omitempty"`

	// SKU of the VMs in the node pool.
	SKU string `json:"sku,omitempty"`

	// Capacity is the number of VMs in a node pool.
	Capacity int32 `json:"capacity,omitempty"`
}
