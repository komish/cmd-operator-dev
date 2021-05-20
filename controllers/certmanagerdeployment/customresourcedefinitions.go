package certmanagerdeployment

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
	"github.com/komish/cmd-operator-dev/cmdoputils"
	"github.com/komish/cmd-operator-dev/controllers/componentry"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// reconcileCRDs reconciles CustomResourceDefinition resource(s) for a given CertManagerDeployment resource.
func (r *CertManagerDeploymentReconciler) reconcileCRDs(instance *operatorsv1alpha1.CertManagerDeployment, reqLogger logr.Logger) error {

	reqLogger.Info("Starting reconciliation: CRDs")
	defer reqLogger.Info("Ending reconciliation: CRDs")

	crds, err := GetCRDsFor(*instance)

	if err != nil {
		reqLogger.Error(err, "Failed to get CRDs")
		// Something happened when trying to get CRDs for this reconciliation
		return err
	}

	for _, crd := range crds {
		found := &apiextv1.CustomResourceDefinition{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: crd.GetName()}, found)
		if err != nil && apierrors.IsNotFound(err) {
			reqLogger.Info("Creating CustomResourceDefinition", "CustomResourceDefinition.Name", crd.GetName())
			r.Eventf(instance, createManagedCRD.etype, createManagedCRD.reason, "%s: %s", createManagedCRD.message, crd.GetName())

			if err := r.Create(context.TODO(), crd); err != nil {
				return err
			}

			// successful create. Move onto next iteration.
			continue
		} else if err != nil {
			return err
		}

		// If needed, update CRD.
		genSpecInterface, err := cmdoputils.Interfacer{Data: crd.Spec}.ToJSONInterface()
		if err != nil { // error indicates marshaling problems
			return err
		}
		foundSpecInterface, err := cmdoputils.Interfacer{Data: found.Spec}.ToJSONInterface()
		if err != nil {
			return err
		}

		genLabelsInterface, err := cmdoputils.Interfacer{Data: crd.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		foundLabelsInterface, err := cmdoputils.Interfacer{Data: found.Labels}.ToJSONInterface()
		if err != nil {
			return err
		}

		genAnnotsInterface, err := cmdoputils.Interfacer{Data: crd.Annotations}.ToJSONInterface()
		if err != nil {
			return err
		}

		foundAnnotsInterface, err := cmdoputils.Interfacer{Data: found.Annotations}.ToJSONInterface()
		if err != nil {
			return err
		}

		// Check for equality
		specsMatch := cmdoputils.ObjectsMatch(genSpecInterface, foundSpecInterface)
		labelsMatch := cmdoputils.ObjectsMatch(genLabelsInterface, foundLabelsInterface)
		annotsMatch := cmdoputils.ObjectsMatch(genAnnotsInterface, foundAnnotsInterface)

		// If not equal, update.
		if !(specsMatch && labelsMatch && annotsMatch) {
			reqLogger.Info("CustomResourceDefinition already exists, but needs an update. Updating.",
				"CustomResourceDefinition.Name", crd.GetName(),
				"HasExpectedLabels", labelsMatch,
				"HasExpectedAnnotations", annotsMatch,
				"HasExpectedSpec", specsMatch)
			r.Eventf(instance, updatingManagedCRD.etype, updatingManagedCRD.reason, "%s: %s", updatingManagedCRD.message, crd.GetName())

			// modify the state of the old object to post to API
			updated := found.DeepCopy()

			if !specsMatch {
				updated.Spec = crd.Spec
			}

			if !labelsMatch {
				updated.ObjectMeta.Labels = crd.GetLabels()
			}

			if !annotsMatch {
				updated.ObjectMeta.Annotations = crd.GetAnnotations()
			}

			reqLogger.Info("Updating CustomResourceDefinition.", "CustomResourceDefinition.Name", crd.GetName())
			if err := r.Update(context.TODO(), updated); err != nil {
				// some issue performing the update.
				return err
			}

			r.Eventf(instance, updatedManagedCRD.etype, updatedManagedCRD.reason, "%s: %s", updatedManagedCRD.message, crd.GetName())
		}
	}

	return nil
}

