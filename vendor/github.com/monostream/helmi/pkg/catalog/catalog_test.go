package catalog

import (
	"testing"
	"strconv"
)

const service string = "201cb950-e640-4453-9d91-4708ea0a1342"
const plan string = "7b16d6aa-260a-4b8d-b12c-464d2cedb9d0"

var c = Catalog{ }

func init() {
	c.Parse("../../catalog.yaml")
}

func red(msg string) (string){
	return "\033[31m" + msg + "\033[39m\n\n"
}

func Test_GetService(t *testing.T) {
	cs, _ := c.GetService(service)

	if cs.Name != "cassandra" {
		t.Error(red("service name is wrong"))
	}
}

func Test_GetServicePlan(t *testing.T) {
	csp, _ := c.GetServicePlan(service, plan)

	if csp.Name != "dev" {
		t.Error(red("service plan is wrong"))
	}
	if value, _ := strconv.Atoi(csp.ChartValues["replicaCount"]); value != 1 {
		t.Error(red("chart value in plan is wrong"))
	}
}