# Running Tests

The included Makefile has a make target `test` which will run `go test`. However, the default tests will use EnvTest to test the operator, and additional binaries are required. Running the following make target will download binaries locally and run tests.

```shell
make envtest
```

# Local Testing

To run a quick set of tests, run the following

```bash
# make sure you have existing credentials in your environment and then 
# tell envtest that you intend to use that cluster.
export USE_EXISTING_CLUSTER=true

# start the controller in a separate window
make run

# tests don't currently clean anything up - clean up and start the tests.
./hack/bin/test-cleanup.sh && sleep 10; ./hack/bin/test-run.sh
```