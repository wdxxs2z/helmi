package kubectl

import (
	"bytes"
	"strings"
	"os/exec"
	"encoding/json"
	"github.com/jmoiron/jsonq"
	"errors"
)

type Node struct {
	Name string

	Hostname   string
	InternalIP string
	ExternalIP string
}

func GetNodes() ([] Node, error) {
	cmd := exec.Command("kubectl", "get", "nodes", "--output", "json")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, errors.New(string(output[:]))
	}

	data := map[string]interface{}{}

	decoder := json.NewDecoder(bytes.NewReader(output))
	decoder.Decode(&data)

	dataQuery := jsonq.NewQuery(data)

	var nodes [] Node

	items, err := dataQuery.ArrayOfObjects("items")

	if err != nil {
		return nil, err
	}

	for _, item := range items {
		itemQuery := jsonq.NewQuery(item)

		node := Node{}

		nodeId, err := itemQuery.String("spec", "externalID")

		if err != nil {
			return nil, err
		}

		node.Name = nodeId

		nodeAddresses, err := itemQuery.ArrayOfObjects("status", "addresses")

		if err != nil {
			return nil, err
		}

		for _, nodeAddress := range nodeAddresses {
			nodeAddressQuery := jsonq.NewQuery(nodeAddress)

			addressType, err := nodeAddressQuery.String("type")

			if err != nil {
				return nil, err
			}

			addressValue, err := nodeAddressQuery.String("address")

			if err != nil {
				return nil, err
			}

			if strings.EqualFold(addressType, "Hostname") {
				node.Hostname = addressValue
			}

			if strings.EqualFold(addressType, "InternalIP") {
				node.InternalIP = addressValue
			}

			if strings.EqualFold(addressType, "ExternalIP") {
				node.ExternalIP = addressValue
			}
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
