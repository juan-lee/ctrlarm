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

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice"
	"github.com/Azure/go-autorest/autorest"
	containerservicesv1alpha1 "github.com/juan-lee/ctrlarm/pkg/apis/containerservices/v1alpha1"
	"github.com/juan-lee/ctrlarm/pkg/services/azure"
	"github.com/juan-lee/ctrlarm/pkg/services/azure/config"
	corev1 "k8s.io/api/core/v1"
)

// ClusterNotFound error indicates the managed cluster doesn't exist
type ClusterNotFound struct{}

// Error implements the error interface
func (e *ClusterNotFound) Error() string {
	return "ClusterNotFound"
}

// ManagedClusterActuator is an interface for managing a ManagedCluster
type ManagedClusterActuator interface {
	Get(ctx context.Context, rg, name string) (*ManagedCluster, error)
	Create(ctx context.Context, desired *containerservicesv1alpha1.ManagedCluster, secret *corev1.Secret) error
	Update(ctx context.Context, rg, name string, mc *ManagedCluster) error
	Delete(ctx context.Context, rg, name string) error
}

type clusterActuator struct {
	client *containerservice.ManagedClustersClient
}

// NewManagedClusterActuator creates an authenticated ManagedClusterActuator instance
func NewManagedClusterActuator(ctx context.Context, subscriptionID string) (ManagedClusterActuator, error) {
	err := config.ParseEnvironment()
	if err != nil {
		return nil, err
	}

	auth, err := azure.GetResourceManagementAuthorizer()
	if err != nil {
		return nil, err
	}

	client := containerservice.NewManagedClustersClient(subscriptionID)
	client.Authorizer = auth

	return &clusterActuator{client: &client}, nil
}

func (c clusterActuator) Get(ctx context.Context, rg, name string) (*ManagedCluster, error) {
	found, err := c.client.Get(ctx, rg, name)
	if err != nil {
		if derr, ok := err.(autorest.DetailedError); ok && derr.StatusCode == 404 {
			return nil, &ClusterNotFound{}
		}
		return nil, err
	}

	// copy to strip out http.Response from found
	mc := ManagedCluster{ManagedCluster: found}
	return mc.DeepCopy(), nil
}

func (c clusterActuator) Create(ctx context.Context, desired *containerservicesv1alpha1.ManagedCluster, secret *corev1.Secret) error {
	empty := &ManagedCluster{}
	mc := empty.Merge(desired, secret)
	_, err := c.client.CreateOrUpdate(ctx, desired.Spec.ResourceGroup, desired.Name, mc.ManagedCluster)
	if err != nil {
		return err
	}
	return nil
}

func (c clusterActuator) Update(ctx context.Context, rg, name string, mc *ManagedCluster) error {
	_, err := c.client.CreateOrUpdate(ctx, rg, name, mc.ManagedCluster)
	if err != nil {
		return err
	}
	return nil
}

func (c clusterActuator) Delete(ctx context.Context, rg, name string) error {
	_, err := c.client.Delete(ctx, rg, name)
	if err != nil {
		return err
	}
	return nil
}
