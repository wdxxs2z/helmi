package catalog

import (
	"log"
	"strings"
	"strconv"
	"fmt"
	"path/filepath"
	"os"
	"io/ioutil"
	"bytes"
	"text/template"

	"gopkg.in/yaml.v2"
	"github.com/Masterminds/sprig"
	"github.com/satori/go.uuid"

	helmi "github.com/wdxxs2z/helmi/pkg/helm"
	"github.com/wdxxs2z/helmi/pkg/kubectl"
)

type Catalog struct {
	Services map[string]Service
}

type Service struct {
	Id          		string 			`yaml:"_id"`
	Name        		string 			`yaml:"_name"`
	Description 		string 			`yaml:"description"`

	Tags        	 	[]string		`yaml:"tags"`
	Requires    	 	[]string		`yaml:"requires"`
	Bindable    	 	bool			`yaml:"bindable"`
	Metadata    	 	ServiceMetadata		`yaml:"metadata"`
	DashboardClient  	map[string]string	`yaml:"dashboard_client"`
	PlanUpdateable   	bool			`yaml:"plan_updateable"`

	Chart        		string            	`yaml:"chart"`
	ChartVersion 		string            	`yaml:"chart-version"`
	ChartOffline            string                  `yaml:"chart-offline"`

	InternalDiscoveryName   string                  `yaml:"internel-discovery-name"`

	Plans 			[]Plan 			`yaml:"plans"`

	valuesTemplate      	*template.Template
	credentialsTemplate 	*template.Template
}

type ServiceMetadata struct {
	DisplayName         string 	`yaml:"displayName"`
	ImageUrl            string 	`yaml:"imageUrl"`
	LongDescription     string 	`yaml:"longDescription"`
	ProviderDisplayName string 	`yaml:"providerDisplayName"`
	DocumentationUrl    string 	`yaml:"documentationUrl"`
	SupportUrl          string 	`yaml:"supportUrl"`
}

type Plan struct {
	Id          		string 			`yaml:"_id"`
	Name        		string 			`yaml:"_name"`
	Description 		string 			`yaml:"description"`

	Free        		*bool 			`yaml:"free"`
	Bindable    		*bool			`yaml:"bindable"`
	Metadata    		PlanMetadata		`yaml:"metadata"`

	Chart        		string            	`yaml:"chart"`
	ChartVersion 		string            	`yaml:"chart-version"`
	ChartValues  		map[string]string 	`yaml:"chart-values"`

	UserCredentials 	map[string]interface{} `yaml:"user-credentials"`
}

type PlanMetadata struct {
	Costs    		[]Cost			`yaml:"costs"`
	Bullets  		[]string		`yaml:"bullets"`
}


type Cost struct {
	Amount    		map[string]float64	`yaml:"amount"`
	Unit      		string			`yaml:"unit"`
}

func (c Catalog) Validate() error {
	for _, s := range c.Services {
		if err := s.Validate(); err != nil {
			return fmt.Errorf("Validating Services configuration: %s", err)
		}
	}

	return nil
}

func (s Service) Validate() error {
	if s.Id == "" {
		return fmt.Errorf("Must provide a non-empty Id (%+v)", s)
	}

	if s.Name == "" {
		return fmt.Errorf("Must provide a non-empty Name (%+v)", s)
	}

	if s.Description == "" {
		return fmt.Errorf("Must provide a non-empty Description (%+v)", s)
	}

	for _, servicePlan := range s.Plans {
		if err := servicePlan.Validate(); err != nil {
			return fmt.Errorf("Validating Plans configuration: %s", err)
		}
	}

	return nil
}

func (sp Plan) Validate() error {
	if sp.Id == "" {
		return fmt.Errorf("Must provide a non-empty ID (%+v)", sp)
	}

	if sp.Name == "" {
		return fmt.Errorf("Must provide a non-empty Name (%+v)", sp)
	}

	if sp.Description == "" {
		return fmt.Errorf("Must provide a non-empty Description (%+v)", sp)
	}

	return nil
}

func (c *Catalog) Parse(rawData []byte) {

	err := yaml.Unmarshal(rawData, c)

	if err != nil {
		log.Fatalf("Catalog.Unmarshal: %v", err)
	}
}

