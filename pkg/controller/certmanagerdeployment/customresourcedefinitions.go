package certmanagerdeployment

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/cmdoputils"
	"github.com/komish/certmanager-operator/pkg/controller/certmanagerdeployment/componentry"
	"gopkg.in/errgo.v2/fmt/errors"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
)

var (
	crdMap = map[string][]string{}
)

// GetCRDs returns a CustomResourceDefinitions for a given CertManagerDeployment
func (r *ResourceGetter) GetCRDs() ([]*apiextv1beta1.CustomResourceDefinition, error) {
	res := make([]*apiextv1beta1.CustomResourceDefinition, 0)

	// Get CertManager version from the r.CustomResource
	version := cmdoputils.CRVersionOrDefaultVersion(
		r.CustomResource.Spec.Version,
		componentry.CertManagerDefaultVersion)

	// Get file paths for the requested version
	crds, err := getCRDListForCertManagerVersion(version)
	if err != nil {
		return []*apiextv1beta1.CustomResourceDefinition{}, err
	}

	// Check that all files exist at the expected path.
	if !allFilesExist(version, crds) {
		return []*apiextv1beta1.CustomResourceDefinition{}, err
	}

	// Attempt to deserialize them all to CRD
	for _, crdPath := range crds {
		c, err := getCRDFromFile(crdPath)
		if err != nil {
			return []*apiextv1beta1.CustomResourceDefinition{}, err
		}

		res = append(res, c)
	}

	// Should have all CRDs here.
	return res, nil
}

// getCRDListForCertManagerVersion returns the CRDs for a requested version of cert-manager.
func getCRDListForCertManagerVersion(version string) ([]string, error) {
	switch version {
	case "v0.15.0", "v0.15.1":
		return []string{
			"cert-manager.io_issuers_crd.yaml",
			"cert-manager.io_certificates_crd.yaml",
			"cert-manager.io_certificaterequest_crd.yaml",
			"cert-manager.io_clusterissuers_crd.yaml",
			"acme.cert-manager.io_challenges_crd.yaml",
			"acme.cert-manager.io_orders_crd.yaml",
		}, nil
	}
	// TODO(): Support v0.14.x by adding case matches
	return []string{}, nil
}

// allFilesExist returns true if the files exist on disk at the specified path. Expected input
// is relative path, where pathPrefix is what precedes the filename and variable files contains a
// list of filenames including the extension.
func allFilesExist(pathPrefix string, files []string) bool {
	for _, file := range files {
		path := strings.Join([]string{pathPrefix, file}, "/")
		if _, err := os.Stat(path); err != nil {
			return false
		}
	}
	return true
}

// getCRDFromFile will read a CRD YAML file from disk and return the CRD as an object.
func getCRDFromFile(filePath string) (*apiextv1beta1.CustomResourceDefinition, error) {
	// get from disk
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		// some kind of error reading from disk
		// TODO(): better logging
		return nil, err
	}

	// decode to CRD object
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if err != nil {
		// some kind of error decoding the object to a CRD
		// TODO(): better logging
		return nil, err
	}

	// ensure we got a CustomResourceDefinition
	crd, ok := obj.(*apiextv1beta1.CustomResourceDefinition)
	if !ok {
		return nil, errors.Newf("Expected CustomResourceDefinition but got type %T from file at path %s", obj, filePath)
	}

	return crd, nil
}
