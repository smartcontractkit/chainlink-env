## Chainlink environment
Disclaimer: This software is in early Alpha stage, use at your own risk
### Local k8s cluster
Read [here](KUBERNETES.md) about how to spin up a local cluster

### Install from source
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
export CHAINLINK_ENV_USER="Satoshi" 
chainlink-env
```

# Develop
#### Pull chaosmesh CRD from your current k8s context (Optional)
```
make chaosmesh
```
#### Running standalone example environment
```shell
ENV_NAMESPACE="..." go run examples/chainlink/env.go
```

Add more CLI presets [here](./cmd/wizard/presets)

Add more programmatic examples [here](./examples/)