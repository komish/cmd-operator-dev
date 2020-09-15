# Bundle Generation

## Export Environment

### Project Variables

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