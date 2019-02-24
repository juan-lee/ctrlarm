// Copyright 2019 The ctrlarm Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package managedcluster

import (
	"context"
	"time"

	containerservicesv1alpha1 "github.com/juan-lee/ctrlarm/pkg/apis/containerservices/v1alpha1"
	"github.com/juan-lee/ctrlarm/pkg/services/azure/containerservices/managedcluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

const managedClusterFinalizer = "managedcluster.containerservices.azure.com"

// Add creates a new ManagedCluster Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileManagedCluster{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("managedcluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to ManagedCluster
	err = c.Watch(&source.Kind{Type: &containerservicesv1alpha1.ManagedCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileManagedCluster{}

// ReconcileManagedCluster reconciles a ManagedCluster object
type ReconcileManagedCluster struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ManagedCluster object and makes changes based on the state read
// and what is in the ManagedCluster.Spec
// +kubebuilder:rbac:groups=containerservices.azure.com,resources=managedclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=containerservices.azure.com,resources=managedclusters/status,verbs=get;update;patch
func (r *ReconcileManagedCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the ManagedCluster instance
	instance := &containerservicesv1alpha1.ManagedCluster{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	secret := &corev1.Secret{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Identity.SecretName, Namespace: instance.Namespace}, secret)
	if err != nil {
		return reconcile.Result{}, err
	}

	log.Info("NewManagedClusterActuator", "subscriptionID", instance.Spec.SubscriptionID)
	mca, err := managedcluster.NewManagedClusterActuator(context.TODO(), instance.Spec.SubscriptionID)
	if err != nil {
		return reconcile.Result{}, err
	}

	found, err := mca.Get(context.TODO(), instance.Spec.ResourceGroup, instance.Name)
	if _, ok := err.(*managedcluster.ClusterNotFound); ok {
		if isDeletePending(instance) {
			err = r.completeFinalize(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{Requeue: true}, err
			}
			return reconcile.Result{}, nil
		}

		log.Info("Create ManagedCluster", "namespace", instance.Namespace, "name", instance.Name)
		err = mca.Create(context.TODO(), instance, secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, err
	} else if err != nil {
		return reconcile.Result{}, err
	}

	instance.Status.ProvisioningState = string(*found.ProvisioningState)
	log.Info("ManagedCluster ProvisioningState", "state", instance.Status.ProvisioningState)
	if err = r.Status().Update(context.TODO(), instance); err != nil {
		return reconcile.Result{}, err
	}

	if found.IsProvisioning() {
		log.Info("Reconciling ManagedCluster", "namespace", instance.Namespace, "name", instance.Name, "state", instance.Status.ProvisioningState)
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, err
	}

	if !isDeletePending(instance) {
		if !hasFinalizer(instance.ObjectMeta.Finalizers) {
			err = r.addFinalizer(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{Requeue: true}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
	} else {
		if hasFinalizer(instance.ObjectMeta.Finalizers) {
			log.Info("Delete ManagedCluster", "namespace", instance.Namespace, "name", instance.Name)
			err = mca.Delete(context.TODO(), instance.Spec.ResourceGroup, instance.Name)
			if err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, err
		}
	}

	desired := found.Merge(instance, nil)
	if !found.IsEqual(desired) {
		log.Info("Update ManagedCluster", "namespace", instance.Namespace, "name", instance.Name)
		err = mca.Update(context.TODO(), instance.Spec.ResourceGroup, instance.Name, desired)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 30}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileManagedCluster) addFinalizer(ctx context.Context, instance *containerservicesv1alpha1.ManagedCluster) error {
	log.Info("Adding ManagedCluster Finalizer", "namespace", instance.Namespace, "name", instance.Name)
	instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, managedClusterFinalizer)
	if err := r.Update(ctx, instance); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileManagedCluster) completeFinalize(ctx context.Context, instance *containerservicesv1alpha1.ManagedCluster) error {
	instance.ObjectMeta.Finalizers = removeFinializer(instance.ObjectMeta.Finalizers)
	if err := r.Update(ctx, instance); err != nil {
		return err
	}
	log.Info("ManagedCluster Finialization Complete", "namespace", instance.Namespace, "name", instance.Name)
	return nil
}

func hasFinalizer(slice []string) bool {
	for _, item := range slice {
		if item == managedClusterFinalizer {
			return true
		}
	}
	return false
}

func removeFinializer(slice []string) (result []string) {
	for _, item := range slice {
		if item == managedClusterFinalizer {
			continue
		}
		result = append(result, item)
	}
	return
}

func isDeletePending(instance *containerservicesv1alpha1.ManagedCluster) bool {
	return !instance.ObjectMeta.DeletionTimestamp.IsZero()
}
