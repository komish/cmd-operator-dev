package cmdoputils

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/komish/cmd-operator-dev/tests/fixtures"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"

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
			},
			expected: []string{"v0.0.1"},
		},
		{
			matrix: map[string]bool{
				"foo": false,
			},
			expected: []string{"foo"},
		},
	}

	for _, c := range testCases {
		actual := GetSupportedCertManagerVersions(c.matrix)
		sort.Strings(actual)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("unexpected result generating slice of supported cert-manager versions.\nGot:  %v\nWant: %v\n", actual, c.expected)
			t.Logf("input matrix: %v", c.matrix)
		}
	}

}

func TestCRVersionOrDefaultVersion(t *testing.T) {
	defaultValue := "v0.0.0"
	crValue := "v0.0.1"

	type testCase struct {
		crValue  *string
		expected string
	}

	testCases := []testCase{
		{
			crValue:  nil,
			expected: defaultValue,
		},
		{
			crValue:  &crValue,
			expected: crValue,
		},
	}

	for _, c := range testCases {
		if actual := CRVersionOrDefaultVersion(c.crValue, defaultValue); actual != c.expected {
			t.Errorf("unexpected result determining if the custom resource version or default version should be used.\nGot:  %v\nWant: %v\n", actual, c.expected)
			t.Logf("input custom resource value: %v", c.crValue)
		}
	}
}

func TestGetStringPointer(t *testing.T) {
	in := "foo"
	expected := &in

	if actual := GetStringPointer(in); *actual != *expected {
		t.Errorf("unexpected result returning a string pointer of a given input string. \nGot:  %v\nWant: %v\n", actual, expected)
		t.Logf("input string: %v", in)
	}
}

func TestHasLabelOrAnnotationWithValue(t *testing.T) {
	type testCase struct {
		searchKey   string
		expectedVal string
		expected    bool
	}

	inputMap := map[string]string{
		"foo":   "bar",
		"hello": "world",
	}

	testCases := []testCase{
		{ // return true if the search key exists with the expected value
			searchKey:   "foo",
			expectedVal: "bar",
			expected:    true,
		},
		{ // return false if the search key exists but with some other value
			searchKey:   "hello",
			expectedVal: "mundo",
			expected:    false,
		},
		{ // return false if the search key is missing
			searchKey:   "this",
			expectedVal: "that",
			expected:    false,
		},
	}

	for _, c := range testCases {
		if actual := HasLabelOrAnnotationWithValue(inputMap, c.searchKey, c.expectedVal); actual != c.expected {
			t.Errorf("unexpected result determining if the input map had the search key with the expected value.\nGot:  %v\nWant: %v\n", actual, c.expected)
			t.Logf("input map: %v", inputMap)
			t.Logf("input search key: %v", c.searchKey)
			t.Logf("input expected value: %v", c.expectedVal)
		}
	}
}

func TestObjectsMatch(t *testing.T) {
	type testCase struct {
		generated interface{}
		persisted interface{}
		expected  bool
	}

	var persistedPod corev1.Pod
	if err := json.Unmarshal(fixtures.PersistedCoreV1PodJSON, &persistedPod); err != nil {
		t.Errorf("error attempting to deserialize fixture")
		t.Logf("error msg: %s", err)
	}

	correctLabels := map[string]string{"app": "zhack"}
	incorrectLabels := map[string]string{"app": "foo"}

	testCases := []testCase{
		{ // the persisted labels and the comparison labels match
			generated: correctLabels,
			persisted: persistedPod.Labels,
			expected:  true,
		},
		{ // the persisted labels and the comparison labels do not match
			generated: incorrectLabels,
			persisted: persistedPod.Labels,
			expected:  false,
		},
		{ // the arbitrary object variation should match the base object
			generated: fixtures.ComplexObjectVariationButPass,
			persisted: fixtures.ComplexObject,
			expected:  true,
		},
		{ // the arbitrary object variation should not match the base object
			generated: fixtures.ComplexObjectVariationButFail,
			persisted: fixtures.ComplexObject,
			expected:  false,
		},
	}

	for _, c := range testCases {
		genInterface, err := Interfacer{Data: c.generated}.ToJSONInterface()
		if err != nil {
			t.Errorf("error attempting to obtain a JSON interface of the \"generated\" pod")
			t.Logf("error msg: %s", err)
		}
		persistedInterface, err := Interfacer{Data: c.persisted}.ToJSONInterface()
		if err != nil {
			t.Errorf("error attempting to obtain a JSON interface of the \"persisted\" pod")
			t.Logf("error msg: %s", err)
		}

		if actual := ObjectsMatch(genInterface, persistedInterface); actual != c.expected {
			t.Errorf("unexpected result attempting to check to see if the interfaced objects match\nGot:  %v\nWant: %v\n", actual, c.expected)
			t.Logf("generated: %s", c.generated)
			t.Logf("persisted: %s", c.persisted)
		}
	}
}

func TestGetSortedStringSliceOf(t *testing.T) {
	t.Skip("Unimplemented")
}

func TestGetSortedFloat64SliceOf(t *testing.T) {
	t.Skip("Unimplemented")
}

func TestInterfacer(t *testing.T) {
	type testCase = Interfacer

	// We expect most regular objects to successfully go through this marshal process.
	testCases := []testCase{
		{
			Data: corev1.Pod{},
		},
		{
			Data: metav1.ObjectMeta{},
		},
		{
			Data: operatorsv1alpha1.CertManagerDeploymentSpec{},
		},
	}

	for _, c := range testCases {
		i := Interfacer{Data: c.Data}
		if _, err := i.ToJSONInterface(); err != nil {
			t.Errorf("unexpected result converting a data blob to a JSON representation in an interface.\nGot:  %v\nWant: No Errors\n", err)
			t.Logf("input data blob: %v", c.Data)
		}
	}
}
