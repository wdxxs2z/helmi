package release

import (
	"errors"
	"regexp"
	"strings"
	"strconv"
	"github.com/satori/go.uuid"
	helmi "github.com/wdxxs2z/helmi/pkg/helm"
	"github.com/wdxxs2z/helmi/pkg/kubectl"
	"github.com/wdxxs2z/helmi/pkg/catalog"
	helmicons "github.com/wdxxs2z/helmi/pkg/constants"
	"code.cloudfoundry.org/lager"
	"os"
	"reflect"
	"fmt"
)

type Status struct {
	IsFailed    bool
	IsDeployed  bool
	IsAvailable bool
}

func Install(catalog *catalog.Catalog,
	serviceId string,
	planId string,
	id string,
	acceptsIncomplete bool,
	parameters map[string]string,
	context map[string]string,
	client *helmi.Client,
	logger lager.Logger) error {

	logger.Debug("release-install",lager.Data{
		helmicons.InstanceIDLogKey: id,
	})

	name := getName(id)

	service, _ := catalog.GetService(serviceId)
	plan, _ := catalog.GetServicePlan(serviceId, planId)

	chart, chartErr := getChart(service, plan)
	chartVersion, chartVersionErr := getChartVersion(service, plan)
	chartValues := getChartValues(service, plan, parameters)
	chartNamespace := getChartNamespace(context, parameters)

	if chartErr != nil {
		logger.Error("failed-install-release", chartErr, lager.Data{
			"id": id,
			"name": name,
			"service-id": serviceId,
			"plan-id": planId,
		})
		return chartErr
	}

	if chartVersionErr != nil {
		chartVersion = ""
	}

	_, err := client.InstallRelease(name, chart, chartVersion, chartValues, chartNamespace, acceptsIncomplete)

	if err != nil {
		logger.Error("failed-install-release", err, lager.Data{
			"id": id,
			"name": name,
			"chart": chart,
			"chart-version": chartVersion,
			"service-id": serviceId,
			"plan-id": planId,
		})
		return err
	}

	logger.Info("release-new-install-success", lager.Data{
		"id": id,
		"name": name,
		"chart": chart,
		"chart-version": chartVersion,
		"serviceId": serviceId,
		"planId": planId,
	})

	return nil
}

func Exists(id string, client *helmi.Client, logger lager.Logger) (bool, error) {
	logger.Debug("release-exist-check",lager.Data{
		helmicons.InstanceIDLogKey: id,
	})
	name := getName(id)
	exists, err := client.ExistRelease(name)

	if err != nil {
		logger.Error("release-exist-check-error", err, lager.Data{
			"id": id,
			"name": name,
		})
	}
	return exists, err
}

func Delete(id string, client *helmi.Client, logger lager.Logger) error {
	logger.Debug("release-delete",lager.Data{
		helmicons.InstanceIDLogKey: id,
	})
	name := getName(id)
	err := client.DeleteRelease(name)
	if err != nil {
		exists, existsErr := client.ExistRelease(name)

		if existsErr == nil && !exists {
			logger.Info("release-instance-not-existed", lager.Data{
				"id": id,
				"name": name,
			})
			return nil
		}

		logger.Error("failed-delete-release", err, lager.Data{
			"id": id,
			"name": name,
		})

		return err
	}

	logger.Info("release-delete-success", lager.Data{
		"id": id,
		"name": name,
	})
	return nil
}

func GetStatus(id string, client *helmi.Client, logger lager.Logger) (Status, error) {
	logger.Debug("release-status",lager.Data{
		helmicons.InstanceIDLogKey: id,
	})
	name := getName(id)

	status, err := client.GetStatus(name)

	if err != nil {
		exists, existsErr := client.ExistRelease(name)

		if existsErr == nil && !exists {
			logger.Info("release-status-delete-already", lager.Data{
				"id": id,
				"name": name,
			})
			return Status{}, err
		}

		logger.Error("failed-get-release-status", err, lager.Data{
			"id": id,
			"name": name,
		})
		return Status{}, err
	}

	logger.Info("release-status-success", lager.Data{
		"id": id,
		"name": name,
	})

	return Status{
		IsFailed:    status.IsFailed,
		IsDeployed:  status.IsDeployed,
		IsAvailable: status.AvailableNodes >= status.DesiredNodes,
	}, nil
}

