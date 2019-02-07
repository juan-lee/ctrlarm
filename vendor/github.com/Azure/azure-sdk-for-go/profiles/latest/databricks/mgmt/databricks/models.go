// +build go1.9

// Copyright 2019 Microsoft Corporation
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

// This code was auto-generated by:
// github.com/Azure/azure-sdk-for-go/tools/profileBuilder

package databricks

import (
	"context"

	original "github.com/Azure/azure-sdk-for-go/services/databricks/mgmt/2018-04-01/databricks"
)

const (
	DefaultBaseURI = original.DefaultBaseURI
)

type ProvisioningState = original.ProvisioningState

const (
	Accepted  ProvisioningState = original.Accepted
	Canceled  ProvisioningState = original.Canceled
	Created   ProvisioningState = original.Created
	Creating  ProvisioningState = original.Creating
	Deleted   ProvisioningState = original.Deleted
	Deleting  ProvisioningState = original.Deleting
	Failed    ProvisioningState = original.Failed
	Ready     ProvisioningState = original.Ready
	Running   ProvisioningState = original.Running
	Succeeded ProvisioningState = original.Succeeded
	Updating  ProvisioningState = original.Updating
)

type BaseClient = original.BaseClient
type ErrorDetail = original.ErrorDetail
type ErrorInfo = original.ErrorInfo
type ErrorResponse = original.ErrorResponse
type Operation = original.Operation
type OperationDisplay = original.OperationDisplay
type OperationListResult = original.OperationListResult
type OperationListResultIterator = original.OperationListResultIterator
type OperationListResultPage = original.OperationListResultPage
type OperationsClient = original.OperationsClient
type Resource = original.Resource
type Sku = original.Sku
type TrackedResource = original.TrackedResource
type Workspace = original.Workspace
type WorkspaceListResult = original.WorkspaceListResult
type WorkspaceListResultIterator = original.WorkspaceListResultIterator
type WorkspaceListResultPage = original.WorkspaceListResultPage
type WorkspaceProperties = original.WorkspaceProperties
type WorkspaceProviderAuthorization = original.WorkspaceProviderAuthorization
type WorkspaceUpdate = original.WorkspaceUpdate
type WorkspacesClient = original.WorkspacesClient
type WorkspacesCreateOrUpdateFuture = original.WorkspacesCreateOrUpdateFuture
type WorkspacesDeleteFuture = original.WorkspacesDeleteFuture
type WorkspacesUpdateFuture = original.WorkspacesUpdateFuture

func New(subscriptionID string) BaseClient {
	return original.New(subscriptionID)
}
func NewOperationListResultIterator(page OperationListResultPage) OperationListResultIterator {
	return original.NewOperationListResultIterator(page)
}
func NewOperationListResultPage(getNextPage func(context.Context, OperationListResult) (OperationListResult, error)) OperationListResultPage {
	return original.NewOperationListResultPage(getNextPage)
}
func NewOperationsClient(subscriptionID string) OperationsClient {
	return original.NewOperationsClient(subscriptionID)
}
func NewOperationsClientWithBaseURI(baseURI string, subscriptionID string) OperationsClient {
	return original.NewOperationsClientWithBaseURI(baseURI, subscriptionID)
}
func NewWithBaseURI(baseURI string, subscriptionID string) BaseClient {
	return original.NewWithBaseURI(baseURI, subscriptionID)
}
func NewWorkspaceListResultIterator(page WorkspaceListResultPage) WorkspaceListResultIterator {
	return original.NewWorkspaceListResultIterator(page)
}
func NewWorkspaceListResultPage(getNextPage func(context.Context, WorkspaceListResult) (WorkspaceListResult, error)) WorkspaceListResultPage {
	return original.NewWorkspaceListResultPage(getNextPage)
}
func NewWorkspacesClient(subscriptionID string) WorkspacesClient {
	return original.NewWorkspacesClient(subscriptionID)
}
func NewWorkspacesClientWithBaseURI(baseURI string, subscriptionID string) WorkspacesClient {
	return original.NewWorkspacesClientWithBaseURI(baseURI, subscriptionID)
}
func PossibleProvisioningStateValues() []ProvisioningState {
	return original.PossibleProvisioningStateValues()
}
func UserAgent() string {
	return original.UserAgent() + " profiles/latest"
}
func Version() string {
	return original.Version()
}
