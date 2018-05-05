package utils

import (
	"k8s.io/helm/pkg/strvals"
	"fmt"
	"gopkg.in/yaml.v2"
)

func ConvertInterfaceToByte(vals map[string]string) ([]byte, error){
	base := map[string]interface{}{}

	for name, value := range vals {
		if err := strvals.ParseInto(name + "=" + value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing data: %s", err)
		}
	}
	return yaml.Marshal(base)
}
