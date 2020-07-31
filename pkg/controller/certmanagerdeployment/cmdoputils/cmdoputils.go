// Package cmdoputils contains utility functions for the certmanagerdeployment-operator.
// It should depend on as little as possible from the packages found in path
// github.com/komish/certmanager-operator/pkg/controller and further to prevent
// cyclical imports errors.
package cmdoputils

import (
	"encoding/json"
	"reflect"

	redhatv1alpha1 "github.com/komish/certmanager-operator/pkg/apis/redhat/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MergeMaps will take two maps (dest, addition) and merge all keys/values from "addition"
// into "dest". Keys from addition will supercede keys from "dest".
func MergeMaps(dest map[string]string, addition map[string]string) map[string]string {
	for k, v := range addition {
		dest[k] = v
	}
	return dest
}

// CertManagerVersionIsSupported returns true if the version of CertManagerDeployment custom resource
// is supported by this operator. An empty version field is always supported because it
// allows the operator to pick.
func CertManagerVersionIsSupported(cr *redhatv1alpha1.CertManagerDeployment, matrix map[string]bool) bool {
	vers := cr.Spec.Version
	// a nil version indicates that the CR didn't have Version set.
	if vers == nil {
		return true
	}

	_, ok := matrix[*vers]
	return ok
}

// GetSupportedCertManagerVersions returns a list of the versions of cert-manager supported by the operator.
// The supported versions are defined in
// github.com/komish/certmanager-operator/pkg/controllers/certmanagerdeployment/componentry
func GetSupportedCertManagerVersions(matrix map[string]bool) []string {
	versions := make([]string, len(matrix))
	for vers := range matrix {
		versions = append(versions, vers)
	}

	return versions
}

// CRVersionOrDefaultVersion accepts the version value from the CR spec and will do
// a check for nil. If nil, it will return the default value def.
func CRVersionOrDefaultVersion(cr *string, def string) string {
	if cr != nil {
		return *cr
	}

	return def
}

// GetStringPointer returns a string pointer to the input string
func GetStringPointer(str string) *string {
	return &str
}

// HasLabelOrAnnotationWithValue checks if the input map has the specified key with the specified value.
// Can be used to facilitate updates on objects where certain labels and annotations need to be in place.
func HasLabelOrAnnotationWithValue(in map[string]string, key, value string) bool {
	if val, ok := in[key]; ok {
		if val == value {
			return true
		}
	}

	return false
}

// LabelsAndAnnotationsMatch returns true if two objects that have ObjectMeta
// both have the same labels and annotations. In this case, dest object must
// have the same labels and annotations as the src object so it should be
// assumed the dest object might have more labels and annotations and this
// is acceptable so long as it has the same ones as the src.
func LabelsAndAnnotationsMatch(src, dest metav1.Object) bool {
	lblsMatch := true
	annotsMatch := true

	dLabels, dAnnots := dest.GetLabels(), dest.GetAnnotations()
	sLabels, sAnnots := src.GetLabels(), src.GetAnnotations()

	for k, v := range sLabels {
		if !HasLabelOrAnnotationWithValue(dLabels, k, v) {
			lblsMatch = false
		}
	}

	for k, v := range sAnnots {
		if !HasLabelOrAnnotationWithValue(dAnnots, k, v) {
			annotsMatch = false
		}
	}

	return lblsMatch && annotsMatch
}

// ObjectsMatch compares the JSON-form of two objects. The src object is considered to be
// the mold, which means that its keys and values must exist in the dest object for this
// to return true. The dest object can have additional keys and values, so long as the
// keys as defined by the src exist and have the same value. Input objects are expected
// be of the same type, or effectively be the same format when marshaled to JSON.
func ObjectsMatch(src, dest interface{}) bool {
	switch typedSrc := src.(type) {
	case map[string]interface{}:
		x := typedSrc
		y, ok := dest.(map[string]interface{})
		if !ok {
			return false
		}
		for k, v := range x {
			switch v.(type) {
			case string, float64, bool:
				if x[k] != y[k] {
					return false
				}
			case map[string]interface{}, []interface{}:
				if ok := ObjectsMatch(x[k], y[k]); !ok {
					return false
				}
			case nil:
			default:
				// we don't know what the input type is.
				return false
			}
		}
	case []interface{}:
		x := typedSrc
		y, ok := dest.([]interface{})
		if !ok {
			return false
		}
		for i, v := range x {
			switch v.(type) {
			case string, float64, bool:
				return reflect.DeepEqual(x, y)
			case map[string]interface{}:
				if ok := ObjectsMatch(x[i], y[i]); !ok {
					return false
				}
			case []interface{}:
				if ok := ObjectsMatch(x[i], y[i]); !ok {
					return false
				}
			default:
				return false
			}
		}
	case nil:
		// If the source type is nil, there's nothing to compare.
	default:
		// we don't know what the input type is.
		return false
	}

	return true
}

// Interfacer is a helper type that helps convert JSON-serializable objects
// to an interface{} of their respective JSON representation  by unmarshaling
// to an interface.
type Interfacer struct {
	Data interface{}
}

// ToJSONInterface converts data to an interface
func (i Interfacer) ToJSONInterface() (interface{}, error) {
	var iface interface{}
	b, err := json.Marshal(i.Data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &iface)
	if err != nil {
		return nil, err
	}
	return iface, nil
}
