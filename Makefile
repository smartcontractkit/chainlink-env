BIN_DIR = bin
export GOPATH ?= $(shell go env GOPATH)
export GO111MODULE ?= on

.PHONY: lint
lint:
	${BIN_DIR}/golangci-lint --color=always run -v

.PHONY: golangci
golangci:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${BIN_DIR} v1.46.2

.PHONY: install_cli
install: # installs CLI wizard
	go install cmd/wizard/chainlink-env.go

.PHONY: install_deps, golangci
install_deps: golangci
	yarn global add cdk8s-cli@2.0.0-rc.1
	curl -LO https://dl.k8s.io/release/v1.24.0/bin/darwin/amd64/kubectl
	chmod +x ./kubectl
	mv kubectl ./bin
	curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
	cdk8s import

.PHONY: test
test:
	go test ./...

.PHONY: examples
examples:
	go run cmd/test.go

.PHONY: chaosmesh
chaosmesh: ## there is currently a bug on JS side to import all CRDs from one yaml file, also a bug with stdin, so using cluster directly trough file
	kubectl get crd networkchaos.chaos-mesh.org -o json > tmp.json && cdk8s import -o imports/k8s/networkchaos tmp.json
	kubectl get crd stresschaos.chaos-mesh.org -o json > tmp.json && cdk8s import -o imports/k8s/stresschaos tmp.json
	kubectl get crd timechaos.chaos-mesh.org -o json > tmp.json && cdk8s import -o imports/k8s/timechaos tmp.json
	kubectl get crd podchaos.chaos-mesh.org -o json > tmp.json && cdk8s import -o imports/k8s/podchaos tmp.json
	kubectl get crd podiochaos.chaos-mesh.org -o json > tmp.json && cdk8s import -o imports/k8s/podiochaos tmp.json
	kubectl get crd httpchaos.chaos-mesh.org -o json > tmp.json && cdk8s import -o imports/k8s/httpchaos tmp.json
	kubectl get crd iochaos.chaos-mesh.org -o json > tmp.json && cdk8s import -o imports/k8s/iochaos tmp.json
	kubectl get crd podnetworkchaos.chaos-mesh.org -o json > tmp.json && cdk8s import -o imports/k8s/podnetworkchaos tmp.json
	rm -rf tmp.json