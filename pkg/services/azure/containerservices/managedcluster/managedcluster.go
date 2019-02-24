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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice"
	"github.com/Azure/go-autorest/autorest/to"
	resourcesv1alpha1 "github.com/juan-lee/ctrlarm/pkg/apis/containerservices/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultKubernetesVersion = "1.11.6"
	defaultAdminUser         = "azureuser"
)

// ManagedCluster provides conversion between crd and azure sdk types
type ManagedCluster struct {
	containerservice.ManagedCluster `json:",inline"`
}

// DeepCopyInto uses json marshaling to deep copy into an existing ManagedCluster instance
func (mc *ManagedCluster) DeepCopyInto(out *ManagedCluster) {
	byt, _ := json.Marshal(*mc)
	json.Unmarshal(byt, out)
}

// DeepCopy clones a ManagedCluster instance
func (mc *ManagedCluster) DeepCopy() *ManagedCluster {
	if mc == nil {
		return nil
	}
	out := new(ManagedCluster)
	mc.DeepCopyInto(out)
	return out
}

// IsEqual compares ManagedCluster instances
func (mc *ManagedCluster) IsEqual(other *ManagedCluster) bool {
	return reflect.DeepEqual(mc, other)
}

// IsProvisioning returns true if the ManagedCluster is being updated
func (mc *ManagedCluster) IsProvisioning() bool {
	return *mc.ProvisioningState != "Succeeded" && *mc.ProvisioningState != "Failed"
}

// Merge combines the current and desired state
func (mc *ManagedCluster) Merge(desired *resourcesv1alpha1.ManagedCluster, s *corev1.Secret) *ManagedCluster {
	result := mc.DeepCopy()
	result.mergeRoot(desired)
	result.mergeNetwork(desired)
	result.mergeIdentity(desired, s)
	result.mergeAddons(desired)
	result.mergeNodePools(desired)
	return result
}

func (mc *ManagedCluster) mergeRoot(desired *resourcesv1alpha1.ManagedCluster) {
	if mc.ManagedClusterProperties == nil {
		mc.ManagedClusterProperties = &containerservice.ManagedClusterProperties{}
	}

	if len(desired.Name) > 0 {
		mc.Name = to.StringPtr(string(desired.Name))
	}

	if len(desired.Spec.Location) > 0 {
		mc.Location = to.StringPtr(string(desired.Spec.Location))
	}

	if len(desired.Spec.KubernetesVersion) > 0 {
		mc.KubernetesVersion = to.StringPtr(string(desired.Spec.KubernetesVersion))
	}

	if mc.KubernetesVersion == nil {
		mc.KubernetesVersion = to.StringPtr(defaultKubernetesVersion)
	}

	if len(desired.Spec.DNSPrefix) > 0 {
		mc.DNSPrefix = to.StringPtr(string(desired.Spec.DNSPrefix))
	}

	if mc.DNSPrefix == nil {
		mc.DNSPrefix = to.StringPtr(fmt.Sprintf("%s-%s", desired.Name, desired.Spec.ResourceGroup))
	}

	mc.EnableRBAC = to.BoolPtr(bool(desired.Spec.Rbac))

	if len(desired.Spec.Tags) > 0 {
		mc.Tags = *to.StringMapPtr(desired.Spec.Tags)
	}
}

func (mc *ManagedCluster) mergeNetwork(desired *resourcesv1alpha1.ManagedCluster) {
	if isEmptyNetworkSpec(&desired.Spec.Network) {
		if mc.NetworkProfile == nil {
			mc.NetworkProfile = &containerservice.NetworkProfile{}
		}

		if len(desired.Spec.Network.Plugin) > 0 {
			mc.NetworkProfile.NetworkPlugin = containerservice.NetworkPlugin(string(desired.Spec.Network.Plugin))
		}

		if len(desired.Spec.Network.Policy) > 0 {
			mc.NetworkProfile.NetworkPolicy = containerservice.NetworkPolicy(string(desired.Spec.Network.Policy))
		}

		if len(desired.Spec.Network.PodCidr) > 0 {
			mc.NetworkProfile.PodCidr = to.StringPtr(string(desired.Spec.Network.PodCidr))
		}

		if len(desired.Spec.Network.ServiceCidr) > 0 {
			mc.NetworkProfile.ServiceCidr = to.StringPtr(string(desired.Spec.Network.ServiceCidr))
		}

		if len(desired.Spec.Network.DNSServiceIP) > 0 {
			mc.NetworkProfile.DNSServiceIP = to.StringPtr(string(desired.Spec.Network.DNSServiceIP))
		}

		if len(desired.Spec.Network.DockerBridgeCidr) > 0 {
			mc.NetworkProfile.DockerBridgeCidr = to.StringPtr(string(desired.Spec.Network.DockerBridgeCidr))
		}
	}
}

