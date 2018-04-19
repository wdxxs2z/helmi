package catalog

import (
	"log"
	"strings"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

type Catalog struct {
	Services []CatalogService `yaml:"services"`
}

type CatalogService struct {
	Id          string `yaml:"_id"`
	Name        string `yaml:"_name"`
	Description string `yaml:"description"`

	Tags        	 []string		`yaml:"tags"`
	Requires    	 []string		`yaml:"requires"`
	Bindable    	 bool			`yaml:"bindable"`
	Metadata    	 map[string]string	`yaml:"metadata"`
	DashboardClient  map[string]string	`yaml:"dashboard_client"`
	PlanUpdateable   bool			`yaml:"plan_updateable"`

	Chart        string            `yaml:"chart"`
	ChartVersion string            `yaml:"chart-version"`
	ChartValues  map[string]string `yaml:"chart-values"`

	UserCredentials map[string]interface{} `yaml:"user-credentials"`

	Plans []CatalogPlan `yaml:"plans"`
}

type CatalogPlan struct {
	Id          string `yaml:"_id"`
	Name        string `yaml:"_name"`
	Description string `yaml:"description"`

	Free        bool 			`yaml:"free"`
	Bindable    bool			`yaml:"bindable"`
	Metadata    PlanMetadata		`yaml:"metadata"`
	Schemas     SchemasObject		`yaml:"schemas"`

	Chart        string            `yaml:"chart"`
	ChartVersion string            `yaml:"chart-version"`
	ChartValues  map[string]string `yaml:"chart-values"`

	UserCredentials map[string]interface{} `yaml:"user-credentials"`
}

type PlanMetadata struct {
	Costs    []Cost		`yaml:"costs"`
	Bullets  []string	`yaml:"bullets"`
}

type SchemasObject struct {
	ServiceInstance	 ServiceInstanceSchemaObject		`yaml:"service_instance"`
	ServiceBinding	 ServiceBindingSchemaObject		`yaml:"service_binding"`
}

type ServiceInstanceSchemaObject struct {
	Create	CreateUpdateSchemaObject		`yaml:"create"`
	Update  CreateUpdateSchemaObject		`yaml:"update"`
}

type ServiceBindingSchemaObject struct {
	Create	CreateUpdateSchemaObject		`yaml:"create"`
}

type CreateUpdateSchemaObject struct {
	Parameters map[string]interface{}		`yaml:"parameters"`
}


type Cost struct {
	Amount    map[string]string	`yaml:"amount"`
	Unit      string		`yaml:"unit"`
}

func (c *Catalog) Parse(path string) {
	input, err := ioutil.ReadFile(path)

	if err != nil {
		log.Printf("Catalog.Read: #%v ", err)
	}

	// insert fake root to allow parsing
	data := "services:\n" + string(input)
	input = []byte(data)

	err = yaml.Unmarshal(input, c)

	if err != nil {
		log.Fatalf("Catalog.Unmarshal: %v", err)
	}
}

func (c *Catalog) GetService(service string) (CatalogService, error) {
	for _, s := range c.Services {
		if strings.EqualFold(s.Id, service) {
			return s, nil
		}
	}

	return *new(CatalogService), nil
}

func (c *Catalog) GetServicePlan(service string, plan string) (CatalogPlan, error) {
	for _, s := range c.Services {
		if strings.EqualFold(s.Id, service) {
			for _, p := range s.Plans {
				if strings.EqualFold(p.Id, plan) {
					return p, nil
				}
			}
		}
	}

	return *new(CatalogPlan), nil
}
