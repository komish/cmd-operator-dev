package certmanagerdeployment

// mergeMaps will take two maps (dest, addition) and merge all keys/values from "addition"
// into "dest". Keys from addition will supercede keys from "dest".
func mergeMaps(dest map[string]string, addition map[string]string) map[string]string {
	for k, v := range addition {
		dest[k] = v
	}
	return dest
}