func GetCredentials(catalog *catalog.Catalog, serviceId string, planId string, id string, client *helmi.Client, logger lager.Logger) (map[string]interface{}, error) {
	logger.Debug("release-get-credentials",lager.Data{
		helmicons.InstanceIDLogKey: id,
	})
	name := getName(id)
	service, _ := catalog.GetService(serviceId)
	plan, _ := catalog.GetServicePlan(serviceId, planId)

	status, err := client.GetStatus(name)

	fmt.Println(status)

	if err != nil {
		exists, existsErr := client.ExistRelease(name)

		if existsErr == nil && !exists {
			logger.Info("release-asked-credentials-delete-already", lager.Data{
				"id": id,
				"name": name,
			})
			return nil, err
		}

		logger.Error("failed-get-release-status", err, lager.Data{
			"id": id,
			"name": name,
		})
		return nil, err
	}

	nodes, err := kubectl.GetNodes()

	if err != nil {
		logger.Error("failed-get-kubernetes-nodes", err, lager.Data{
			"id": id,
			"name": name,
		})
		return nil, err
	}

	values, err := client.GetReleaseValues(name)

	if err != nil {
		logger.Error("failed-get-helm-values", err, lager.Data{
			"id": id,
			"name": name,
		})
		return nil, err
	}

	credentials := getUserCredentials(service, plan, nodes, status, values)

	logger.Info("sending-release-credentials", lager.Data{
		"id": id,
		"name": name,
	})
	return credentials, nil
}

func getName(value string) string {
	const prefix = "helmi"

	if strings.HasPrefix(value, prefix) {
		return value
	}

	name := strings.ToLower(value)
	name = strings.Replace(name, "-", "", -1)
	name = strings.Replace(name, "_", "", -1)

	return prefix + name[:14]
}

func getChart(service catalog.CatalogService, plan catalog.CatalogPlan) (string, error) {
	if len(plan.Chart) > 0 {
		return plan.Chart, nil
	}

	if len(service.Chart) > 0 {
		return service.Chart, nil
	}

	return "", errors.New("no helm chart specified")
}

func getChartVersion(service catalog.CatalogService, plan catalog.CatalogPlan) (string, error) {
	if len(plan.ChartVersion) > 0 {
		return plan.ChartVersion, nil
	}

	if len(service.ChartVersion) > 0 {
		return service.ChartVersion, nil
	}

	return "", errors.New("no helm chart version specified")
}

//choice context namespace first
func getChartNamespace(context map[string]string, parameters map[string]string) string {
	for key, value := range context {
		if strings.EqualFold(key, "namespace") {
			return value
		}
	}
	for key, value := range parameters {
		if strings.EqualFold(key, "namespace") {
			return value
		}
	}
	return ""
}

func getChartValues(service catalog.CatalogService, plan catalog.CatalogPlan, parameters map[string]string) map[string]string {
	values := map[string]string{}
	templates := map[string]string{}

	for key, value := range service.ChartValues {
		templates[key] = value
	}

	for key, value := range plan.ChartValues {
		templates[key] = value
	}

	for key, value := range parameters {
		templates[key] = value
	}

	usernames := map[string]string{}
	passwords := map[string]string{}

	r := regexp.MustCompile(helmicons.LookupRegex)
	groupNames := r.SubexpNames()

	for key, template := range templates {
		value := r.ReplaceAllStringFunc(template, func(m string) string {
			var lookupType string
			var lookupPath string

			for groupKey, groupValue := range r.FindStringSubmatch(m) {
				groupName := groupNames[groupKey]

				if strings.EqualFold(groupName, helmicons.LookupRegexType) {
					lookupType = groupValue
				}

				if strings.EqualFold(groupName, helmicons.LookupRegexPath) {
					lookupPath = groupValue
				}
			}

			if strings.EqualFold(lookupType, helmicons.LookupUsername) {
				username := usernames[lookupPath]

				if len(username) == 0 {
					username = uuid.NewV4().String()
					username = strings.Replace(username, "-", "", -1)
					usernames[lookupPath] = username
				}

				return username
			}

			if strings.EqualFold(lookupType, helmicons.LookupPassword) {
				password := passwords[lookupPath]

				if len(password) == 0 {
					password = uuid.NewV4().String()
					password = strings.Replace(password, "-", "", -1)
					passwords[lookupPath] = password
				}

				return password
			}

			if strings.EqualFold(lookupType, helmicons.LookupEnv) {
				env, _ := os.LookupEnv(lookupPath)
				return env
			}

			return ""
		})

		if len(value) > 0 {
			values[key] = value
		}
	}

	return values
}

