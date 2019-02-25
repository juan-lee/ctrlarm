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

package cluster

import (
	"context"

	"github.com/go-logr/logr"
	k8sv1alpha1 "github.com/juan-lee/ctrlarm/pkg/apis/kubernetes/v1alpha1"
	"github.com/juan-lee/ctrlarm/pkg/reconciler/network"
	kubeadmv1beta1 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta1"
)

// Reconciler defines the interface for reconciling desired state
type NetworkReconciler interface {
	Reconcile(ctx context.Context, c *kubeadmv1beta1.Networking) (*k8sv1alpha1.ClusterStatus, error)
}

type Reconciler struct {
	log logr.Logger
	net NetworkReconciler
}

func provideReconciler(log logr.Logger, nr *network.Reconciler) *Reconciler {
	return &Reconciler{
		log: log,
		net: nr,
	}
}

func (r Reconciler) Reconcile(ctx context.Context, c *k8sv1alpha1.Cluster) (*k8sv1alpha1.ClusterStatus, error) {
	r.log.Info("Reconciling network")
	status, err := r.net.Reconcile(ctx, &c.Spec.Config.Networking)
	if err != nil {
		r.log.Error(err, "Failed to reconcile network")
		return status, err
	}
	r.log.Info("Network Reconciled")
	return status, nil
}
