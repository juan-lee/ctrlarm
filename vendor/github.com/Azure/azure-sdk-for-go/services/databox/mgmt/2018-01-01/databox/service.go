package databox

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"context"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/validation"
	"github.com/Azure/go-autorest/tracing"
	"net/http"
)

// ServiceClient is the client for the Service methods of the Databox service.
type ServiceClient struct {
	BaseClient
}

// NewServiceClient creates an instance of the ServiceClient client.
func NewServiceClient(subscriptionID string) ServiceClient {
	return NewServiceClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewServiceClientWithBaseURI creates an instance of the ServiceClient client.
func NewServiceClientWithBaseURI(baseURI string, subscriptionID string) ServiceClient {
	return ServiceClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// ListAvailableSkus this method provides the list of available skus for the given subscription and location.
// Parameters:
// location - the location of the resource
// availableSkuRequest - filters for showing the available skus.
func (client ServiceClient) ListAvailableSkus(ctx context.Context, location string, availableSkuRequest AvailableSkuRequest) (result AvailableSkusResultPage, err error) {
	if tracing.IsEnabled() {
		ctx = tracing.StartSpan(ctx, fqdn+"/ServiceClient.ListAvailableSkus")
		defer func() {
			sc := -1
			if result.asr.Response.Response != nil {
				sc = result.asr.Response.Response.StatusCode
			}
			tracing.EndSpan(ctx, sc, err)
		}()
	}
	if err := validation.Validate([]validation.Validation{
		{TargetValue: availableSkuRequest,
			Constraints: []validation.Constraint{{Target: "availableSkuRequest.TransferType", Name: validation.Null, Rule: true, Chain: nil},
				{Target: "availableSkuRequest.Country", Name: validation.Null, Rule: true, Chain: nil},
				{Target: "availableSkuRequest.Location", Name: validation.Null, Rule: true, Chain: nil}}}}); err != nil {
		return result, validation.NewError("databox.ServiceClient", "ListAvailableSkus", err.Error())
	}

	result.fn = client.listAvailableSkusNextResults
	req, err := client.ListAvailableSkusPreparer(ctx, location, availableSkuRequest)
	if err != nil {
		err = autorest.NewErrorWithError(err, "databox.ServiceClient", "ListAvailableSkus", nil, "Failure preparing request")
		return
	}

	resp, err := client.ListAvailableSkusSender(req)
	if err != nil {
		result.asr.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "databox.ServiceClient", "ListAvailableSkus", resp, "Failure sending request")
		return
	}

	result.asr, err = client.ListAvailableSkusResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "databox.ServiceClient", "ListAvailableSkus", resp, "Failure responding to request")
	}

	return
}

// ListAvailableSkusPreparer prepares the ListAvailableSkus request.
func (client ServiceClient) ListAvailableSkusPreparer(ctx context.Context, location string, availableSkuRequest AvailableSkuRequest) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"location":       autorest.Encode("path", location),
		"subscriptionId": autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2018-01-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/providers/Microsoft.DataBox/locations/{location}/availableSkus", pathParameters),
		autorest.WithJSON(availableSkuRequest),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// ListAvailableSkusSender sends the ListAvailableSkus request. The method will close the
// http.Response Body if it receives an error.
func (client ServiceClient) ListAvailableSkusSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req,
		azure.DoRetryWithRegistration(client.Client))
}

// ListAvailableSkusResponder handles the response to the ListAvailableSkus request. The method always
// closes the http.Response Body.
func (client ServiceClient) ListAvailableSkusResponder(resp *http.Response) (result AvailableSkusResult, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// listAvailableSkusNextResults retrieves the next set of results, if any.
func (client ServiceClient) listAvailableSkusNextResults(ctx context.Context, lastResults AvailableSkusResult) (result AvailableSkusResult, err error) {
	req, err := lastResults.availableSkusResultPreparer(ctx)
	if err != nil {
		return result, autorest.NewErrorWithError(err, "databox.ServiceClient", "listAvailableSkusNextResults", nil, "Failure preparing next results request")
	}
	if req == nil {
		return
	}
	resp, err := client.ListAvailableSkusSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		return result, autorest.NewErrorWithError(err, "databox.ServiceClient", "listAvailableSkusNextResults", resp, "Failure sending next results request")
	}
	result, err = client.ListAvailableSkusResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "databox.ServiceClient", "listAvailableSkusNextResults", resp, "Failure responding to next results request")
	}
	return
}

