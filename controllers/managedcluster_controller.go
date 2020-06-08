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

package controllers

import (
	"context"
	"errors"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2019-08-01/containerservice"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	azurev1 "github.com/juan-lee/ctrlarm/api/v1alpha1"
)

const (
	statePending   = "Pending"
	stateUpdating  = "Updating"
	stateCreating  = "Creating"
	stateDeleting  = "Deleting"
	stateSucceeded = "Succeeded"
	stateFailed    = "Failed"

	// Use a dummy ssh public key because AKS nodes don't have a reachable endpoint.
	dummySSHPublicKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDiP1Zw3fPYXM7t0YrwAeSmVQn7U6MbUiLcunZ3Rwg9wtnouOYeCbnbfhrtqS04DCoKafg4u/V5vbIW8PLmmJJ5jM5eSX99k2nUjSSYTBEDy+DjxfYjJ9qsPvtWVEnzl/IEQpzvw/emXCsx1S9WN4Dhmbw6ONNGLjncvQuAzNmZgvpmKOVyXBknhUNsSBPpyEpbjo6DFpOs6v2uv4yckL3Otjgd05y/VG2qAfjDsnx61+h4/0WkW55nW+kRhV9wzEAVuhlTFcVBCWDNWmPbfTztAdBGvga4z8iDk/JqliFIihQNgUVRZclqfOpY7RL/LWTcrsE9aluDlc0A+gNGmh4KqhpBeElJ3pX7vFQGjNhrZ2tokbEMU93PboucEfQvcDBYSkNx2zzQtwH5Kt1quuP1Mcf51xyYYxKPdliGL/3J4KxRLJc0OzBRapb/+3WD1w3QK2BSLCwvxdCqcIwApYS5sCGye+AAnhTi2xkPNCDKEDSqmHJHj+2mYQAzmfnPkDpUaDw9rgqrmTjK3rqdjdUlBeiR3+xY3U2MZNy5jMvnkY3TsTtqPi3xUoZ/XnUSskBsK7IHQeem8LR3f/QPyXOCp1dRpNAtijNElF70gxT4NA3G5LXM2px5VcNaMCIAxy4Tyd68GIZbuRtCwAYlacI9GhhKhwCLXgPWCmAtr0tK3w== dummy@key` //nolint:lll
)

// ManagedClusterReconciler reconciles a ManagedCluster object
type ManagedClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type managedCluster struct {
	azurev1.ManagedCluster
	ClientID     string
	ClientSecret string
}

func newClient(subscriptionID string) (*containerservice.ManagedClustersClient, error) {
	managedClusters := containerservice.NewManagedClustersClient(subscriptionID)
	err := authorizeFromFile(&managedClusters.Client)
	if err != nil {
		return nil, err
	}
	return &managedClusters, nil
}

// +kubebuilder:rbac:groups=azure.jpang.dev,resources=managedclusters;managedclusters/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

func (r *ManagedClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&azurev1.ManagedCluster{}).
		Complete(r)
}

func (r *ManagedClusterReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) { //nolint:funlen,gocyclo
	ctx := context.Background()
	log := r.Log.WithValues("managedcluster", req.NamespacedName)

	var instance azurev1.ManagedCluster
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		return ctrl.Result{}, ignoreNotFound(err)
	}

	secretName := types.NamespacedName{Name: instance.Spec.CredentialsRef.Name, Namespace: instance.Spec.CredentialsRef.Namespace}
	var creds corev1.Secret
	if err := r.Get(ctx, secretName, &creds); err != nil {
		log.Error(err, "unable to fetch managedCluster secret")
		return ctrl.Result{}, nil
	}

	managedClusters, err := newClient(instance.Spec.SubscriptionID)
	if err != nil {
		log.Error(err, "unable to authorize containerservice client")
		return ctrl.Result{}, nil
	}

	desired := managedCluster{
		ManagedCluster: *instance.DeepCopy(),
		ClientID:       string(creds.Data["clientID"]),
		ClientSecret:   string(creds.Data["clientSecret"]),
	}

	patcher, perr := patch.NewHelper(&desired.ManagedCluster, r.Client)
	if perr != nil {
		return ctrl.Result{}, perr
	}

	defer func() {
		if derr := patcher.Patch(ctx, &desired.ManagedCluster); derr != nil {
			reterr = derr
		}
	}()

	if !desired.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, log, managedClusters, &desired)
	}

	return r.reconcile(ctx, log, managedClusters, &desired)
}

func (r *ManagedClusterReconciler) reconcile(
	ctx context.Context,
	log logr.Logger,
	c *containerservice.ManagedClustersClient,
	desired *managedCluster,
) (ctrl.Result, error) {
	if !desired.HasFinalizer() {
		desired.AddFinalizer()
	}

	found := managedCluster{ManagedCluster: *desired.ManagedCluster.DeepCopy()}
	err := r.getResource(ctx, c, &found)
	if err != nil && notFound(err) {
		err = r.reconcileCluster(ctx, log, c, desired)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	if found.Status.State != desired.Status.State {
		log.Info("ManagedCluster State Changed", "before", desired.Status.State, "after", found.Status.State)
		desired.Status.State = found.Status.State
		return ctrl.Result{}, nil
	}

	if isProvisioning(found.Status) {
		log.Info("Resource is being provisioned", "State", found.Status.State)
		return ctrl.Result{}, nil
	}

	if !reflect.DeepEqual(desired.Spec, found.Spec) {
		err = r.reconcileCluster(ctx, log, c, desired)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *ManagedClusterReconciler) reconcileDelete(
	ctx context.Context,
	log logr.Logger,
	c *containerservice.ManagedClustersClient,
	instance *managedCluster,
) (ctrl.Result, error) {
	log.Info("Deleting managedCluster")
	future, err := c.Delete(ctx, instance.Spec.ResourceGroup, instance.Spec.Name)
	if err != nil && notFound(err) {
		instance.RemoveFinalizer()
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}
	go func() {
		patcher, err := patch.NewHelper(&instance.ManagedCluster, r.Client)
		if err != nil {
			log.Error(err, "unable to create patch helper")
			return
		}
		defer func() {
			if err != nil {
				instance.Status.State = stateFailed
			}
			if derr := patcher.Patch(ctx, &instance.ManagedCluster); derr != nil {
				log.Error(derr, "unable to patch instance")
			}
		}()

		err = future.WaitForCompletionRef(ctx, c.Client)
		if err != nil {
			log.Error(err, "unable to wait for delete")
			return
		}
		_, err = future.Result(*c)
		if err != nil {
			log.Error(err, "unable to get delete result")
			return
		}
		instance.RemoveFinalizer()
		log.Info("Deleted managedCluster")
	}()
	instance.Status.State = stateDeleting
	return ctrl.Result{}, nil
}

func (r *ManagedClusterReconciler) getResource(
	ctx context.Context,
	c *containerservice.ManagedClustersClient,
	instance *managedCluster,
) error {
	found, err := c.Get(ctx, instance.Spec.ResourceGroup, instance.Spec.Name)
	if err != nil {
		return err
	}
	result, err := makeManagedCluster(instance, &found)
	if err != nil {
		return err
	}
	result.DeepCopyInto(&instance.ManagedCluster)
	return nil
}

func (r *ManagedClusterReconciler) reconcileCluster(
	ctx context.Context,
	log logr.Logger,
	c *containerservice.ManagedClustersClient,
	instance *managedCluster,
) error {
	log.Info("Reconciling managedCluster")
	future, err := c.CreateOrUpdate(ctx, instance.Spec.ResourceGroup, instance.Spec.Name, *instance.Parameters())
	if err != nil {
		return err
	}
	go func() {
		patcher, err := patch.NewHelper(&instance.ManagedCluster, r.Client)
		if err != nil {
			log.Error(err, "unable to create patch helper")
			return
		}
		defer func() {
			if err != nil {
				instance.Status.State = stateFailed
			}
			if derr := patcher.Patch(ctx, &instance.ManagedCluster); derr != nil {
				log.Error(derr, "unable to patch instance")
			}
		}()

		err = future.WaitForCompletionRef(ctx, c.Client)
		if err != nil {
			log.Error(err, "unable to wait for put")
			instance.Status.State = stateFailed
			return
		}
		found, err := future.Result(*c)
		if err != nil {
			log.Error(err, "unable to get put result")
			return
		}
		result, err := makeManagedCluster(instance, &found)
		if err != nil {
			log.Error(err, "unable to get make result")
			return
		}
		result.DeepCopyInto(&instance.ManagedCluster)
		instance.Status.State = stateSucceeded
		log.Info("Reconciled managedCluster")
	}()
	instance.Status.State = statePending
	return nil
}

func makeManagedCluster(instance *managedCluster, mc *containerservice.ManagedCluster) (*managedCluster, error) {
	if mc == nil {
		return nil, errors.New("containerservice.ManagedCluster is nil")
	}
	if mc.ManagedClusterProperties == nil {
		return nil, errors.New("containerservice.ManagedClusterProperties is nil")
	}
	if mc.ManagedClusterProperties.AgentPoolProfiles == nil {
		return nil, errors.New("containerservice.ManagedClusterProperties.AgentPoolProfiles is nil")
	}
	return &managedCluster{
		ManagedCluster: azurev1.ManagedCluster{
			TypeMeta:   instance.TypeMeta,
			ObjectMeta: *instance.ObjectMeta.DeepCopy(),
			Spec: azurev1.ManagedClusterSpec{
				AzureMeta: azurev1.AzureMeta{
					SubscriptionID: instance.Spec.SubscriptionID,
					ResourceGroup:  instance.Spec.AzureMeta.ResourceGroup,
					Location:       *mc.Location,
				},
				Name:      *mc.Name,
				Version:   *mc.KubernetesVersion,
				NodePools: makeNodePools(*mc.AgentPoolProfiles),
				CredentialsRef: corev1.SecretReference{
					Name:      instance.Spec.CredentialsRef.Name,
					Namespace: instance.Spec.CredentialsRef.Namespace,
				},
			},
			Status: azurev1.ManagedClusterStatus{
				ID:    *mc.ID,
				FQDN:  *mc.Fqdn,
				State: *mc.ProvisioningState,
			},
		},
		ClientID: *mc.ServicePrincipalProfile.ClientID,
	}, nil
}

func makeNodePools(agentpools []containerservice.ManagedClusterAgentPoolProfile) []azurev1.NodePool {
	var nodepools []azurev1.NodePool
	for n := range agentpools {
		nodepools = append(nodepools, azurev1.NodePool{
			Name:     *agentpools[n].Name,
			SKU:      string(agentpools[n].VMSize),
			Capacity: *agentpools[n].Count,
		})
	}
	return nodepools
}

func (mc *managedCluster) Parameters() *containerservice.ManagedCluster {
	return &containerservice.ManagedCluster{
		Name:     &mc.Spec.Name,
		Location: &mc.Spec.Location,
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			APIServerAccessProfile: &containerservice.ManagedClusterAPIServerAccessProfile{
				AuthorizedIPRanges: &mc.Spec.AuthorizedIPRanges,
			},
			KubernetesVersion: &mc.Spec.Version,
			DNSPrefix:         &mc.Spec.Name,
			LinuxProfile: &containerservice.LinuxProfile{
				AdminUsername: to.StringPtr("azureuser"),
				SSH: &containerservice.SSHConfiguration{
					PublicKeys: &[]containerservice.SSHPublicKey{{KeyData: to.StringPtr(dummySSHPublicKey)}},
				},
			},
			AgentPoolProfiles: makeAgentPoolProfiles(mc.Spec.NodePools),
			ServicePrincipalProfile: &containerservice.ManagedClusterServicePrincipalProfile{
				ClientID: &mc.ClientID,
				Secret:   &mc.ClientSecret,
			},
		},
	}
}

func makeAgentPoolProfiles(nodePools []azurev1.NodePool) *[]containerservice.ManagedClusterAgentPoolProfile {
	var result []containerservice.ManagedClusterAgentPoolProfile
	for _, np := range nodePools {
		result = append(result, containerservice.ManagedClusterAgentPoolProfile{
			Name:   &np.Name,
			Count:  &np.Capacity,
			VMSize: containerservice.VMSizeTypes(np.SKU),
		})
	}
	return &result
}

func authorizeFromFile(c *autorest.Client) error {
	authorizer, err := auth.NewAuthorizerFromFileWithResource(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return err
	}
	c.Authorizer = authorizer
	if err := c.AddToUserAgent("ctrlarm"); err != nil {
		return err
	}
	return nil
}

func isProvisioning(status azurev1.ManagedClusterStatus) bool {
	switch status.State {
	case statePending:
		return true
	case stateCreating:
		return true
	case "Scaling":
		return true
	case stateDeleting:
		return true
	case stateUpdating:
		return true
	default:
		return false
	}
}

func ignoreNotFound(err error) error {
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func notFound(err error) bool {
	if derr, ok := err.(autorest.DetailedError); ok && derr.StatusCode == 404 {
		return true
	}
	return false
}
