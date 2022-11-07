FROM 795953128386.dkr.ecr.us-west-2.amazonaws.com/test-base-image:latest
COPY . .
RUN env GOOS=linux GOARCH=amd64 go build -o test ./examples/remote-test-runner/env.go
RUN chmod +x ./test
ENTRYPOINT ["./test"]