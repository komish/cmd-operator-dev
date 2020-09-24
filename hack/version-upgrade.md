On new cert-manager release:

* Diff previous version's YAML manifests to upcoming support version's CRDs (from cert-manager release assets)
* Create a new top-level directory containing cert-manager CRDs and rename `managed-by` label value.
* Update `SupportedVersions` Map in **controllers/componentry/componentry.go**
* Update `CertManagerDefaultVersion` Variable in **controllers/componentry/constants.go**
* Update `GetComponentFor*` Func for each component.
  * Change defaults
  * Transition adjustments for previous supported default version.
* Update `getCRDListForCertManagerVersion` Func in **controllers/customresourcedefinitions.go**
* Update CertManagerDeployment type markers such that they reflect new supported versions in **api/v1alpha1/certmanagerdeployment_types.go**.