func getUserCredentials(service catalog.CatalogService, plan catalog.CatalogPlan, kubernetesNodes [] kubectl.Node, helmStatus helmi.Status, helmValues map[string]string) map[string]interface{} {
	values := map[string]interface{}{}
	templates := map[string]interface{}{}

	for key, value := range service.UserCredentials {
		templates[key] = value
	}

	for key, value := range plan.UserCredentials {
		templates[key] = value
	}

	r := regexp.MustCompile(helmicons.LookupRegex)
	groupNames := r.SubexpNames()

	replaceTemplate := func(template string) string {
		var lookupType string
		var lookupPath string

		for groupKey, groupValue := range r.FindStringSubmatch(template) {
			groupName := groupNames[groupKey]

			if strings.EqualFold(groupName, helmicons.LookupRegexType) {
				lookupType = groupValue
			}

			if strings.EqualFold(groupName, helmicons.LookupRegexPath) {
				lookupPath = groupValue
			}
		}

		if strings.EqualFold(lookupType, helmicons.LookupRelease) {
			if strings.EqualFold(lookupPath, "name") {
				return helmStatus.Name
			}
			if strings.EqualFold(lookupPath, "namespace") {
				return helmStatus.Namespace
			}
		}

		if strings.EqualFold(lookupType, helmicons.LookupUsername) {
			username := helmValues[lookupPath]
			return username
		}

		if strings.EqualFold(lookupType, helmicons.LookupPassword) {
			password := helmValues[lookupPath]
			return password
		}

		if strings.EqualFold(lookupType,helmicons.LookupValue) {
			value := helmValues[lookupPath]
			return value
		}

		if strings.EqualFold(lookupType, helmicons.LookupCluster) {
			if strings.HasPrefix(strings.ToLower(lookupPath), "port") {
				portParts := strings.Split(lookupPath, ":")

				if len(helmStatus.IngressPorts) > 0 {
					return strconv.Itoa(helmStatus.IngressPorts[0])
				}

				for clusterPort, nodePort := range helmStatus.NodePorts {
					if len(portParts) == 1 || strings.EqualFold(strconv.Itoa(clusterPort), portParts[1]) {
						return strconv.Itoa(nodePort)
					}
				}

				for containerPort, clusterPort := range  helmStatus.ClusterPorts {
					if len(portParts) == 1 || strings.EqualFold(strconv.Itoa(containerPort), portParts[1]) {
						return strconv.Itoa(clusterPort)
					}
				}

				return portParts[1]
			}

			// single host

			if strings.EqualFold(lookupPath, "address") {
				// return dns name if set as environment variable
				if value, ok := os.LookupEnv("DOMAIN"); ok {
					return value
				}

				if len(helmStatus.IngressHosts) > 0 {
					return helmStatus.IngressHosts[0]
				}

				if helmStatus.ServiceType == "ClusterIP" {
					if value, ok := os.LookupEnv("CLUSTER_DNS"); ok {
						return fmt.Sprintf("%s-%s.%s.%s", helmStatus.Name, service.Name, helmStatus.Namespace, value)
					}

				} else if helmStatus.ServiceType == "NodePort" {
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
				} else if helmStatus.ServiceType == "LoadBalancer" {
					//TODO
				} else if helmStatus.ServiceType == "ExternalName" {
					//TODO
				}
			}

			if strings.EqualFold(lookupPath, "hostname") {
				for _, node := range kubernetesNodes {
					if len(node.Hostname) > 0 {
						return node.Hostname
					}
				}
			}
		}

		return ""
	}

	for key, templateInterface := range templates {
		// string
		templateString, ok := reflect.ValueOf(templateInterface).Interface().(string)

		if ok {
			value := r.ReplaceAllStringFunc(templateString, replaceTemplate)

			if len(value) > 0 {
				values[key] = value
			}

			continue
		}

		// string array
		templateStringArray, ok := reflect.ValueOf(templateInterface).Interface().([]interface{})

		if ok {
			valueArray := []string{}

			for _, templateValue := range templateStringArray {
				templateString, ok := reflect.ValueOf(templateValue).Interface().(string)

				if ok {
					value := r.ReplaceAllStringFunc(templateString, replaceTemplate)

					if len(value) > 0 {
						valueArray = append(valueArray, value)
					}
				}
			}

			if len(valueArray) > 0 {
				values[key] = valueArray
			}

			continue
		}
	}

	newValues := make(map[string]interface{})
	for name, value := range values {
		if len(helmStatus.NodePorts) != 0 {
			if strings.Contains(value.(string), "|") && strings.Contains(value.(string), ">") {
				rs := substring(value.(string), "<", ">")
				v := strings.Replace(strings.Replace(rs, " |", "", -1), "| ", "", -1)
				newValues[name] = v
			} else {
				newValues[name] = value
			}
		}else {
			if strings.Contains(value.(string), "|") && strings.Contains(value.(string), ">") {
				rs := substring(value.(string), "|", "|")
				v := strings.Replace(strings.Replace(rs, "< ", "", -1), " >", "",-1)
				newValues[name] = v
			}else {
				newValues[name] = value
			}
		}
	}

	return newValues
}

func substring(s, begin, last string)  string {
	start := strings.Index(s, begin)
	end := strings.LastIndex(s, last)
	rs := strings.Replace(s, s[start:end + 1], "", -1)
	return rs
}