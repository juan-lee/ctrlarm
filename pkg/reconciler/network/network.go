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

package network

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-11-01/network"
	aznet "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-11-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/go-logr/logr"
	k8sv1alpha1 "github.com/juan-lee/ctrlarm/pkg/apis/kubernetes/v1alpha1"
	kubeadmv1beta1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta1"
)

type Reconciler struct {
	log        logr.Logger
	vnetClient *aznet.VirtualNetworksClient
}

// ProvideAzureReconciler provides a reconciler for cluster networking
func ProvideReconciler(log logr.Logger, vc *aznet.VirtualNetworksClient) *Reconciler {
	return &Reconciler{
		log:        log,
		vnetClient: vc,
	}
}

func (r Reconciler) Reconcile(ctx context.Context, c *kubeadmv1beta1.Networking) (*k8sv1alpha1.ClusterStatus, error) {
	r.log.Info("Reconciling network")
	_, err := r.vnetClient.CreateOrUpdate(ctx, "jpang-rg", "jpang-rg-a90s", aznet.VirtualNetwork{
		Location: to.StringPtr("some-location"),
		VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
			AddressSpace: &network.AddressSpace{
				AddressPrefixes: &[]string{"10.0.0.0/8"},
			},
		},
	})
	if err != nil {
		r.log.Error(err, "Failed to reconcile network")
		return nil, err
	}
	r.log.Info("Network Reconciled")
	return nil, nil
}