func ParseDir(dir string) (Catalog, error) {
	catalog := Catalog{
		Services: make(map[string]Service),
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Unable to read catalog file: %q: %s", path, err)
			return nil
		}
		ext := filepath.Ext(path)
		if info.IsDir() || ( ext != ".yaml" && ext != ".yml" ) {
			return nil
		}
		input, ioErr := ioutil.ReadFile(path)
		if ioErr != nil {
			return ioErr
		}
		return catalog.parseServiceDefinition(input, path)
	})

	if err != nil {
		return catalog, err
	}

	if len(catalog.Services) == 0 {
		err = fmt.Errorf("no services found in catalog directory: %s", dir)
		return catalog, err
	}

	return catalog, nil
}

func (c *Catalog) parseServiceDefinition(input []byte, file string) error {
	documents := bytes.Split(input, []byte("\n---"))
	if n := len(documents); n != 3 {
		return fmt.Errorf("service file %s: must contain 3 yaml document parts, found %d", file, n)
	}
	var s struct{ Service }
	err := yaml.Unmarshal(documents[0], &s)
	if err != nil {
		return fmt.Errorf("failed to parse service definition: %s: %s", file, err)
	}
	fMap := templateFuncMap()
	valuesTemplate, valuesErr := template.New("values").Funcs(fMap).Parse(string(documents[1]))
	if valuesErr != nil {
		return fmt.Errorf("failed to parse values template: %s: %s", file, valuesErr)
	}
	credentialsTemplate, credentialsErr := template.New("credentials").Funcs(fMap).Parse(string(documents[2]))
	if credentialsErr != nil {
		return fmt.Errorf("failed to parse credentials template: %s: %s", file, credentialsErr)
	}

	s.valuesTemplate = valuesTemplate
	s.credentialsTemplate = credentialsTemplate

	c.Services[s.Id] = s.Service
	return nil
}

func templateFuncMap() template.FuncMap {
	f := sprig.TxtFuncMap()

	randomUuid := func() string {
		s := uuid.NewV4().String()
		s = strings.Replace(s, "-", "", -1)
		return s
	}

	f["generateUsername"] = randomUuid
	f["generatePassword"] = randomUuid

	return f
}

func (c *Catalog) GetService(service string) (Service, error) {
	for _, s := range c.Services {
		if strings.EqualFold(s.Id, service) {
			return s, nil
		}
	}

	return *new(Service), nil
}

func (c *Catalog) GetServicePlan(service string, plan string) (Plan, error) {
	for _, s := range c.Services {
		if strings.EqualFold(s.Id, service) {
			for _, p := range s.Plans {
				if strings.EqualFold(p.Id, plan) {
					return p, nil
				}
			}
		}
	}

	return *new(Plan), nil
}

type chartValueVars struct {
	*Service
	*Plan
}

func (s *Service) ChartValues(p *Plan) (map[string]string, error) {
	b := new(bytes.Buffer)
	data := chartValueVars{s, p}
	err := s.valuesTemplate.Execute(b, data)
	if err != nil {
		return nil, err
	}

	var v struct {
		ChartValues map[string]string `yaml:"chart-values"`
	}


	err = yaml.Unmarshal(b.Bytes(), &v)
	if err != nil {
		return nil, err
	}

	if v.ChartValues == nil {
		v.ChartValues = make(map[string]string)
	}

	for key, value := range p.ChartValues {
		v.ChartValues[key] = value
	}

	return v.ChartValues, nil
}

type credentialVars struct {
	Service *Service
	Plan    *Plan
	Values  valueVars
	Release releaseVars
	Cluster clusterVars
}

type valueVars map[interface{}]interface{}

type releaseVars struct {
	Name      string
	Namespace string
}

type clusterVars struct {
	Address    	string
	HaAddress       string
	Hostname   	string
	IngressAddress  string
	IngressPort     string
	helmStatus 	helmi.Status
}

func ingressPort(helmStatus helmi.Status) string {
	if len(helmStatus.Ingresses) > 0 {
		return strconv.Itoa(helmStatus.Ingresses[0].IngressPort)
	} else {
		return ""
	}
}

func ingressAddress(helmStatus helmi.Status) string {
	if len(helmStatus.Ingresses) > 0 {
		return helmStatus.Ingresses[0].IngressHosts[0]
	} else {
		return ""
	}
}

func (c clusterVars) Port(port ...int) string {

	service := c.helmStatus.Services[0]

	if service.ServiceType == "LoadBalancer" {
		for clusterPort, nodePort := range service.ClusterPorts {
			if len(port) == 0 || port[0] == clusterPort {
				return strconv.Itoa(nodePort)
			}
		}
	}

	for clusterPort, nodePort := range service.NodePorts {
		if len(port) == 0 || port[0] == clusterPort {
			return strconv.Itoa(nodePort)
		}
	}

	for clusterPort, nodePort := range service.ClusterPorts {
		if len(port) == 0 || port[0] == clusterPort {
			return strconv.Itoa(nodePort)
		}
	}

	if len(port) > 0 {
		return strconv.Itoa(port[0])
	}

	return ""
}

