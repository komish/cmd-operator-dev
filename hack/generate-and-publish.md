# Generate and Publish

## Export Environment

These help inform the Makefile-specific variables. 

```bash
# the registry of choice
export REGISTRY="quay.io"

# Namespace in Container Registry
export REGISTRY_NAMESPACE="placeholder"

# Image of the controller
export OPERATOR_IMG="cmd-operator"

# Image of the operator bundle
export OPERATOR_BUNDLE_IMG="${OPERATOR_IMG}-bundle"

# Operator Release Version
export OPERATOR_VERSION="0.0.1"

# Index Image Name
export INDEX_IMAGE_NAME="${OPERATOR_IMG}-index"
```

## Generate Bundle and Make Bundle Container Image

From project directory, use `make` to generate the bundle components.

```bash
make bundle \
	IMG="${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_IMG}:v${OPERATOR_VERSION}" \
	VERSION="${OPERATOR_VERSION}" \
	CHANNELS="alpha" \
	DEFAULT_CHANNEL="alpha"
```

Having created the bundle on disk, next generate, tag, and publish the bundle image.

```bash
make bundle-build \
	BUNDLE_IMG="${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_BUNDLE_IMG}:v${OPERATOR_VERSION}"
```


## Generate and Publish Controller and Bundle Container Images

From project directory, use `make` to generate controller container image

```bash
make docker-build \
	IMG="${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_IMG}:v${OPERATOR_VERSION}"

# alternate if you want to skip the make test target
# docker build . -t "${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_IMG}:v${OPERATOR_VERSION}"

make docker-push \
	IMG="${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_IMG}:v${OPERATOR_VERSION}"
```

Do the same for the bundle container image.

```bash
make docker-push \
	IMG="${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_BUNDLE_IMG}:v${OPERATOR_VERSION}"
```

## Generate an index image using OPM (assumes docker as the container tool)

TODO: add a make target for this

```bash
opm index add \
    --bundles "${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_BUNDLE_IMG}:v${OPERATOR_VERSION}" \
    --tag "${REGISTRY}/${REGISTRY_NAMESPACE}/${INDEX_IMAGE_NAME}:v1.0.0" \
    --container-tool docker

make docker-push IMG="${REGISTRY}/${REGISTRY_NAMESPACE}/${INDEX_IMAGE_NAME}:v1.0.0"
```

## Use the index image in a cluster

This is just a matter of referring to the index image as a CatalogSource. Assumes OLM.

```bash
cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: ${OPERATOR_IMG}-catalog
  namespace: default
spec:
  sourceType: grpc
  image: ${REGISTRY}/${REGISTRY_NAMESPACE}/${INDEX_IMAGE_NAME}:v1.0.0
EOF
```


## Generate Package Manifests and Run

For backwards compatibility, use `make` to generate package manifests for use with OLM.

```bash
make packagemanifests \
	IMG="${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_IMG}:v${OPERATOR_VERSION}" \
	VERSION="${OPERATOR_VERSION}"

make run-packagemanifests VERSION="${OPERATOR_VERSION}"
```

### Clean up Package Manifests Execution

Once done testing via package manifests, clean up extra resources in the cluster. Namespace scope is assumed to match your current-context.

```bash
# the dashed version here will vary based on your version variable
oc delete subscription cmd-operator-dev-v0-0-1-sub

# the name of the catalog source may vary
oc delete catalogsource cmd-operator-dev-ocs

# This version will vary depending on your version variable
oc delete csv "cmd-operator-dev.v${OPERATOR_VERSION}"
```