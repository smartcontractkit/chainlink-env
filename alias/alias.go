package alias

import (
	jsii "github.com/aws/jsii-runtime-go"
	"github.com/fatih/structs"
	"github.com/smartcontractkit/chainlink-env/config"
	"github.com/smartcontractkit/chainlink-env/imports/k8s"
	"reflect"
)

func Str(value string) *string {
	return jsii.String(value)
}

func Num(value float64) *float64 {
	return jsii.Number(value)
}

// MustChartEnvVarsFromStruct parses typed configs into manifest env vars
func MustChartEnvVarsFromStruct(prefix string, defaults interface{}, s interface{}) *[]*k8s.EnvVar {
	config.MustEnvCodeOverrideStruct(prefix, defaults, s)
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
		Name:  Str(k),
		Value: Str(v),
	}
}

// ContainerResources container resource requirements
func ContainerResources(reqCPU, reqMEM, limCPU, limMEM string) *k8s.ResourceRequirements {
	return &k8s.ResourceRequirements{
		Requests: &map[string]k8s.Quantity{
			"cpu":    k8s.Quantity_FromString(Str(reqCPU)),
			"memory": k8s.Quantity_FromString(Str(reqMEM)),
		},
		Limits: &map[string]k8s.Quantity{
			"cpu":    k8s.Quantity_FromString(Str(limCPU)),
			"memory": k8s.Quantity_FromString(Str(limMEM)),
		},
	}
}
