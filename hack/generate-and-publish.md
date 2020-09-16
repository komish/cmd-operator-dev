# Generate and Publish

## Export Environment

These help inform the Makefile-specific variables. 

```shell
# the registry of choice
export REGISTRY="quay.io"

# Namespace in Container Registry
export REGISTRY_NAMESPACE="placeholder"

# Image of the controller
export OPERATOR_IMG="cmd-operator"
export OPERATOR_VERSION="0.0.1"
```

## Generate Bundle

From project directory, use `make` to generate the bundle components.

```shell
make bundle \
	IMG="${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_IMG}:v${OPERATOR_VERSION}" \
	VERSION="${OPERATOR_VERSION}" \
	CHANNELS="alpha" \
	DEFAULT_CHANNEL="alpha"
```

## Generate and Publish Controller Container Image

From project directory, use `make` to generate controller container image

```shell
make docker-build \
	IMG="${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_IMG}:v${OPERATOR_VERSION}"

# alternate if you want to skip the make test target
# docker build . -t "${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_IMG}:v${OPERATOR_VERSION}"

make docker-push \
	IMG="${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_IMG}:v${OPERATOR_VERSION}"
```

## Generate Package Manifests and Run

For backwards compatibility, use `make` to generate package manifests for use with OLM.

```shell
make packagemanifests \
	IMG="${REGISTRY}/${REGISTRY_NAMESPACE}/${OPERATOR_IMG}:v${OPERATOR_VERSION}" \
	VERSION="${OPERATOR_VERSION}"

make run-packagemanifests VERSION="${OPERATOR_VERSION}"
```

## Clean up Package Manifests Execution

Once done testing via package manifests, clean up extra resources in the cluster. Namespace scope is assumed to match your current-context.

```shell
# the dashed version here will vary based on your version variable
oc delete subscription cmd-operator-dev-v0-0-1-sub

# the name of the catalog source may vary
oc delete catalogsource cmd-operator-dev-ocs

# This version will vary depending on your version variable
oc delete csv "cmd-operator-dev.v${OPERATOR_VERSION}"
```