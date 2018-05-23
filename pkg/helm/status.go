package helm

import (
	"bufio"
	"bytes"
	"strings"
	"strconv"
	"time"
	"os"
)

const (
	StatusFailed = "STATUS: FAILED"
	StatusDeployed = "STATUS: DEPLOYED"

	ResourcePrefix = "==> "

	NamespacePrefix = "NAMESPACE: "
	DeploymentTimePrefix = "LAST DEPLOYED: "

	DesiredLabel = "DESIRED"
	CurrentLabel = "CURRENT"
	AvailableLabel = "AVAILABLE"
	PortsLabel = "PORT(S)"
	TypeLabel = "TYPE"
	IngressHostsLabel = "HOSTS"
	IngressPortsLabel = "PORTS"
)

type Status struct {
	Name       string
	Namespace  string
	IsFailed   bool
	IsDeployed bool

	DesiredNodes int
	AvailableNodes int

	NodePorts map[int] int
	ClusterPorts map[int] int

	IngressHosts []string
	IngressPorts []int

	ServiceType string
}

func convertByteToStatus(release, namespace string, lastDeploymentTime time.Time, deployed bool, rawdata []byte) (Status, error) {

	status := Status{
		DesiredNodes: 0,
		AvailableNodes: 0,

		NodePorts:    map[int]int{},
		ClusterPorts: map[int]int{},
	}

	scanner := bufio.NewScanner(bytes.NewReader(rawdata))

	columnDesired := -1
	columnCurrent := -1
	columnAvailable := -1
	columnPort := -1
	columnType := -1
	columnIngressHosts := -1
	columnIngressPorts := -1

	var lastResource string

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			lastResource = ""

			columnDesired = -1
			columnCurrent = -1
			columnAvailable = -1
			columnPort = -1
			columnType = -1
			columnIngressHosts = -1
			columnIngressPorts = -1
		}

		if strings.HasPrefix(line, ResourcePrefix) {
			lastResource = strings.TrimPrefix(line, ResourcePrefix)
		}

		indexDesired := strings.Index(line, DesiredLabel)
		indexCurrent := strings.Index(line, CurrentLabel)
		indexAvailable := strings.Index(line, AvailableLabel)
		indexPort := strings.Index(line, PortsLabel)
		indexType := strings.Index(line, TypeLabel)
		indexIngressHosts := strings.Index(line, IngressHostsLabel)
		indexIngressPorts := strings.Index(line, IngressPortsLabel)

		if indexDesired >= 0 && indexCurrent >= 0 {
			columnDesired = indexDesired
			columnCurrent = indexCurrent

			if indexAvailable >= 0 {
				columnAvailable = indexAvailable
			}
		} else {
			if columnDesired >= 0 && columnCurrent >= 0 {
				nodesDesired := 0
				nodesAvailable := 0

				desired, desiredErr := strconv.Atoi(strings.Fields(line[columnDesired:])[0])
				current, currentErr := strconv.Atoi(strings.Fields(line[columnCurrent:])[0])

				if desiredErr == nil {
					nodesDesired = desired
				}

				if currentErr == nil {
					nodesAvailable = current
				}

				if columnAvailable >= 0 {
					available, availableErr := strconv.Atoi(strings.Fields(line[columnAvailable:])[0])

					if availableErr == nil {
						nodesAvailable = available
					}
				}

				status.DesiredNodes += nodesDesired
				status.AvailableNodes += nodesAvailable
			}
		}

		if indexIngressHosts >= 0 {
			columnIngressHosts = indexIngressHosts
		} else {
			if columnIngressHosts >= 0 {
				status.IngressHosts = strings.Split(strings.Fields(line[columnIngressHosts:])[0], ",")
			}
		}

		if indexIngressPorts >= 0 {
			columnIngressPorts = indexIngressPorts
		} else {
			if columnIngressPorts >= 0 {
				s := strings.Split(strings.Fields(line[columnIngressPorts:])[0], ",")
				for index, port := range s {
					ingressPort, ingressPortErr := strconv.Atoi(port)
					if ingressPortErr == nil {
						status.IngressPorts[index] = ingressPort
					}
				}

			}
		}

		if indexPort >= 0 && indexType >= 0 {
			columnPort = indexPort
			columnType = indexType
		} else {
			if columnPort >= 0 && columnType >= 0 {
				status.ServiceType = strings.Fields(line[columnType - 1:])[0]

				for _, portPair := range strings.Split(strings.Fields(line[columnPort:])[0], ",") {
					portFields := strings.FieldsFunc(portPair, func(c rune) bool {
						return c == ':' || c == '/'
					})

					if len(portFields) == 2 {
						clusterPort, clusterPortErr := strconv.Atoi(portFields[0])

						if clusterPortErr == nil {
							status.ClusterPorts[clusterPort] = clusterPort

						}
					}

					if len(portFields) == 3 {
						nodePort, nodePortErr := strconv.Atoi(portFields[1])
						clusterPort, clusterPortErr := strconv.Atoi(portFields[0])

						if nodePortErr == nil && clusterPortErr == nil {
							status.NodePorts[clusterPort] = nodePort
							status.ClusterPorts[clusterPort] = clusterPort
						}
					}
				}
			}
		}

		_ = lastResource
	}

	// timeout
	timeout, exists := os.LookupEnv("TIMEOUT")
	if !exists {
		timeout = "30m"
	}
	duration, _ := time.ParseDuration(timeout)
	if time.Now().After(lastDeploymentTime.Add(duration)) && status.AvailableNodes < status.DesiredNodes {
		status.IsFailed = true
	}

	status.Name = release
	status.Namespace = namespace
	status.IsDeployed = deployed

	return status, nil
}