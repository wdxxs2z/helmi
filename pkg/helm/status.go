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
	ResourcePrefix = "==> "

	DesiredLabel = "DESIRED"
	CurrentLabel = "CURRENT"
	AvailableLabel = "AVAILABLE"
	PortsLabel = "PORT(S)"
	TypeLabel = "TYPE"
	ExternalIPLabel = "EXTERNAL-IP"
	IngressHostsLabel = "HOSTS"
	IngressPortsLabel = "PORTS"
	NameLabel = "NAME"
	ClusterIPLabel = "CLUSTER-IP"
	AgeLabel = "AGE"
	AddressLable = "ADDRESS"


)

type Status struct {
	Name       string
	Namespace  string
	IsFailed   bool
	IsDeployed bool

	DesiredNodes int
	AvailableNodes int

	Services []StatusService
	Ingresses []StatusIngress
}

type StatusService struct {
	Name       	string
	ServiceType 	string
	NodePorts 	map[int] int
	ClusterPorts 	map[int] int
	ExternalIP 	string
}

type StatusIngress struct {
	Name       	string
	IngressHosts 	[]string
	IngressPort 	int
}

func convertByteToStatus(release, namespace string, lastDeploymentTime time.Time, deployed bool, rawdata []byte) (Status, error) {

	status := Status{
		DesiredNodes: 0,
		AvailableNodes: 0,
	}

	scanner := bufio.NewScanner(bytes.NewReader(rawdata))

	columnDesired := -1
	columnCurrent := -1
	columnAvailable := -1
	columnPort := -1
	columnType := -1
	columnIngressHosts := -1
	columnIngressPorts := -1
	columnExternalIP := -1
	columnName := -1
	columnClusterIP := -1
	columnAge := -1
	columnAddress := -1

	var lastResource string

	//init helm status
	var statusServices = []StatusService{}
	var statusIngresses = []StatusIngress{}

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
			columnExternalIP = -1
			columnName = -1
			columnClusterIP = -1
			columnAge = -1
			columnAddress = -1
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
		indexExternalIP := strings.Index(line, ExternalIPLabel)
		indexName := strings.Index(line, NameLabel)
		indexClusterIP := strings.Index(line, ClusterIPLabel)
		indexAge := strings.Index(line, AgeLabel)
		indexAddress := strings.Index(line, AddressLable)


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

		if indexIngressHosts >= 0 && indexIngressPorts >= 0 && indexName >= 0 && indexAddress >= 0 && indexAge >= 0 {
			columnIngressHosts = indexIngressHosts
			columnIngressPorts = indexIngressPorts
			columnName = indexName
			columnAddress = indexAddress
			columnAge = indexAge
		} else {
			if columnIngressHosts >= 0  && columnIngressPorts >= 0 && columnName >= 0  && columnAddress >= 0 && columnAge >= 0 {

				ingressesHosts := strings.Fields(line[columnIngressHosts : columnAddress])
				ingressesPorts := strings.Fields(line[columnIngressPorts : columnAge])
				ingressesName := strings.Fields(line[columnName : columnIngressHosts])

				for i:= 0; i < len(ingressesName); i++ {
					var port int
					hosts := strings.Split(ingressesHosts[i], ",")
					ingressPort , portErr := strconv.Atoi(strings.Split(ingressesPorts[i], ",")[0])
					if portErr == nil {
						port = ingressPort
					}
					statusIngress := StatusIngress{
						Name:        	ingressesName[i],
						IngressHosts:	hosts,
						IngressPort:    port,
					}
					statusIngresses = append(statusIngresses, statusIngress)
				}
				status.Ingresses = statusIngresses
			}
		}

		if indexPort >= 0 && indexType >= 0 && indexExternalIP >= 0 && indexName >= 0 && indexClusterIP >= 0 && indexAge >= 0 {
			columnPort = indexPort
			columnType = indexType
			columnExternalIP = indexExternalIP
			columnName = indexName
			columnClusterIP = indexClusterIP
			columnAge = indexAge
		} else {
			if columnPort >= 0 && columnType >= 0 && columnExternalIP >= 0 && columnName >= 0 && columnClusterIP >= 0 && columnAge >= 0{

				servicesName := strings.Fields(line[columnName : columnType])
				servicesType := strings.Fields(line[columnType  : columnClusterIP])
				servicesExternalIP := strings.Fields(line[columnExternalIP : columnPort])
				servicesPort := strings.Fields(line[columnPort : columnAge])


				for i := 0; i < len(servicesName); i++ {
					clusterPorts := make(map[int]int)
					nodePorts := make(map[int]int)

					for _, portPair := range strings.Split(servicesPort[i], ",") {
						portFields := strings.FieldsFunc(portPair, func(c rune) bool {
							return c == ':' || c == '/'
						})
						if len(portFields) == 2 {
							clusterPort, clusterPortErr := strconv.Atoi(portFields[0])

							if clusterPortErr == nil {
								clusterPorts[clusterPort] = clusterPort
							}
						}
						if len(portFields) == 3 {
							nodePort, nodePortErr := strconv.Atoi(portFields[1])
							clusterPort, clusterPortErr := strconv.Atoi(portFields[0])

							if nodePortErr == nil && clusterPortErr == nil {
								nodePorts[clusterPort] = nodePort
								clusterPorts[clusterPort] = clusterPort
							}
						}
					}
					statusService := StatusService{
						Name:		servicesName[i],
						ServiceType:    servicesType[i],
						ExternalIP:     servicesExternalIP[i],
						ClusterPorts:   clusterPorts,
						NodePorts:      nodePorts,

					}
					statusServices = append(statusServices, statusService)
				}
				status.Services = statusServices
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