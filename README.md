## Chainlink environment

### Local k8s cluster
Read [here](KUBERNETES.md) about how to spin up a local cluster

### Install
```shell
make install
```
If you have chaos-mesh installed in your cluster use, that will download CRDs from your current k8s ctx
```shell
make chaosmesh
```

### Running example environment
```shell
ENV_NAMESPACE="zclcdk-deployment" go run examples/chainlink/env.go
```
