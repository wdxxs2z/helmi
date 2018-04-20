package catalog

import (
	"log"
	"strings"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
)

type Catalog struct {
	Services []CatalogService `yaml:"services"`
}

type CatalogService struct {
	Id          		string 			`yaml:"_id"`
	Name        		string 			`yaml:"_name"`
	Description 		string 			`yaml:"description"`

	Tags        	 	[]string		`yaml:"tags"`
	Requires    	 	[]string		`yaml:"requires"`
	Bindable    	 	bool			`yaml:"bindable"`
	Metadata    	 	map[string]string	`yaml:"metadata"`
	DashboardClient  	map[string]string	`yaml:"dashboard_client"`
	PlanUpdateable   	bool			`yaml:"plan_updateable"`

	Chart        		string            	`yaml:"chart"`
	ChartVersion 		string            	`yaml:"chart-version"`
	ChartValues  		map[string]string 	`yaml:"chart-values"`

	UserCredentials 	map[string]interface{} `yaml:"user-credentials"`

	Plans 			[]CatalogPlan 		`yaml:"plans"`
}

type CatalogPlan struct {
	Id          		string 			`yaml:"_id"`
	Name        		string 			`yaml:"_name"`
	Description 		string 			`yaml:"description"`

	Free        		bool 			`yaml:"free"`
	Bindable    		bool			`yaml:"bindable"`
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
	Amount    		map[string]interface{}	`yaml:"amount"`
	Unit      		string			`yaml:"unit"`
}

func (c Catalog) Validate() error {
	for _, service := range c.Services {
		if err := service.Validate(); err != nil {
			return fmt.Errorf("Validating Services configuration: %s", err)
		}
	}

	return nil
}

func (s CatalogService) Validate() error {
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

func (sp CatalogPlan) Validate() error {
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
