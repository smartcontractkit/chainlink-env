## How to run the same environment deployment inside k8s

You can build a `Dockerfile` to run exactly the same environment interactions inside k8s in case you need to run long-running tests
Base image is [here](Dockerfile.base)
```Dockerfile
FROM 795953128386.dkr.ecr.us-west-2.amazonaws.com/test-base-image:latest
COPY . .
RUN env GOOS=linux GOARCH=amd64 go build -o test ./examples/remote-test-runner/env.go
RUN chmod +x ./test
ENTRYPOINT ["./test"]
```
Build and upload it
```bash
./upload_test_image.sh 795953128386.dkr.ecr.us-west-2.amazonaws.com v1.1
```
Then run it
```bash
# all environment variables with a prefix TEST_ would be provided for k8s job
export TEST_ENV_VAR=myTestVarForAJob
# your image to run as a k8s job
export ENV_JOB_IMAGE="795953128386.dkr.ecr.us-west-2.amazonaws.com/core-integration-tests:v1.1"
# your example test file to run inside k8s
# if ENV_JOB_IMAGE is present chainlink-env will create a job, wait until it finished and get logs
go run examples/remote-test-runner/env.go
```