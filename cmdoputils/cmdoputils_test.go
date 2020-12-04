package cmdoputils

import (
	"reflect"
	"testing"

	operatorsv1alpha1 "github.com/komish/cmd-operator-dev/api/v1alpha1"
)

func TestMergeMaps(t *testing.T) {
	type testCase = []map[string]string
	testCases := []testCase{
		{ // first case
			{
				"foo": "bar", // dest
				"one": "two",
			},
			{
				"foo":   "baz", // addition
				"three": "four",
			},
			{
				"foo":   "baz", // result
				"one":   "two",
				"three": "four",
			},
		},
		{ // second case
			{}, // dest
			{
				"hello": "world", // addition
				"this":  "that",
			},
			{
				"this":  "that", // result
				"hello": "world",
			},
		},
	}

	for _, v := range testCases {
		res := MergeMaps(v[0], v[1])
		if !reflect.DeepEqual(res, v[2]) {
			t.Error("The test has failed!") // TODO fix
		}
	}
}

func TestCertManagerVersionIsSupported(t *testing.T) {
	vers := "v0.0.1"
	unsupportedVers := "v0.0.2"
	supportedVersions := map[string]bool{vers: true}

	type testCase struct {
		customResource    operatorsv1alpha1.CertManagerDeployment
		expectedToBeValid bool
	}

	testCases := []testCase{
		// no spec.Version means default version - this is a valid case
		{operatorsv1alpha1.CertManagerDeployment{Spec: operatorsv1alpha1.CertManagerDeploymentSpec{}}, true},
		// spec.Version as found in matrix is a valid case
		{operatorsv1alpha1.CertManagerDeployment{Spec: operatorsv1alpha1.CertManagerDeploymentSpec{Version: &vers}}, true},
		// spec.Version not found in the matrix is an invalid case.
		{operatorsv1alpha1.CertManagerDeployment{Spec: operatorsv1alpha1.CertManagerDeploymentSpec{Version: &unsupportedVers}}, false},
	}

	for _, c := range testCases {
		if res := CertManagerVersionIsSupported(&c.customResource, supportedVersions); res != c.expectedToBeValid {
			t.Errorf("unexpected result checking the validity of the version stored in the provided custom resource.\nGot:  %t\nWant: %t", res, c.expectedToBeValid)
			t.Logf("custom resource spec.Version: %v\n", *c.customResource.Spec.Version)
			t.Logf("supported version matrix: %v\n", supportedVersions)
		}
	}
}

func TestGetSupportedCertManagerVersions(t *testing.T) {
	// the boolean values are not important
	type testCase struct {
		matrix   map[string]bool
		expected []string
	}

	testCases := []testCase{
		{
			matrix: map[string]bool{
				"v0.0.1": true,
				"v0.0.2": true,
			},
			expected: []string{"v0.0.1", "v0.0.2"},
		},
		{
			matrix: map[string]bool{
				"foo": false,
				"bar": false,
			},
			expected: []string{"foo", "bar"},
		},
	}

	for _, c := range testCases {
		if actual := GetSupportedCertManagerVersions(c.matrix); !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("unexpected result generating slice of supported cert-manager versions.\nGot:  %v\nWant: %v\n", actual, c.expected)
			t.Logf("input matrix: %v", c.matrix)
		}
	}

}

func TestCRVersionOrDefaultVersion(t *testing.T) {
	t.Skip("Unimplemented")
}

func TestGetStringPointer(t *testing.T) {
	t.Skip("Unimplemented")
}

func TestHasLabelOrAnnotationsWithValue(t *testing.T) {
	t.Skip("Unimplemented")
}

func TestLabelsAndAnnotationsMatch(t *testing.T) {
	t.Skip("Unimplemented")
}

func TestObjectsMatch(t *testing.T) {
	t.Skip("Unimplemented")
}

func TestGetSortedStringSliceOf(t *testing.T) {
	t.Skip("Unimplemented")
}

func TestGetSortedFloat64SliceOf(t *testing.T) {
	t.Skip("Unimplemented")
}

// TODO: figure out how to test the type and its method
// func TestInterface(t *testing.T) {
//  	t.Skip("Unimplemented")
// }