func (mc *ManagedCluster) mergeIdentity(desired *resourcesv1alpha1.ManagedCluster, s *corev1.Secret) {
	if mc.ServicePrincipalProfile == nil {
		mc.ServicePrincipalProfile = &containerservice.ManagedClusterServicePrincipalProfile{}
	}

	if len(desired.Spec.Identity.ClientID) > 0 {
		mc.ServicePrincipalProfile.ClientID = to.StringPtr(string(desired.Spec.Identity.ClientID))
	}

	if s != nil {
		if password, ok := s.Data["password"]; ok {
			data, err := base64.StdEncoding.DecodeString(string(password))
			if err != nil {
				// TODO: log some error here
			}
			mc.ServicePrincipalProfile.Secret = to.StringPtr(string(data))
		}
	}
}

func (mc *ManagedCluster) mergeAddons(desired *resourcesv1alpha1.ManagedCluster) {
	for k, v := range desired.Spec.Addons {
		if mc.AddonProfiles == nil {
			mc.AddonProfiles = map[string]*containerservice.ManagedClusterAddonProfile{}
		}

		mc.AddonProfiles[k] = &containerservice.ManagedClusterAddonProfile{
			Enabled: to.BoolPtr(bool(v.Enabled)),
			Config:  *to.StringMapPtr(v.Config),
		}
	}

}

func (mc *ManagedCluster) mergeNodePools(desired *resourcesv1alpha1.ManagedCluster) {
	for n, v := range desired.Spec.NodePools {
		if mc.AgentPoolProfiles == nil {
			pools := make([]containerservice.ManagedClusterAgentPoolProfile, len(desired.Spec.NodePools))
			mc.AgentPoolProfiles = &pools
		}

		pool := containerservice.ManagedClusterAgentPoolProfile{}
		if len(v.Name) > 0 {
			pool.Name = to.StringPtr(string(v.Name))
		} else {
			pool.Name = to.StringPtr("default")
		}

		if v.Size > 0 {
			pool.Count = to.Int32Ptr(v.Size)
		} else {
			pool.Count = to.Int32Ptr(3)
		}

		if len(v.NodeSpec.VMSize) > 0 {
			pool.VMSize = containerservice.VMSizeTypes(string(v.NodeSpec.VMSize))
		} else {
			pool.VMSize = containerservice.StandardD2V2
		}

		if v.NodeSpec.OsDiskSizeGB > 0 {
			pool.OsDiskSizeGB = to.Int32Ptr(v.NodeSpec.OsDiskSizeGB)
		} else {
			pool.OsDiskSizeGB = to.Int32Ptr(30)
		}

		if len(v.NodeSpec.VnetSubnetID) > 0 {
			pool.VnetSubnetID = to.StringPtr(string(v.NodeSpec.VnetSubnetID))
		}

		if v.NodeSpec.MaxPods > 0 {
			pool.MaxPods = to.Int32Ptr(v.NodeSpec.MaxPods)
		} else {
			pool.MaxPods = to.Int32Ptr(110)
		}

		pool.OsType = containerservice.OSType("Linux")
		pool.StorageProfile = containerservice.StorageProfileTypes("ManagedDisks")

		(*mc.AgentPoolProfiles)[n] = pool
	}

	if mc.AgentPoolProfiles == nil {
		mc.AgentPoolProfiles = &[]containerservice.ManagedClusterAgentPoolProfile{
			{
				Count:          to.Int32Ptr(3),
				MaxPods:        to.Int32Ptr(110),
				Name:           to.StringPtr("default"),
				OsDiskSizeGB:   to.Int32Ptr(30),
				StorageProfile: containerservice.StorageProfileTypes("ManagedDisks"),
				VMSize:         containerservice.StandardD2V2,
				OsType:         containerservice.OSType("Linux"),
			},
		}
	}
}

func (mc *ManagedCluster) mergeLinuxConfig(desired *resourcesv1alpha1.ManagedCluster) {
	if mc.LinuxProfile == nil {
		mc.LinuxProfile = &containerservice.LinuxProfile{AdminUsername: to.StringPtr(defaultAdminUser)}
	}

	if len(desired.Spec.Linux.AdminUsername) > 0 {
		mc.LinuxProfile.AdminUsername = to.StringPtr(string(desired.Spec.Linux.AdminUsername))
	}

	for n, v := range desired.Spec.Linux.AuthorizedKeys {
		if mc.LinuxProfile.SSH == nil {
			mc.LinuxProfile.SSH = &containerservice.SSHConfiguration{}
			pubKeys := make([]containerservice.SSHPublicKey, len(desired.Spec.Linux.AuthorizedKeys))
			mc.LinuxProfile.SSH.PublicKeys = &pubKeys
		}

		(*mc.LinuxProfile.SSH.PublicKeys)[n].KeyData = to.StringPtr(string(v))
	}
}

func isEmptyNetworkSpec(ns *resourcesv1alpha1.NetworkSpec) bool {
	return len(ns.Plugin) > 0 || len(ns.Policy) > 0 || len(ns.PodCidr) > 0 || len(ns.ServiceCidr) > 0 || len(ns.DNSServiceIP) > 0 || len(ns.DockerBridgeCidr) > 0
}
