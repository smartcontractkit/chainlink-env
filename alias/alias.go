package alias

import (
	jsii "github.com/aws/jsii-runtime-go"
	"github.com/fatih/structs"
	"github.com/imdario/mergo"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
	"reflect"
)

func Jss(value string) *string {
	return jsii.String(value)
}

func Jsn(value float64) *float64 {
	return jsii.Number(value)
}

func MustOverrideStruct(prefix string, s interface{}) {
	if err := envconfig.Process(prefix, s); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func MustEnvVarsFromEnvconfigPrefix(prefix string, defaults interface{}, s interface{}) *[]*k8s.EnvVar {
	if err := mergo.Merge(defaults, s, mergo.WithOverride); err != nil {
		log.Fatal().Err(err).Send()
	}
	MustOverrideStruct(prefix, defaults)
	var e []*k8s.EnvVar
	ma := structs.Map(defaults)
	for k, v := range ma {
		field, _ := reflect.TypeOf(defaults).Elem().FieldByName(k)
		tag := field.Tag.Get("envconfig")
		e = append(e, EnvVarStr(tag, v.(string)))
	}
	return &e
}

// EnvVarStr quick shortcut for string/string key/value var
func EnvVarStr(k, v string) *k8s.EnvVar {
	return &k8s.EnvVar{
		Name:  Jss(k),
		Value: Jss(v),
	}
}

// ContainerResources container resource requirements
func ContainerResources(reqCPU, reqMEM, limCPU, limMEM string) *k8s.ResourceRequirements {
	return &k8s.ResourceRequirements{
		Requests: &map[string]k8s.Quantity{
			"cpu":    k8s.Quantity_FromString(Jss(reqCPU)),
			"memory": k8s.Quantity_FromString(Jss(reqMEM)),
		},
		Limits: &map[string]k8s.Quantity{
			"cpu":    k8s.Quantity_FromString(Jss(limCPU)),
			"memory": k8s.Quantity_FromString(Jss(limMEM)),
		},
	}
}
