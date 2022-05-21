## Chainlink environment

### Local k8s cluster
Read [here](KUBERNETES.md) about how to spin up a local cluster

### Install
Set up deps
```shell
make install_deps
```
Install CLI wizard
```
make install
```

### Use
Start up a wizard to create/connect to environments and more
```
chainlink-env
```

# Develop
#### Pull chaosmesh CRD from your current k8s context (Optional)
```
make chaosmesh
```
#### Running standalone example environment
```shell
ENV_NAMESPACE="zclcdk-deployment" go run examples/chainlink/env.go
```