// GetCRDsFor returns CustomResourceDefinitions for a given CertManagerDeployment.
func GetCRDsFor(cr operatorsv1alpha1.CertManagerDeployment) ([]*apiextv1.CustomResourceDefinition, error) {
	// The managed CRD representations are coming directly from YAML files
	// for a given release. These YAMLs are released by the cert-manager
	// project for each release of the application to ensure compatibility
	// with the upstream project.
	//
	// A directory should exist in the operator image for each supported release.
	// Each release directory contains YAMLs for each custom resource definition
	// supported by that release.

	res := make([]*apiextv1.CustomResourceDefinition, 0)

	version := cmdoputils.CRVersionOrDefaultVersion(
		cr.Spec.Version,
		componentry.CertManagerDefaultVersion)

	// get the file paths for the version of cert-manager requested.
	crds, err := getCRDListForCertManagerVersion(version)
	if err != nil {
		return []*apiextv1.CustomResourceDefinition{}, err
	}

	// check that all files exist at the given path.
	if ok, missing := allFilesExist(crds); !ok {
		return []*apiextv1.CustomResourceDefinition{}, fmt.Errorf("unable to find CRDs for version %s. Missing %s", version, missing)
	}

	// deserialize to struct. Only crd @ v1 is supported.
	for _, crdPath := range crds {
		c, err := getCRDFromFile(crdPath)
		if err != nil {
			return []*apiextv1.CustomResourceDefinition{}, err
		}

		res = append(res, c)
	}

	return res, nil
}

// getCRDListForCertManagerVersion returns the CRDs for a requested version of cert-manager.
func getCRDListForCertManagerVersion(version string) ([]string, error) {
	switch version {
	case "v1.3.1", "v1.3.0", "v1.2.0":
		return addPathPrefixToPathList(version, []string{
			"cert-manager.io_issuers_crd.yaml",
			"cert-manager.io_certificates_crd.yaml",
			"cert-manager.io_certificaterequests_crd.yaml",
			"cert-manager.io_clusterissuers_crd.yaml",
			"acme.cert-manager.io_challenges_crd.yaml",
			"acme.cert-manager.io_orders_crd.yaml",
		}), nil
	default:
		// sanity check / precuation.
		// this case should never be hit because the reconciler
		// logic should halt reconciliation if the supported version
		// reflected in the cr is incorrect.
		return []string{}, errors.New("requested version is unsupported by this operator")
	}
}

// allFilesExist returns true if the files exist on disk at the specified path.
// Path format is typically vX.Y.Z/filename.yaml
func allFilesExist(files []string) (bool, string) {
	for _, file := range files {

		if _, err := os.Stat(file); err != nil {
			return false, file
		}
	}
	return true, ""
}

// getCRDFromFile will read a CRD YAML file from disk and return the CRD as an object.
func getCRDFromFile(filePath string) (*apiextv1.CustomResourceDefinition, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		// some kind of error reading from disk
		return nil, err
	}

	// decode to CRD object
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if err != nil {
		// some kind of error decoding the object to a CRD
		return nil, err
	}

	// ensure we got a CustomResourceDefinition
	crd, ok := obj.(*apiextv1.CustomResourceDefinition)
	if !ok {
		return nil, fmt.Errorf("expected CustomResourceDefinition but got type %T from file at path %s", obj, filePath)
	}

	return crd, nil
}

// addPathPrefixToPathList prepends each path in paths with the prefix pathPrefix.
func addPathPrefixToPathList(pathPrefix string, paths []string) []string {
	new := make([]string, 0)
	for _, p := range paths {
		new = append(new, path.Join(crdPathOrWD(), pathPrefix, p))
	}
	return new
}

// crdPathOrWD returns the path where the CRDs should be found or the current working directory
// for the binary.
func crdPathOrWD() string {
	// TODO: handle this error
	dir, _ := os.Getwd()
	return dir
}
