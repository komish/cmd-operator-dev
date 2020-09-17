# Running Tests

The included Makefile has a make target `test` which will run `go test`. However, the default tests will use EnvTest to test the operator, and additional binaries are required. Running the following make target will download binaries locally and run tests.

```shell
make envtest
```