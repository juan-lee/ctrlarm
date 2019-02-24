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
	"encoding/json"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/cosmos-db/mgmt/documentdb"
	"github.com/Azure/go-autorest/autorest/to"
	cosmosdbv1alpha1 "github.com/juan-lee/ctrlarm/pkg/apis/cosmosdb/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type connectionStrings struct {
	ConnectionStrings []string `json:"connectionStrings,omitempty"`
}

// DatabaseAccount provides conversion between crd and azure sdk types
type DatabaseAccount struct {
	documentdb.DatabaseAccount `json:",inline"`
	connectionStrings          `json:",inline"`
}

// DeepCopyInto uses json marshaling to deep copy into an existing DatabaseAccount instance
func (db *DatabaseAccount) DeepCopyInto(out *DatabaseAccount) {
	byt, _ := json.Marshal(*db)
	json.Unmarshal(byt, out)
}

// DeepCopy clones a DatabaseAccount instance
func (db *DatabaseAccount) DeepCopy() *DatabaseAccount {
	if db == nil {
		return nil
	}
	out := new(DatabaseAccount)
	db.DeepCopyInto(out)
	return out
}

// IsEqual compares DatabaseAccount instances
func (db *DatabaseAccount) IsEqual(other *DatabaseAccount) bool {
	return reflect.DeepEqual(db, other)
}

// IsProvisioning returns true if the DatabaseAccount is being updated
func (db *DatabaseAccount) IsProvisioning() bool {
	return *db.ProvisioningState != "Succeeded" && *db.ProvisioningState != "Failed"
}

// Merge combines the current and desired state
func (db *DatabaseAccount) Merge(desired *cosmosdbv1alpha1.DatabaseAccount, s *corev1.Secret) *DatabaseAccount {
	result := db.DeepCopy()
	return result
}

func (db *DatabaseAccount) createUpdateParameters() *documentdb.DatabaseAccountCreateUpdateParameters {
	dacup := documentdb.DatabaseAccountCreateUpdateParameters{
		Kind:     db.Kind,
		ID:       db.ID,
		Name:     db.Name,
		Type:     db.Type,
		Location: db.Location,
		Tags:     db.Tags,
	}
	if db.DatabaseAccountProperties != nil {
		locations := append(*db.ReadLocations, *db.WriteLocations...)
		dacup.DatabaseAccountCreateUpdateProperties = &documentdb.DatabaseAccountCreateUpdateProperties{
			Locations:                     &locations,
			DatabaseAccountOfferType:      to.StringPtr(string(db.DatabaseAccountOfferType)),
			IPRangeFilter:                 db.IPRangeFilter,
			IsVirtualNetworkFilterEnabled: db.IsVirtualNetworkFilterEnabled,
			EnableAutomaticFailover:       db.EnableAutomaticFailover,
			Capabilities:                  db.Capabilities,
			VirtualNetworkRules:           db.VirtualNetworkRules,
			EnableMultipleWriteLocations:  db.EnableMultipleWriteLocations,
		}

		if db.DatabaseAccountProperties.ConsistencyPolicy != nil {
			dacup.ConsistencyPolicy = &documentdb.ConsistencyPolicy{
				DefaultConsistencyLevel: db.ConsistencyPolicy.DefaultConsistencyLevel,
				MaxStalenessPrefix:      db.ConsistencyPolicy.MaxStalenessPrefix,
				MaxIntervalInSeconds:    db.ConsistencyPolicy.MaxIntervalInSeconds,
			}
		}
	}
	return &dacup
}
