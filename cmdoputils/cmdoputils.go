// Package cmdoputils contains utility functions for the certmanagerdeployment-operator.
// It should depend on as little as possible from the packages found in path
// github.com/komish/cmd-operator-dev/controller/ and further to prevent
// cyclical imports errors.
package cmdoputils

import (
	"encoding/json"
	"reflect"
	"sort"

	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
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
func CertManagerVersionIsSupported(cr *operatorsv1alpha1.CertManagerDeployment, matrix map[string]bool) bool {
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
// github.com/komish/cmd-operator-dev/controllers/componentry
func GetSupportedCertManagerVersions(matrix map[string]bool) []string {
	versions := make([]string, len(matrix))
	i := 0
	for vers := range matrix {
		versions[i] = vers
		i++
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

// ObjectsMatch compares the JSON-form of two objects. The src object is considered to be
// the mold, which means that its keys and values must exist in the dest object for this
// to return true. The dest object can have additional keys and values, so long as the
// keys as defined by the src exist and have the same value. Input objects are expected
// be of the same type, or effectively be the same format when marshaled to JSON.
func ObjectsMatch(src, dest interface{}) bool {
	// TODO: This could use some logging.
	switch typedSrc := src.(type) {
	case map[string]interface{}:
		// if x is a map[string]interface{}, y should also be
		x := typedSrc
		// convert y, or we assume they're not the same if it fails
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
		// if x is a []interface{}, y should also be
		y, ok := dest.([]interface{})
		// convert y, or we assume they're not the same if it fails
		if !ok {
			return false
		}

		// if y is not the same length as x, then it's not possible
		// for it to be the same so bail out early.
		// TODO: evaluate if we need to allow slices that are larger
		// in the dest value (y). This would allow for x to be smaller
		// but y contains all x values. If we do allow for this, we
		// need to make it optional - because some things need to be
		// evaluated for exactness. Might be worth introducing a "strictness" for
		// the slice comparisons.
		if len(y) != len(x) {
			return false
		}

		for i, v := range x {
			switch v.(type) {
			case string:
				// this needs to be sorted before we can DeepEqual
				sx := getSortedStringSliceOf(x)
				sy := getSortedStringSliceOf(y)
				return reflect.DeepEqual(sx, sy)
			case float64:
				sx := getSortedFloat64SliceOf(x)
				sy := getSortedFloat64SliceOf(y)
				return reflect.DeepEqual(sx, sy)
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

// ToJSONInterface takes the input Interfacer.Data, converts it to JSON, and then unmarshals
// to an interface{} type. This gives the return interface the same data represented as a subset
// of relatively easy-to-compare
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

func getSortedStringSliceOf(data []interface{}) []string {
	res := make([]string, len(data))
	for i, v := range data {
		res[i] = v.(string)
	}

	sort.Strings(res)
	return res
}

func getSortedFloat64SliceOf(data []interface{}) []float64 {
	res := make([]float64, len(data))
	for i, v := range data {
		res[i] = v.(float64)
	}

	sort.Float64s(res)
	return res
}
