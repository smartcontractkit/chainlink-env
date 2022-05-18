package ethereum

import (
	"fmt"
	"github.com/aws/constructs-go/constructs/v10"
	a "github.com/smartcontractkit/chainlink-env/alias"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
)

type Props struct {
	DevPeriod int
	GasPrice  int
	GasTarget int
}

func DefaultDevProps() *Props {
	return &Props{
		DevPeriod: 1,
		GasPrice:  10000000000,
		GasTarget: 80000000000,
	}
}

// SharedConstructVars some shared labels/selectors and names that must match in resources
type SharedConstructVars struct {
	Labels        *map[string]*string
	BaseName      string
	ConfigMapName string
	Props         *Props
}

func configMap(chart constructs.Construct, shared SharedConstructVars) {
	k8s.NewKubeConfigMap(chart, a.Jss(shared.ConfigMapName), &k8s.KubeConfigMapProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Jss(shared.ConfigMapName),
			Labels: &map[string]*string{
				"app": a.Jss(shared.ConfigMapName),
			},
		},
		Data: &map[string]*string{
			"password.txt": a.Jss(""),
			"key1":         a.Jss(`{"address":"f39fd6e51aad88f6f4ce6ab8827279cfffb92266","crypto":{"cipher":"aes-128-ctr","ciphertext":"c36afd6e60b82d6844530bd6ab44dbc3b85a53e826c3a7f6fc6a75ce38c1e4c6","cipherparams":{"iv":"f69d2bb8cd0cb6274535656553b61806"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"80d5f5e38ba175b6b89acfc8ea62a6f163970504af301292377ff7baafedab53"},"mac":"f2ecec2c4d05aacc10eba5235354c2fcc3776824f81ec6de98022f704efbf065"},"id":"e5c124e9-e280-4b10-a27b-d7f3e516b408","version":3}`),
			"key2":         a.Jss(`{"address":"70997970c51812dc3a010c7d01b50e0d17dc79c8","crypto":{"cipher":"aes-128-ctr","ciphertext":"f8183fa00bc112645d3e23e29a233e214f7c708bf49d72750c08af88ad76c980","cipherparams":{"iv":"796d08e3e1f71bde89ed826abda96cda"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"03c864a22a1f7b06b1da12d8b93e024ac144f898285907c58b2abc135fc8a35c"},"mac":"5fe91b1a1821c0d9f85dfd582354ead9612e9a7e9adc38b06a2beff558c119ac"},"id":"d2cab765-5e30-42ae-bb91-f090d9574fae","version":3}`),
			"key3":         a.Jss(`{"address":"3c44cdddb6a900fa2b585dd299e03d12fa4293bc","crypto":{"cipher":"aes-128-ctr","ciphertext":"2cd6ab87086c47f343f2c4d957eace7986f3b3c87fc35a2aafbefb57a06d9f1c","cipherparams":{"iv":"4e16b6cd580866c1aa642fb4d7312c9b"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"0cabde93877f6e9a59070f9992f7a01848618263124835c90d4d07a0041fc57c"},"mac":"94b7776ea95b0ecd8406c7755acf17b389b7ebe489a8942e32082dfdc1f04f57"},"id":"ade1484b-a3bb-426f-9223-a1f5e3bde2e8","version":3}`),
			"init.sh": a.Jss(`#!/bin/bash
if [ ! -d /root/.ethereum/keystore ]; then
	echo "/root/.ethereum/keystore not found, running 'geth init'..."
	geth init /root/ethconfig/genesis.json
	echo "...done!"
fi

geth "$@"`),
			"genesis.json": a.Jss(
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

func service(chart constructs.Construct, shared SharedConstructVars) {
	k8s.NewKubeService(chart, a.Jss(fmt.Sprintf("%s-service", shared.BaseName)), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name: a.Jss(shared.BaseName),
		},
		Spec: &k8s.ServiceSpec{
			Ports: &[]*k8s.ServicePort{
				{
					Name:       a.Jss("ws-rpc"),
					Port:       a.Jsn(8546),
					TargetPort: k8s.IntOrString_FromNumber(a.Jsn(8546)),
				},
				{
					Name:       a.Jss("http-rpc"),
					Port:       a.Jsn(8544),
					TargetPort: k8s.IntOrString_FromNumber(a.Jsn(8544)),
				}},
			Selector: shared.Labels,
		},
	})
}

func deployment(chart constructs.Construct, shared SharedConstructVars) {
	k8s.NewKubeDeployment(
		chart,
		a.Jss(fmt.Sprintf("%s-deployment", shared.BaseName)),
		&k8s.KubeDeploymentProps{
			Metadata: &k8s.ObjectMeta{
				Name: a.Jss(shared.BaseName),
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
								Name: a.Jss(shared.ConfigMapName),
								ConfigMap: &k8s.ConfigMapVolumeSource{
									Name: a.Jss(shared.ConfigMapName),
								},
							},
						},
						ServiceAccountName: a.Jss("default"),
						Containers: &[]*k8s.Container{
							container(shared),
						},
					},
				},
			},
		})
}

