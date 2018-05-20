package release

import (
	"errors"
	"strings"
	helmi "github.com/wdxxs2z/helmi/pkg/helm"
	"github.com/wdxxs2z/helmi/pkg/kubectl"
	"github.com/wdxxs2z/helmi/pkg/catalog"
	helmicons "github.com/wdxxs2z/helmi/pkg/constants"
	"code.cloudfoundry.org/lager"
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
	chartValues, chartValuesErr := getChartValues(service, plan, parameters)
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

	if chartValuesErr != nil {
		logger.Error("failed-parse-chart-values", chartValuesErr, lager.Data{
			"id": id,
			"name": name,
			"service-id": serviceId,
			"plan-id": planId,
		})
	}

	_, err := client.InstallRelease(name, chart, chartVersion, service.ChartOffline, chartValues, chartNamespace, acceptsIncomplete)

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

func Update(id string, catalog *catalog.Catalog, serviceId string, planId string, client *helmi.Client, acceptsIncomplete bool, parameters map[string]string, context map[string]string,logger lager.Logger) error {
	logger.Debug("release-update",lager.Data{
		helmicons.InstanceIDLogKey: id,
	})

	name := getName(id)
	service, _ := catalog.GetService(serviceId)
	plan, _ := catalog.GetServicePlan(serviceId, planId)

	chartName, chartErr := getChart(service, plan)
	chartVersion, chartVersionErr := getChartVersion(service, plan)
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

	_, err := client.UpdateRelease(name, chartName, chartVersion, service.ChartOffline, parameters, chartNamespace, acceptsIncomplete)

	if err != nil {
		logger.Error("failed-upgrade-release", err, lager.Data{
			"id": id,
			"name": name,
			"chart": chartName,
			"chart-version": chartVersion,
			"service-id": serviceId,
			"plan-id": planId,
		})
		return err
	}

	logger.Info("release-upgrade-success", lager.Data{
		"id": id,
		"name": name,
		"chart": chartName,
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

	//values, err := client.GetReleaseValues(name)
	values, err := client.ParseReleaseValues(name)

	if err != nil {
		logger.Error("failed-get-helm-values", err, lager.Data{
			"id": id,
			"name": name,
		})
		return nil, err
	}

	credentials, err := service.UserCredentials(&plan, nodes, status, values)
	if err != nil {
		logger.Error("failed-get-usercredentials", err, lager.Data{
			"id": id,
			"name": name,
		})
	}

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

func getChart(service catalog.Service, plan catalog.Plan) (string, error) {
	if len(plan.Chart) > 0 {
		return plan.Chart, nil
	}

	if len(service.Chart) > 0 {
		return service.Chart, nil
	}

	return "", errors.New("no helm chart specified")
}

func getChartVersion(service catalog.Service, plan catalog.Plan) (string, error) {
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

func getChartValues(service catalog.Service, plan catalog.Plan, parameters map[string]string) (map[string]string, error) {
	templates := map[string]string{}

	chartValues , err := service.ChartValues(&plan)
	if err != nil {
		return nil, err
	}

	for key, value := range chartValues {
		templates[key] = value
	}

	for key, value := range parameters {
		templates[key] = value
	}

	return templates, nil
}