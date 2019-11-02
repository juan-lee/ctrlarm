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
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	azurev1 "github.com/juan-lee/ctrlarm/api/v1alpha1"
)

const (
	// Use a dummy ssh public key because AKS nodes don't have a reachable endpoint.
	dummySSHPublicKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDiP1Zw3fPYXM7t0YrwAeSmVQn7U6MbUiLcunZ3Rwg9wtnouOYeCbnbfhrtqS04DCoKafg4u/V5vbIW8PLmmJJ5jM5eSX99k2nUjSSYTBEDy+DjxfYjJ9qsPvtWVEnzl/IEQpzvw/emXCsx1S9WN4Dhmbw6ONNGLjncvQuAzNmZgvpmKOVyXBknhUNsSBPpyEpbjo6DFpOs6v2uv4yckL3Otjgd05y/VG2qAfjDsnx61+h4/0WkW55nW+kRhV9wzEAVuhlTFcVBCWDNWmPbfTztAdBGvga4z8iDk/JqliFIihQNgUVRZclqfOpY7RL/LWTcrsE9aluDlc0A+gNGmh4KqhpBeElJ3pX7vFQGjNhrZ2tokbEMU93PboucEfQvcDBYSkNx2zzQtwH5Kt1quuP1Mcf51xyYYxKPdliGL/3J4KxRLJc0OzBRapb/+3WD1w3QK2BSLCwvxdCqcIwApYS5sCGye+AAnhTi2xkPNCDKEDSqmHJHj+2mYQAzmfnPkDpUaDw9rgqrmTjK3rqdjdUlBeiR3+xY3U2MZNy5jMvnkY3TsTtqPi3xUoZ/XnUSskBsK7IHQeem8LR3f/QPyXOCp1dRpNAtijNElF70gxT4NA3G5LXM2px5VcNaMCIAxy4Tyd68GIZbuRtCwAYlacI9GhhKhwCLXgPWCmAtr0tK3w== dummy@key` //nolint:lll
)

// ManagedClusterReconciler reconciles a ManagedCluster object
type ManagedClusterReconciler struct {
	client.Client
	Log logr.Logger
}

type managedClusterClient struct {
	*containerservice.ManagedClustersClient

	instance *azurev1.ManagedCluster
}

type managedCluster struct {
	azurev1.ManagedCluster
	ClientID     string
	ClientSecret string
}

func newClient(instance *azurev1.ManagedCluster) (*managedClusterClient, error) {
	managedClusters := containerservice.NewManagedClustersClient(instance.Spec.SubscriptionID)
	err := authorizeFromFile(&managedClusters.Client)
	if err != nil {
		return nil, err
	}
	return &managedClusterClient{
		instance:              instance,
		ManagedClustersClient: &managedClusters,
	}, nil
}

// +kubebuilder:rbac:groups=azure.jpang.dev,resources=managedclusters;managedclusters/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

func (r *ManagedClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&azurev1.ManagedCluster{}).
		Complete(r)
}

func (r *ManagedClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) { //nolint:funlen,gocyclo
	ctx := context.Background()
	log := r.Log.WithValues("managedcluster", req.NamespacedName)

	var instance azurev1.ManagedCluster
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		log.Error(err, "unable to fetch managedCluster")
		return ctrl.Result{}, ignoreNotFound(err)
	}

	log.Info("Reconciling managedCluster", "instance", instance)

	var creds corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{
		Name:      instance.Spec.CredentialsRef.Name,
		Namespace: req.NamespacedName.Namespace,
	}, &creds); err != nil {
		log.Error(err, "unable to fetch managedCluster")
		return ctrl.Result{}, nil
	}

	managedClusters, err := newClient(&instance)
	if err != nil {
		log.Error(err, "unable to authorize containerservice client")
		return ctrl.Result{}, nil
	}

	desired := managedCluster{
		ManagedCluster: *instance.DeepCopy(),
		ClientID:       string(creds.Data["clientID"]),
		ClientSecret:   string(creds.Data["clientSecret"]),
	}

	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if instance.HasFinalizer() {
			if err = managedClusters.Delete(ctx, &desired); err != nil {
				log.Error(err, "unable to delete cluster")
				return ctrl.Result{}, nil
			}
			instance.RemoveFinalizer()
			if err = r.Update(ctx, &instance); err != nil {
				log.Error(err, "unable to update cluster spec")
				return ctrl.Result{}, nil
			}
		}
		return ctrl.Result{}, nil
	}

	found := managedCluster{}
	err = managedClusters.Get(ctx, &found)
	if err != nil && notFound(err) {
		log.Info("Creating managedCluster", "desired.Spec", desired.Spec)
		err = managedClusters.Create(ctx, &desired)
		if err != nil {
			log.Error(err, "unable to create cluster")
			return ctrl.Result{}, nil
		}
		desired.AddFinalizer()
		if err = r.Update(ctx, &desired.ManagedCluster); err != nil {
			log.Error(err, "unable to update cluster spec")
			return ctrl.Result{}, nil
		}
		err = r.Status().Update(ctx, &desired.ManagedCluster)
		if err != nil {
			log.Error(err, "unable to update cluster status")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "unable to get managedCluster")
		return ctrl.Result{}, nil
	}

	if !reflect.DeepEqual(desired.Spec, found.Spec) {
		log.Info("Updating managedCluster", "desired.Spec", desired.Spec)
		found.Spec = desired.Spec
		err = managedClusters.Update(ctx, &found)
		if err != nil {
			log.Error(err, "unable to update cluster")
			return ctrl.Result{}, nil
		}
	}

	if !reflect.DeepEqual(desired.Status, found.Status) {
		log.Info("Updating Status", "found.Status", found.Status)
		err = r.Status().Update(ctx, &found.ManagedCluster)
		if err != nil {
			log.Error(err, "unable to update cluster status")
			return ctrl.Result{}, nil
		}
	}
	return ctrl.Result{}, nil
}

func (c managedClusterClient) Get(ctx context.Context, instance *managedCluster) error {
	found, err := c.ManagedClustersClient.Get(ctx, c.instance.Spec.ResourceGroup, c.instance.Spec.Name)
	if err != nil {
		return err
	}
	result, err := c.makeManagedCluster(&found)
	if err != nil {
		return err
	}
	result.DeepCopyInto(&instance.ManagedCluster)
	return nil
}

func (c managedClusterClient) Create(ctx context.Context, instance *managedCluster) error {
	return c.Update(ctx, instance)
}

func (c managedClusterClient) Update(ctx context.Context, instance *managedCluster) error {
	future, err := c.CreateOrUpdate(ctx, instance.Spec.ResourceGroup, instance.Spec.Name, *instance.Parameters())
	if err != nil {
		return err
	}
	err = future.WaitForCompletionRef(ctx, c.Client)
	if err != nil {
		return err
	}
	found, err := future.Result(*c.ManagedClustersClient)
	if err != nil {
		return err
	}
	result, err := c.makeManagedCluster(&found)
	if err != nil {
		return err
	}
	result.DeepCopyInto(&instance.ManagedCluster)
	return nil
}

func (c managedClusterClient) Delete(ctx context.Context, instance *managedCluster) error {
	future, err := c.ManagedClustersClient.Delete(ctx, instance.Spec.ResourceGroup, instance.Spec.Name)
	if err != nil {
		return err
	}
	err = future.WaitForCompletionRef(ctx, c.Client)
	if err != nil {
		return err
	}
	_, err = future.Result(*c.ManagedClustersClient)
	if err != nil {
		return err
	}
	return nil
}

func (c managedClusterClient) makeManagedCluster(mc *containerservice.ManagedCluster) (*managedCluster, error) {
	if mc == nil {
		return nil, errors.New("containerservice.ManagedCluster is nil")
	}
	if mc.ManagedClusterProperties == nil {
		return nil, errors.New("containerservice.ManagedClusterProperties is nil")
	}
	return &managedCluster{
		ManagedCluster: azurev1.ManagedCluster{
			ObjectMeta: *c.instance.ObjectMeta.DeepCopy(),
			Spec: azurev1.ManagedClusterSpec{
				AzureMeta: azurev1.AzureMeta{
					SubscriptionID: c.instance.Spec.SubscriptionID,
					ResourceGroup:  c.instance.Spec.AzureMeta.ResourceGroup,
					Location:       *mc.Location,
				},
				Name:      *mc.Name,
				Version:   *mc.KubernetesVersion,
				NodePools: makeNodePools(*mc.AgentPoolProfiles),
				CredentialsRef: corev1.LocalObjectReference{
					Name: c.instance.Spec.CredentialsRef.Name,
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
