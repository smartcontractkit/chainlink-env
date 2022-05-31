## Chainlink environment
Disclaimer: This software is in early Alpha stage, use at your own risk
### Local k8s cluster
Read [here](KUBERNETES.md) about how to spin up a local cluster

### Install
Easiest way is to use a release [binaries](https://github.com/smartcontractkit/chainlink-env/releases)

#### From source
Set up deps, you need to have [yarn](https://classic.yarnpkg.com/lang/en/docs/install/#mac-stable)
```shell
make install_deps
```
Install CLI wizard
```
make install
```

### Usage
Run 
```
export CHAINLINK_IMAGE="public.ecr.aws/chainlink/chainlink"
export CHAINLINK_TAG="1.4.0-root"
export CHAINLINK_ENV_USER="Satoshi"
chainlink-env
```

# Develop
#### Running standalone example environment
```shell
go run examples/simple/env.go
```
If you have another env of that type, you can connect by overriding environment name
```
ENV_NAMESPACE="..."  go run examples/chainlink/enc.go
```

Add more CLI presets [here](./cmd/wizard/presets)

Add more programmatic examples [here](./examples/)

If you have [chaosmesh]() installed in your cluster you can pull and generated CRD in go like that
```
make chaosmesh
```