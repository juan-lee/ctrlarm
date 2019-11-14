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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (in *ManagedCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// nolint:lll
// +kubebuilder:webhook:path=/mutate-azure-jpang-dev-v1alpha1-managedcluster,mutating=true,failurePolicy=fail,groups=azure.jpang.dev,resources=managedclusters,verbs=create;update,versions=v1alpha1,name=mmanagedcluster.kb.io

var _ webhook.Defaulter = &ManagedCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *ManagedCluster) Default() {
}

// nolint:lll
// +kubebuilder:webhook:verbs=create;update;delete,path=/validate-azure-jpang-dev-v1alpha1-managedcluster,mutating=false,failurePolicy=fail,groups=azure.jpang.dev,resources=managedclusters,versions=v1alpha1,name=vmanagedcluster.kb.io

var _ webhook.Validator = &ManagedCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *ManagedCluster) ValidateCreate() error {
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *ManagedCluster) ValidateUpdate(old runtime.Object) error {
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *ManagedCluster) ValidateDelete() error {
	return nil
}

var _ conversion.Hub = &ManagedCluster{}

func (in *ManagedCluster) Hub() {}
