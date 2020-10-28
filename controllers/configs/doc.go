// Package configs provides a series of helpers allowing a user to authoritatively unmarshal a JSON
// representation of a cert-manager version's commandline arguments into a struct. This is used to help
// ensure declarative representations of the commandline arguments can be validated against a stable
// representation of possible values for a given cert-manager binary's version.
//
// Cert-Manager binaries contain a series of configuration options, ingested via commandline flags,
// that change the controller's execution. These configuration options may change per version of
// cert-manager that's released. As a result, the configs package will provide a stable way to access
// the struct-representation of those flags per versions of cert-manager supported by the CertManagerDeployment
// operator.
//
// These structs fulfill the runtime.Object interface, but they are never intended to persist into the Kubernetes
// API for any reason.
//
// As a result, the GVK of this "pseudo" runtime.Object should remain static across various versions of types
// that might exist in this configs package and any subpackages associated with configs.
//
// It's also possible that the default flags passed to the various cert-manager components might change while the
// underlying types do not. For this reason, the defaults and types are separate packages and can be advanced independently.
//
// In addition, while the subpackages may be tied to a version in their import path, it's important to note that multiple
// versions of cert-manager can be associated with another version's structs. The version represented in the import path
// is relative to the version of cert-manager where the configuration type definition changed. It's possible that the configuration
// structs were used across n+1, n+2, ... and as such getters may return config structs that reference older versions.
package configs
