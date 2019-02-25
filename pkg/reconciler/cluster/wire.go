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

// +build wireinject

package cluster

import (
	aznet "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-11-01/network"
	"github.com/go-logr/logr"
	"github.com/google/wire"
	"github.com/juan-lee/ctrlarm/pkg/reconciler/network"
	"github.com/juan-lee/ctrlarm/pkg/services/azure"
	"github.com/juan-lee/ctrlarm/pkg/services/azure/config"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func NewReconciler() (*Reconciler, error) {
	panic(wire.Build(
		provideLogger,
		provideNetworkClient,
		network.ProvideReconciler,
		provideReconciler,
	))
}

func provideLogger() logr.Logger {
	return logf.Log.WithName("azure-reconciler")
}

func provideNetworkClient() (*aznet.VirtualNetworksClient, error) {
	err := config.ParseEnvironment()
	if err != nil {
		return nil, err
	}

	auth, err := azure.GetResourceManagementAuthorizer()
	if err != nil {
		return nil, err
	}

	client := aznet.NewVirtualNetworksClient(config.SubscriptionID())
	client.Authorizer = auth

	return &client, nil
}
