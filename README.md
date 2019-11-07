# ctrlarm

## Overview
ctrlarm uses [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) libraries and tools to create Kubernetes APIs for managing Azure resources.

- Azure Kubernetes Service (preview)
- _More to come..._

## Prerequisites
- [kustomize](https://github.com/kubernetes-sigs/kustomize)
- [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

## Quickstart
``` bash
export AZURE_SUBSCRIPTION_ID="$(az account show | jq -j -r '.id')"
export AZURE_AUTH_LOCATION="${HOME}/creds.json"

# Create Service Principal used by the ctrlarm controller
az ad sp create-for-rbac --sdk-auth \
    --role "Contributor" \
    --scope "/subscriptions/${AZURE_SUBSCRIPTION_ID}" > "${AZURE_AUTH_LOCATION}"

export AZURE_B64ENCODED_CREDENTIALS="$(cat ${AZURE_AUTH_LOCATION} | base64 -w0)"
export AZURE_CLIENT_ID="$(cat ${AZURE_AUTH_LOCATION} | jq -j -r '.clientId' | base64 -w0)"
export AZURE_CLIENT_SECRET="$(cat ${AZURE_AUTH_LOCATION} | jq -j -r '.clientSecret' | base64 -w0)"

# Replace "your-docker-registry"
export IMG=your-docker-registry/ctrlarm-controller:latest

# Build and Deploy
make docker-build docker-push install deploy

# Provision an AKS cluster
cat config/samples/azure_v1alpha1_managedcluster.yaml | envsubst | kubectl apply -f -
```
