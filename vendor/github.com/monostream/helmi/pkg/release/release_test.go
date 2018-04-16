package release

import (
	"testing"
	"github.com/monostream/helmi/pkg/catalog"
	"github.com/monostream/helmi/pkg/helm"
	"github.com/monostream/helmi/pkg/kubectl"
)

var csp = catalog.CatalogPlan{
	Id:          "67890",
	Name:        "test_plan",
	Description: "plan_description",

	Chart: "plan_chart",
	ChartVersion: "1.2.3",
	ChartValues: map[string]string{
		"foo": "bar",
		"password": "{{ lookup('password', 'password') }}",
	},

	UserCredentials: map[string]interface{}{
		"key": "{{ lookup('value', 'foo') }}",
	},
}
var cs = catalog.CatalogService{
	Id:          "12345",
	Name:        "test_service",
	Description: "service_description",

	Chart: "service_chart",
	ChartVersion: "1.2.3",
	ChartValues: map[string]string{
		"foo": "bar",
		"password": "{{ lookup('password', 'password') }}",
	},

	UserCredentials: map[string]interface{}{
		"key": "{{ lookup('value', 'foo') }}",
		"hostname": "{{ lookup('cluster', 'address') }}",
		"port": "{{ lookup('cluster', 'port') }}",
	},

	Plans: []catalog.CatalogPlan{
		csp,
	},
}
var nodes = [] kubectl.Node {
	{
		Name: "test_node",

		Hostname:   "test_hostname",
		InternalIP: "1.1.1.1",
		ExternalIP: "2.2.2.2",
	},
}
var status = helm.Status{
	IsFailed:   false,
	IsDeployed: true,

	DesiredNodes: 1,
	AvailableNodes: 1,

	NodePorts: map[int]int{
		80: 30001,
	},
}

func red(msg string) (string){
	return "\033[31m" + msg + "\033[39m\n\n"
}

func Test_GetName(t *testing.T) {
	const input string = "this_is-a_test_name_which-is_pretty-long"
	const expected string = "helmithisisatestnam"

	name := getName(input)

	if len(name) != len(expected) {
		t.Error(red("length is wrong"))
	}
	if name != expected {
		t.Error(red("name is wrong"))
	}
}

func Test_GetChart(t *testing.T) {
	chart, _ := getChart(cs, catalog.CatalogPlan{})
	if chart != "service_chart" {
		t.Error(red("service chart not returned"))
	}
	chart, _ = getChart(cs, cs.Plans[0])
	if chart != "plan_chart" {
		t.Error(red("plan chart not returned"))
	}
	// no chart in plan
	csp.Chart = ""
	chart, _ = getChart(cs, csp)
	if chart != "service_chart" {
		t.Error(red("service chart for empty plan not returned"))
	}
}

func Test_GetChartValues(t *testing.T) {
	values := getChartValues(cs, csp)

	if values["foo"] != "bar" {
		t.Error(red("incorrect helm value returned"))
	}
	if len(values["password"]) != 32 {
		t.Error(red("incorrect helm value returned"))
	}
}

func Test_GetChartVersion(t *testing.T) {
	version, _ := getChartVersion(cs, csp)

	if version != "1.2.3" {
		t.Error(red("incorrect chart version returned"))
	}
}

func Test_GetUserCredentials(t *testing.T) {
	values := getUserCredentials(cs, csp, nodes, status, getChartValues(cs, catalog.CatalogPlan{}))

	if values["key"] != "bar" {
		t.Error(red("incorrect lookup value returned"))
	}
	if values["hostname"] != "2.2.2.2" {
		t.Error(red("incorrect hostname value returned"))
	}
	if values["port"] != "30001" {
		t.Error(red("incorrect port value returned"))
	}
}