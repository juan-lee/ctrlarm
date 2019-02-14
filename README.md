# ctrlarm

## Overview
ctrlarm uses (kubebuilder)[https://github.com/kubernetes-sigs/kubebuilder] libraries and tools to create Kubernetes APIs for managing Azure resources.

- Azure Kubernetes Service (preview)
- CosmosDB (preview)
- _More to come..._

## Quickstart

### Prerequisites
- Install (dep)[https://github.com/golang/dep]
- Install (kustomize)[https://github.com/kubernetes-sigs/kustomize]
- Install (kubebuilder)[https://github.com/kubernetes-sigs/kubebuilder]
- Install and Start (minikube)[https://github.com/kubernetes/minikube]

### Customize Sample
Update config/samples/cosmosdb_v1alpha1_databaseaccount.yaml with your own azure subscription ID and resource group.
```
apiVersion: cosmosdb.azure.com/v1alpha1
kind: DatabaseAccount
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: databaseaccount-sample
spec:
  subscriptionID: <***insert sub id***>
  resourceGroup: <***insert resource group***>
  kind: "MongoDB"
  location: westus2
```

### Create a Sample Resource
```
export AZURE_ENVIRONMENT=AzurePublicCloud
export AZURE_SUBSCRIPTION_ID=<*** insert subscription id ***>
export AZURE_TENANT_ID=<*** insert tenant id ***>
export AZURE_CLIENT_ID=<*** insert client id ***>
export AZURE_CLIENT_SECRET=<*** insert client secret ***>

make install
make run
kubectl apply -f config/samples/cosmosdb_v1alpha1_databaseaccount.yaml
```

