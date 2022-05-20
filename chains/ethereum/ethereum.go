package ethereum

import (
	"fmt"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	a "github.com/smartcontractkit/chainlink-env/alias"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
)

type Props struct {
	DevPeriod string `envconfig:"DEV_PERIOD"`
	GasPrice  string `envconfig:"GAS_PRICE"`
	GasTarget string `envconfig:"GAS_TARGET"`
}

func defaultDevChainProps() *Props {
	return &Props{
		DevPeriod: "1",
		GasPrice:  "10000000000",
		GasTarget: "80000000000",
	}
}

// chartData some shared labels/selectors and names that must match in resources
type chartData struct {
	Labels        *map[string]*string
	BaseName      string
	ConfigMapName string
	Props         *Props
}

func configMap(chart cdk8s.Chart, data chartData) {
	k8s.NewKubeConfigMap(chart, a.Str(data.ConfigMapName), &k8s.KubeConfigMapProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Str(data.ConfigMapName),
			Labels: &map[string]*string{
				"app": a.Str(data.ConfigMapName),
			},
		},
		Data: &map[string]*string{
			"password.txt": a.Str(""),
			"key1":         a.Str(`{"address":"f39fd6e51aad88f6f4ce6ab8827279cfffb92266","crypto":{"cipher":"aes-128-ctr","ciphertext":"c36afd6e60b82d6844530bd6ab44dbc3b85a53e826c3a7f6fc6a75ce38c1e4c6","cipherparams":{"iv":"f69d2bb8cd0cb6274535656553b61806"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"80d5f5e38ba175b6b89acfc8ea62a6f163970504af301292377ff7baafedab53"},"mac":"f2ecec2c4d05aacc10eba5235354c2fcc3776824f81ec6de98022f704efbf065"},"id":"e5c124e9-e280-4b10-a27b-d7f3e516b408","version":3}`),
			"key2":         a.Str(`{"address":"70997970c51812dc3a010c7d01b50e0d17dc79c8","crypto":{"cipher":"aes-128-ctr","ciphertext":"f8183fa00bc112645d3e23e29a233e214f7c708bf49d72750c08af88ad76c980","cipherparams":{"iv":"796d08e3e1f71bde89ed826abda96cda"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"03c864a22a1f7b06b1da12d8b93e024ac144f898285907c58b2abc135fc8a35c"},"mac":"5fe91b1a1821c0d9f85dfd582354ead9612e9a7e9adc38b06a2beff558c119ac"},"id":"d2cab765-5e30-42ae-bb91-f090d9574fae","version":3}`),
			"key3":         a.Str(`{"address":"3c44cdddb6a900fa2b585dd299e03d12fa4293bc","crypto":{"cipher":"aes-128-ctr","ciphertext":"2cd6ab87086c47f343f2c4d957eace7986f3b3c87fc35a2aafbefb57a06d9f1c","cipherparams":{"iv":"4e16b6cd580866c1aa642fb4d7312c9b"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"0cabde93877f6e9a59070f9992f7a01848618263124835c90d4d07a0041fc57c"},"mac":"94b7776ea95b0ecd8406c7755acf17b389b7ebe489a8942e32082dfdc1f04f57"},"id":"ade1484b-a3bb-426f-9223-a1f5e3bde2e8","version":3}`),
			"init.sh": a.Str(`#!/bin/bash
if [ ! -d /root/.ethereum/keystore ]; then
	echo "/root/.ethereum/keystore not found, running 'geth init'..."
	geth init /root/ethconfig/genesis.json
	echo "...done!"
fi

geth "$@"`),
			"genesis.json": a.Str(
				`{
      "config": {
        "chainId": 1337,
        "homesteadBlock": 0,
        "eip150Block": 0,
        "eip155Block": 0,
        "eip158Block": 0,
        "eip160Block": 0,
        "byzantiumBlock": 0,
        "constantinopleBlock": 0,
        "petersburgBlock": 0,
        "istanbulBlock": 0,
        "muirGlacierBlock": 0,
        "berlinBlock": 0,
        "londonBlock": 0
      },
      "nonce": "0x0000000000000042",
      "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "difficulty": "1",
      "coinbase": "0x3333333333333333333333333333333333333333",
      "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "extraData": "0x",
      "gasLimit": "8000000000",
      "alloc": {
        "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266": {
          "balance": "20000000000000000000000"
        },
        "0x70997970C51812dc3A010C7d01b50e0d17dc79C8": {
          "balance": "20000000000000000000000"
        },
        "0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC": {
          "balance": "20000000000000000000000"
        }
      }
    }`),
		},
	})
}

func service(chart cdk8s.Chart, data chartData) {
	k8s.NewKubeService(chart, a.Str(fmt.Sprintf("%s-service", data.BaseName)), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Str(data.BaseName),
		},
		Spec: &k8s.ServiceSpec{
			Ports: &[]*k8s.ServicePort{
				{
					Name:       a.Str("ws-rpc"),
					Port:       a.Num(8546),
					TargetPort: k8s.IntOrString_FromNumber(a.Num(8546)),
				},
				{
					Name:       a.Str("http-rpc"),
					Port:       a.Num(8544),
					TargetPort: k8s.IntOrString_FromNumber(a.Num(8544)),
				}},
			Selector: data.Labels,
		},
	})
}