func container(shared SharedConstructVars) *k8s.Container {
	return &k8s.Container{
		Name:            a.Jss(shared.BaseName),
		Image:           a.Jss(fmt.Sprintf("%s:%s", "ethereum/client-go", "v1.10.17")),
		ImagePullPolicy: a.Jss("Always"),
		Command: &[]*string{
			a.Jss(`sh`),
			a.Jss(`./root/init.sh`),
		},
		Args: &[]*string{
			a.Jss("--dev"),
			a.Jss("--password=/root/config/password.txt"),
			a.Jss("--datadir=/root/.ethereum/devchain"),
			a.Jss("--unlock=0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"),
			a.Jss("--unlock=0x70997970C51812dc3A010C7d01b50e0d17dc79C8"),
			a.Jss("--unlock=0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC"),
			a.Jss("--mine"),
			a.Jss("--miner.etherbase=0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266"),
			a.Jss("--ipcdisable"),
			a.Jss("--http"),
			a.Jss("--http.addr=0.0.0.0"),
			a.Jss("--http.port=8544"),
			a.Jss("--http.vhosts=*"),
			a.Jss("--ws"),
			a.Jss("--ws.origins=*"),
			a.Jss("--ws.addr=0.0.0.0"),
			a.Jss("--ws.port=8546"),
			a.Jss("--graphql"),
			a.Jss("--graphql.corsdomain=*"),
			a.Jss("--allow-insecure-unlock"),
			a.Jss("--rpc.allow-unprotected-txs"),
			a.Jss("--http.corsdomain=*"),
			a.Jss("--vmdebug"),
			a.Jss("--networkid=1337"),
			a.Jss("--rpc.txfeecap=0"),

			a.Jss(fmt.Sprintf("--dev.period=%d", shared.Props.DevPeriod)),
			a.Jss(fmt.Sprintf("--miner.gasprice=%d", shared.Props.GasPrice)),
			a.Jss(fmt.Sprintf("--miner.gastarget=%d", shared.Props.GasTarget)),
		},
		VolumeMounts: &[]*k8s.VolumeMount{
			{
				Name:      a.Jss(shared.ConfigMapName),
				MountPath: a.Jss("/root/init.sh"),
				SubPath:   a.Jss("init.sh"),
			},
			{
				Name:      a.Jss(shared.ConfigMapName),
				MountPath: a.Jss("/root/config"),
			},
			{
				Name:      a.Jss(shared.ConfigMapName),
				MountPath: a.Jss("/root/.ethereum/devchain/keystore/key1"),
				SubPath:   a.Jss("key1"),
			},
			{
				Name:      a.Jss(shared.ConfigMapName),
				MountPath: a.Jss("/root/.ethereum/devchain/keystore/key2"),
				SubPath:   a.Jss("key2"),
			},
			{
				Name:      a.Jss(shared.ConfigMapName),
				MountPath: a.Jss("/root/.ethereum/devchain/keystore/key3"),
				SubPath:   a.Jss("key3"),
			},
		},
		Ports: &[]*k8s.ContainerPort{
			{
				Name:          a.Jss("http-rpc"),
				ContainerPort: a.Jsn(8899),
			},
			{
				Name:          a.Jss("ws-rpc"),
				ContainerPort: a.Jsn(8900),
			},
		},
		Env:       &[]*k8s.EnvVar{},
		Resources: a.ContainerResources("200m", "528Mi", "200m", "528Mi"),
	}
}

func NewEthereumChart(chart constructs.Construct, props *Props) constructs.Construct {
	s := SharedConstructVars{
		Labels: &map[string]*string{
			"app": a.Jss("geth"),
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
