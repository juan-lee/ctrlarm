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

package cosmosdb

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2015-04-08/documentdb"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	cosmosdbv1alpha1 "github.com/juan-lee/ctrlarm/pkg/apis/cosmosdb/v1alpha1"
	"github.com/juan-lee/ctrlarm/pkg/services/azure"
	"github.com/juan-lee/ctrlarm/pkg/services/azure/config"
	corev1 "k8s.io/api/core/v1"
)

// DatabaseAccountNotFound error indicates the managed cluster doesn't exist
type DatabaseAccountNotFound struct{}

// Error implements the error interface
func (e *DatabaseAccountNotFound) Error() string {
	return "DatabaseAccountNotFound"
}

// CosmosConnectionStringsNotFound error indicates the managed cluster doesn't exist
type CosmosConnectionStringsNotFound struct{}

// Error implements the error interface
func (e *CosmosConnectionStringsNotFound) Error() string {
	return "CosmosConnectionStringsNotFound"
}

// DatabaseAccountActuator is an interface for managing a DatabaseAccount
type DatabaseAccountActuator interface {
	Get(ctx context.Context, rg, name string) (*DatabaseAccount, error)
	Create(ctx context.Context, desired *cosmosdbv1alpha1.DatabaseAccount, secret *corev1.Secret) error
	Update(ctx context.Context, rg, name string, db *DatabaseAccount) error
	Delete(ctx context.Context, rg, name string) error
}

type databaseAccountActuator struct {
	client *documentdb.DatabaseAccountsClient
}

// NewDatabaseAccountActuator creates an authenticated DatabaseAccountActuator instance
func NewDatabaseAccountActuator(ctx context.Context, subscriptionID string) (DatabaseAccountActuator, error) {
	err := config.ParseEnvironment()
	if err != nil {
		return nil, err
	}

	auth, err := azure.GetResourceManagementAuthorizer()
	if err != nil {
		return nil, err
	}

	client := documentdb.NewDatabaseAccountsClient(subscriptionID)
	client.Authorizer = auth

	return &databaseAccountActuator{client: &client}, nil
}

func (c *databaseAccountActuator) Get(ctx context.Context, rg, name string) (*DatabaseAccount, error) {
	found, err := c.client.Get(ctx, rg, name)
	if err != nil {
		if derr, ok := err.(autorest.DetailedError); ok && derr.StatusCode == 404 {
			return nil, &DatabaseAccountNotFound{}
		}
		return nil, err
	}

	csr, err := c.client.ListConnectionStrings(ctx, rg, name)
	if err != nil {
		return nil, &CosmosConnectionStringsNotFound{}
	}

	// copy to strip out http.Response from found
	db := DatabaseAccount{DatabaseAccount: found}
	db.ConnectionStrings = []string{}
	for _, cs := range *csr.ConnectionStrings {
		db.ConnectionStrings = append(db.ConnectionStrings, *cs.ConnectionString)
	}
	return db.DeepCopy(), nil
}

func (c *databaseAccountActuator) Create(ctx context.Context, desired *cosmosdbv1alpha1.DatabaseAccount, secret *corev1.Secret) error {
	dacup := documentdb.DatabaseAccountCreateUpdateParameters{
		Kind:     documentdb.DatabaseAccountKind(desired.Spec.Kind),
		Name:     to.StringPtr(desired.Name),
		Location: to.StringPtr(desired.Spec.Location),
		DatabaseAccountCreateUpdateProperties: &documentdb.DatabaseAccountCreateUpdateProperties{
			Locations: &[]documentdb.Location{
				{
					LocationName:     to.StringPtr(desired.Spec.Location),
					FailoverPriority: to.Int32Ptr(0),
				},
			},
			DatabaseAccountOfferType: to.StringPtr(string(documentdb.Standard)),
		},
	}
	_, err := c.client.CreateOrUpdate(ctx, desired.Spec.ResourceGroup, desired.Name, dacup)
	if err != nil {
		return err
	}
	return nil
}

func (c *databaseAccountActuator) Update(ctx context.Context, rg, name string, db *DatabaseAccount) error {
	_, err := c.client.CreateOrUpdate(ctx, rg, name, *db.createUpdateParameters())
	if err != nil {
		return err
	}
	return nil
}

func (c *databaseAccountActuator) Delete(ctx context.Context, rg, name string) error {
	_, err := c.client.Delete(ctx, rg, name)
	if err != nil {
		return err
	}
	return nil
}
