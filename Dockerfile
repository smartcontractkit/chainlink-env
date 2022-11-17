ARG BASE_IMAGE=795953128386.dkr.ecr.us-west-2.amazonaws.com/test-base-image:v5.0
FROM $BASE_IMAGE
COPY . testdir/
WORKDIR testdir
RUN go build -o test examples/remote-test-runner/env.go
RUN chmod +x ./test
ENTRYPOINT ["./test"]