// ListAvailableSkusComplete enumerates all values, automatically crossing page boundaries as required.
func (client ServiceClient) ListAvailableSkusComplete(ctx context.Context, location string, availableSkuRequest AvailableSkuRequest) (result AvailableSkusResultIterator, err error) {
	if tracing.IsEnabled() {
		ctx = tracing.StartSpan(ctx, fqdn+"/ServiceClient.ListAvailableSkus")
		defer func() {
			sc := -1
			if result.Response().Response.Response != nil {
				sc = result.page.Response().Response.Response.StatusCode
			}
			tracing.EndSpan(ctx, sc, err)
		}()
	}
	result.page, err = client.ListAvailableSkus(ctx, location, availableSkuRequest)
	return
}

// ValidateAddressMethod this method validates the customer shipping address and provide alternate addresses if any.
// Parameters:
// location - the location of the resource
// validateAddress - shipping address of the customer.
func (client ServiceClient) ValidateAddressMethod(ctx context.Context, location string, validateAddress ValidateAddress) (result AddressValidationOutput, err error) {
	if tracing.IsEnabled() {
		ctx = tracing.StartSpan(ctx, fqdn+"/ServiceClient.ValidateAddressMethod")
		defer func() {
			sc := -1
			if result.Response.Response != nil {
				sc = result.Response.Response.StatusCode
			}
			tracing.EndSpan(ctx, sc, err)
		}()
	}
	if err := validation.Validate([]validation.Validation{
		{TargetValue: validateAddress,
			Constraints: []validation.Constraint{{Target: "validateAddress.ShippingAddress", Name: validation.Null, Rule: true,
				Chain: []validation.Constraint{{Target: "validateAddress.ShippingAddress.StreetAddress1", Name: validation.Null, Rule: true, Chain: nil},
					{Target: "validateAddress.ShippingAddress.Country", Name: validation.Null, Rule: true, Chain: nil},
					{Target: "validateAddress.ShippingAddress.PostalCode", Name: validation.Null, Rule: true, Chain: nil},
				}}}}}); err != nil {
		return result, validation.NewError("databox.ServiceClient", "ValidateAddressMethod", err.Error())
	}

	req, err := client.ValidateAddressMethodPreparer(ctx, location, validateAddress)
	if err != nil {
		err = autorest.NewErrorWithError(err, "databox.ServiceClient", "ValidateAddressMethod", nil, "Failure preparing request")
		return
	}

	resp, err := client.ValidateAddressMethodSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "databox.ServiceClient", "ValidateAddressMethod", resp, "Failure sending request")
		return
	}

	result, err = client.ValidateAddressMethodResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "databox.ServiceClient", "ValidateAddressMethod", resp, "Failure responding to request")
	}

	return
}

// ValidateAddressMethodPreparer prepares the ValidateAddressMethod request.
func (client ServiceClient) ValidateAddressMethodPreparer(ctx context.Context, location string, validateAddress ValidateAddress) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"location":       autorest.Encode("path", location),
		"subscriptionId": autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2018-01-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/providers/Microsoft.DataBox/locations/{location}/validateAddress", pathParameters),
		autorest.WithJSON(validateAddress),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// ValidateAddressMethodSender sends the ValidateAddressMethod request. The method will close the
// http.Response Body if it receives an error.
func (client ServiceClient) ValidateAddressMethodSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req,
		azure.DoRetryWithRegistration(client.Client))
}

// ValidateAddressMethodResponder handles the response to the ValidateAddressMethod request. The method always
// closes the http.Response Body.
func (client ServiceClient) ValidateAddressMethodResponder(resp *http.Response) (result AddressValidationOutput, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}
