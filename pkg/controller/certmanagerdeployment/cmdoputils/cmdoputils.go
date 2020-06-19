// Package cmdoputils contains utility functions for the certmanagerdeployment-operator.
package cmdoputils

// MergeMaps will take two maps (dest, addition) and merge all keys/values from "addition"
// into "dest". Keys from addition will supercede keys from "dest".
func MergeMaps(dest map[string]string, addition map[string]string) map[string]string {
	for k, v := range addition {
		dest[k] = v
	}
	return dest
}