func extractHAAddress(kubernetesNodes []kubectl.Node, helmStatus helmi.Status) string {
	addresses := []string{}
	for _, service := range helmStatus.Services {
		if service.ServiceType == "ClusterIP" {
			if clusterDns, ok := os.LookupEnv("CLUSTER_DNS"); ok {
				var port string
				for _, nodePort := range service.ClusterPorts{
					port = strconv.Itoa(nodePort)
				}
				address := fmt.Sprintf("%s.%s.%s:%s", service.Name, helmStatus.Namespace, clusterDns, port)
				addresses = append(addresses, address)
			} else {
				return ""
			}
		} else if service.ServiceType == "NodePort" {
			var port string
			for _, nodePort := range service.NodePorts {
				port = strconv.Itoa(nodePort)
			}
			var ip string
			for _, node := range kubernetesNodes {
				if len(node.ExternalIP) > 0 {
					ip = node.ExternalIP
				}
			}
			for _, node := range kubernetesNodes {
				if len(node.InternalIP) > 0 {
					ip = node.InternalIP
				}
			}
			address := fmt.Sprintf("%s:%s", ip, port)
			addresses = append(addresses, address)
		} else if service.ServiceType == "LoadBalancer" {
			var port string
			for _, nodePort := range service.ClusterPorts{
				port = strconv.Itoa(nodePort)
			}
			address := fmt.Sprintf("%s:%s", service.ExternalIP, port)
			addresses = append(addresses, address)
		}
	}
	return strings.Join(addresses, ",")
}

func extractAddress(kubernetesNodes []kubectl.Node, helmStatus helmi.Status, discoveryName string) string {
	var masterService helmi.StatusService
	for _,service := range helmStatus.Services {
		if strings.Contains(service.Name, discoveryName){
			masterService = service
		}
	}

	// return dns name if set as environment variable
	if value, ok := os.LookupEnv("DOMAIN"); ok {
		return value
	}

	if masterService.ServiceType == "ClusterIP" {
		if value, ok := os.LookupEnv("CLUSTER_DNS"); ok {
			return fmt.Sprintf("%s-%s.%s.%s", helmStatus.Name, discoveryName, helmStatus.Namespace, value)
		}

	} else if masterService.ServiceType == "NodePort" {
		for _, node := range kubernetesNodes {
			if len(node.ExternalIP) > 0 {
				return node.ExternalIP
			}
		}
		for _, node := range kubernetesNodes {
			if len(node.InternalIP) > 0 {
				return node.InternalIP
			}
		}
	} else if masterService.ServiceType == "LoadBalancer" {
		return masterService.ExternalIP
	} else if masterService.ServiceType == "ExternalName" {
		//TODO
		return ""
	}

	return ""
}

func extractHostname(kubernetesNodes []kubectl.Node) string {
	for _, node := range kubernetesNodes {
		if len(node.Hostname) > 0 {
			return node.Hostname
		}
	}

	return ""
}

func (s *Service) UserCredentials(plan *Plan, kubernetesNodes []kubectl.Node, helmStatus helmi.Status, values map[interface{}]interface{}) (map[string]interface{}, error) {

	env := credentialVars{
		Service: s,
		Plan:    plan,
		Values:  values,
		Release: releaseVars{
			Name:      helmStatus.Name,
			Namespace: helmStatus.Namespace,
		},
		Cluster: clusterVars{
			Address:    extractAddress(kubernetesNodes, helmStatus, s.InternalDiscoveryName),
			HaAddress:  extractHAAddress(kubernetesNodes, helmStatus),
			Hostname:   extractHostname(kubernetesNodes),
			IngressAddress: ingressAddress(helmStatus),
			IngressPort: ingressPort(helmStatus),
			helmStatus: helmStatus,
		},
	}

	b := new(bytes.Buffer)
	err := s.credentialsTemplate.Execute(b, env)
	if err != nil {
		return nil, err
	}

	var v struct {
		UserCredentials map[string]interface{} `yaml:"user-credentials"`
	}

	err = yaml.Unmarshal(b.Bytes(), &v)

	if err != nil {
		return nil, err
	}

	return v.UserCredentials, nil
}