package componentry

const (
	// CertManagerDefaultVersion is the version of CertManager that the
	// operator will install by default.
	CertManagerDefaultVersion = "v1.3.1"

	// CertManagerBaseName is the base name to use for objects that need to
	// include the name in their object names.
	CertManagerBaseName string = "cert-manager"

	// CertManagerDeploymentNamespace is the namespace that is used to deploy
	// namespaced resources (e.g. serviceaccounts, roles) used by the cert-manager
	// controllers.
	CertManagerDeploymentNamespace string = CertManagerBaseName
)