func deployment(chart cdk8s.Chart, shared chartData) {
	k8s.NewKubeDeployment(
		chart,
		a.Str(fmt.Sprintf("%s-deployment", shared.BaseName)),
		&k8s.KubeDeploymentProps{
			Metadata: &k8s.ObjectMeta{
				Name: a.Str(shared.BaseName),
			},
			Spec: &k8s.DeploymentSpec{
				Selector: &k8s.LabelSelector{
					MatchLabels: shared.Labels,
				},
				Template: &k8s.PodTemplateSpec{
					Metadata: &k8s.ObjectMeta{
						Labels: shared.Labels,
					},
					Spec: &k8s.PodSpec{
						Volumes: &[]*k8s.Volume{
							{
								Name: a.Str(shared.ConfigMapName),
								ConfigMap: &k8s.ConfigMapVolumeSource{
									Name: a.Str(shared.ConfigMapName),
								},
							},
						},
						ServiceAccountName: a.Str("default"),
						Containers: &[]*k8s.Container{
							container(shared),
						},
					},
				},
			},
		})
}

func container(data chartData) *k8s.Container {
	props := defaultDevChainProps()
	config.MustEnvCodeOverrideStruct("", props, data.Props)
	return &k8s.Container{
		Name:            a.Str(data.BaseName),
		Image:           a.Str(fmt.Sprintf("%s:%s", "ethereum/client-go", "v1.10.17")),
		ImagePullPolicy: a.Str("Always"),
		Command: &[]*string{
			a.Str(`sh`),
			a.Str(`./root/init.sh`),
		},
		Args: &[]*string{
			a.Str("--dev"),
			a.Str("--password=/root/config/password.txt"),
			a.Str("--datadir=/root/.ethereum/devchain"),
			a.Str("--unlock=0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"),
			a.Str("--unlock=0x70997970C51812dc3A010C7d01b50e0d17dc79C8"),
			a.Str("--unlock=0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"),
			a.Str("--mine"),
			a.Str("--miner.etherbase=0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"),
			a.Str("--ipcdisable"),
			a.Str("--http"),
			a.Str("--http.addr=0.0.0.0"),
			a.Str("--http.port=8544"),
			a.Str("--http.vhosts=*"),
			a.Str("--ws"),
			a.Str("--ws.origins=*"),
			a.Str("--ws.addr=0.0.0.0"),
			a.Str("--ws.port=8546"),
			a.Str("--graphql"),
			a.Str("--graphql.corsdomain=*"),
			a.Str("--allow-insecure-unlock"),
			a.Str("--rpc.allow-unprotected-txs"),
			a.Str("--http.corsdomain=*"),
			a.Str("--vmdebug"),
			a.Str("--networkid=1337"),
			a.Str("--rpc.txfeecap=0"),

			a.Str(fmt.Sprintf("--dev.period=%s", props.DevPeriod)),
			a.Str(fmt.Sprintf("--miner.gasprice=%s", props.GasPrice)),
			a.Str(fmt.Sprintf("--miner.gastarget=%s", props.GasTarget)),
		},
		VolumeMounts: &[]*k8s.VolumeMount{
			{
				Name:      a.Str(data.ConfigMapName),
				MountPath: a.Str("/root/init.sh"),
				SubPath:   a.Str("init.sh"),
			},
			{
				Name:      a.Str(data.ConfigMapName),
				MountPath: a.Str("/root/config"),
			},
			{
				Name:      a.Str(data.ConfigMapName),
				MountPath: a.Str("/root/.ethereum/devchain/keystore/key1"),
				SubPath:   a.Str("key1"),
			},
			{
				Name:      a.Str(data.ConfigMapName),
				MountPath: a.Str("/root/.ethereum/devchain/keystore/key2"),
				SubPath:   a.Str("key2"),
			},
			{
				Name:      a.Str(data.ConfigMapName),
				MountPath: a.Str("/root/.ethereum/devchain/keystore/key3"),
				SubPath:   a.Str("key3"),
			},
		},
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Str("http-rpc"),
				ContainerPort: a.Num(8544),
			},
			{
				Name:          a.Str("ws-rpc"),
				ContainerPort: a.Num(8546),
			},
		},
		Resources: a.ContainerResources("200m", "528Mi", "200m", "528Mi"),
	}
}

func NewEthereum(chart cdk8s.Chart, props *Props) cdk8s.Chart {
	s := chartData{
		Labels: &map[string]*string{
			"app": a.Str("geth"),
		},
		ConfigMapName: "geth-cm",
		BaseName:      "geth",
		Props:         props,
	}
	service(chart, s)
	configMap(chart, s)
	deployment(chart, s)
	return chart
}
