package fixtures

import (
	"time"
)

const (
	Namespace        = "test-cmd-operator"
	Certificate      = "cmd-operator-tests-cert"
	Secret           = "cmd-operator-tests-cert"
	Timeout          = time.Second * 60
	Duration         = time.Second * 60
	Interval         = time.Second * 2
	TestResourceName = "test-cmd-operator-pod-refresher"
)

type complexStruct struct {
	SomeSlice  []string          `json:"someSlice"`
	SomeMap    map[string]string `json:"someMap"`
	SomeStruct complexSubStruct  `json:"SomeStruct"`
}

type complexSubStruct struct {
	Numbers []int32 `json:"numbers"`
}

var (
	// PreviousSupportedVersion is the previous supported Y version as identified
	// by a semantic version string and is used throughout tests as needed. This is
	// expected to increment as the default/supported version increments.
	PreviousSupportedVersion = "v1.2.0"

	// ComplexObject represents a base object for testing the ObjectsMatch comparison logic.
	ComplexObject = complexStruct{
		SomeSlice: []string{"fo", "fum"},
		SomeMap: map[string]string{
			"this":  "that",
			"hello": "world",
		},
		SomeStruct: complexSubStruct{
			Numbers: []int32{0, 1, 2},
		},
	}

	// ComplexObjectVariationButPass represents the base object ComplexObject but with what is expected to be
	// a safe variation. In other words, this should match the base object because the base object has everything
	// represented in this variation, and potentially more (which is ok).
	ComplexObjectVariationButPass = complexStruct{
		SomeSlice: []string{"fo", "fum"},
		SomeMap: map[string]string{
			"hello": "world",
		},
		SomeStruct: complexSubStruct{
			Numbers: []int32{0, 1, 2},
		},
	}

	// ComplexObjectVariationButFail represents the base object ComplexObject but with what is expected to be
	// an unsafe variation. In other words, this should not match the base object because this object has an additional
	// value that is not represented in the base.
	ComplexObjectVariationButFail = complexStruct{
		SomeSlice: []string{"fo", "fum"},
		SomeMap: map[string]string{
			"this":  "that",
			"hello": "world",
			"foo":   "bar",
		},
		SomeStruct: complexSubStruct{
			Numbers: []int32{0, 1, 2},
		},
	}

	// PersistedCoreV1PodJSON is a representation of a pod that would be received by a GET request
	// to the API for use in testing.
	PersistedCoreV1PodJSON = []byte(
		`{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {
        "annotations": {
            "kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"annotations\":{},\"labels\":{\"app\":\"zhack\"},\"name\":\"infinity\",\"namespace\":\"default\"},\"spec\":{\"containers\":[{\"command\":[\"/bin/sleep\",\"infinity\"],\"image\":\"ubuntu:latest\",\"name\":\"infinity\"}]}}\n"
        },
        "creationTimestamp": "2020-12-07T21:07:24Z",
        "labels": {
            "app": "zhack"
        },
        "name": "infinity",
        "namespace": "default",
        "resourceVersion": "1224325",
        "selfLink": "/api/v1/namespaces/default/pods/infinity",
        "uid": "f15fe383-b850-4492-ada5-6ac452059959"
    },
    "spec": {
        "containers": [
            {
                "command": [
                    "/bin/sleep",
                    "infinity"
                ],
                "image": "ubuntu:latest",
                "imagePullPolicy": "Always",
                "name": "infinity",
                "resources": {},
                "terminationMessagePath": "/dev/termination-log",
                "terminationMessagePolicy": "File",
                "volumeMounts": [
                    {
                        "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
                        "name": "default-token-jj6kk",
                        "readOnly": true
                    }
                ]
            }
        ],
        "dnsPolicy": "ClusterFirst",
        "enableServiceLinks": true,
        "imagePullSecrets": [
            {
                "name": "default-dockercfg-cdr4k"
            }
        ],
        "nodeName": "crc-j55b9-master-0",
        "priority": 0,
        "restartPolicy": "Always",
        "schedulerName": "default-scheduler",
        "securityContext": {},
        "serviceAccount": "default",
        "serviceAccountName": "default",
        "terminationGracePeriodSeconds": 30,
        "tolerations": [
            {
                "effect": "NoExecute",
                "key": "node.kubernetes.io/not-ready",
                "operator": "Exists",
                "tolerationSeconds": 300
            },
            {
                "effect": "NoExecute",
                "key": "node.kubernetes.io/unreachable",
                "operator": "Exists",
                "tolerationSeconds": 300
            }
        ],
        "volumes": [
            {
                "name": "default-token-jj6kk",
                "secret": {
                    "defaultMode": 420,
                    "secretName": "default-token-jj6kk"
                }
            }
        ]
    },
    "status": {
        "conditions": [
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2020-12-07T21:07:24Z",
                "status": "True",
                "type": "Initialized"
            },
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2020-12-07T21:07:36Z",
                "status": "True",
                "type": "Ready"
            },
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2020-12-07T21:07:36Z",
                "status": "True",
                "type": "ContainersReady"
            },
            {
                "lastProbeTime": null,
                "lastTransitionTime": "2020-12-07T21:07:24Z",
                "status": "True",
                "type": "PodScheduled"
            }
        ],
        "containerStatuses": [
            {
                "containerID": "cri-o://e159fb9a1008efecd0908654f42bc3ad3f24c5ca2224f94bf6812238e5250154",
                "image": "docker.io/library/ubuntu:latest",
                "imageID": "docker.io/library/ubuntu@sha256:4e4bc990609ed865e07afc8427c30ffdddca5153fd4e82c20d8f0783a291e241",
                "lastState": {},
                "name": "infinity",
                "ready": true,
                "restartCount": 0,
                "started": true,
                "state": {
                    "running": {
                        "startedAt": "2020-12-07T21:07:36Z"
                    }
                }
            }
        ],
        "hostIP": "192.168.126.11",
        "phase": "Running",
        "podIP": "10.116.0.45",
        "podIPs": [
            {
                "ip": "10.116.0.45"
            }
        ],
        "qosClass": "BestEffort",
        "startTime": "2020-12-07T21:07:24Z"
    }
}`)
)
