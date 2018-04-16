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

	Chart        string            `yaml:"chart"`
	ChartVersion string            `yaml:"chart-version"`
	ChartValues  map[string]string `yaml:"chart-values"`

	UserCredentials map[string]interface{} `yaml:"user-credentials"`
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
