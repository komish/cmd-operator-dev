package fixtures

var (
	ComplexObject = []byte(`
	{
		"foo": [
			"bar",
			"baz",
			0,
			1,
			{
				"hello": "world",
				"hola": "mundo"
			}
		]
	}`)

	ComplexObjectVariationButPass = []byte(`
	{
		"foo": [
			"bar",
			0,
			1,
			{
				"hello": "world",
				"hola": "mundo"
			}
		]
	}`)

	ComplexObjectVariationButFail = []byte(`
	{
		"foo": [
			"bar",
			"baz",
			0,
			1,
			{
				"hello": "mundo",
				"hola": "mundo"
			}
		]
	}`)
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
