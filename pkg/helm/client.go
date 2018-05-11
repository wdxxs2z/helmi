package helm

import (
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"github.com/wdxxs2z/helmi/pkg/config"
	"fmt"
	"k8s.io/helm/pkg/helm/environment"
	"code.cloudfoundry.org/lager"
	"k8s.io/helm/pkg/kube"
	"github.com/wdxxs2z/helmi/pkg/utils"
	rspb "k8s.io/helm/pkg/proto/hapi/release"
	"time"
	"k8s.io/helm/pkg/timeconv"
	"k8s.io/helm/pkg/chartutil"
	"github.com/kylelemons/go-gypsy/yaml"
	"bytes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/helm/pkg/helm/helmpath"
	"github.com/wdxxs2z/helmi/pkg/kubectl"
	"strings"
)

type Client struct {
	helm           	*helm.Client
	env       	environment.EnvSettings
	config          config.Config
	logger		lager.Logger
}

func NewClient (config config.Config, logger lager.Logger) *Client {
	sessionLogger := logger.Session("helm-client")

	helmEnv := getHelmEnvironment(config)
	err := handingHelmDirectors(helmEnv.Home)
	if err != nil {
		sessionLogger.Error("handing-helm-directors", err, lager.Data{})
		return nil
	}
	err = initRepos(helmEnv, sessionLogger, config)
	if err != nil {
		sessionLogger.Error("init-helm-repo", err, lager.Data{})
		return nil
	}
	helmClient, err := getHelmClient(config)
	if err != nil {
		sessionLogger.Error("create-helm-client", err, lager.Data{})
		return nil
	}
	return &Client{
		helm:		helmClient,
		env:		helmEnv,
		logger:		sessionLogger,
		config:		config,
	}
}

func (c *Client) TillerCheck() error {
	return c.helm.PingTiller()
}

func (c *Client) ExistRelease(release string) (bool, error) {
	c.logger.Debug("exist release", lager.Data{
		"check-release-exist": release,
	})
	statusRes, err := c.helm.ReleaseStatus(release)
	if err != nil {
		return false, fmt.Errorf("check release cause an error: %s", err)
	}
	if statusRes == nil {
		return false, nil
	}
	return true, nil
}

func (c *Client) InstallRelease(release string, chartName string, version string, chartOffline string, values map[string]string, namespace string, acceptsIncomplete bool) (*rspb.Release, error) {
	displayValues := make(map[string]string)
	for name, value := range values {
		if strings.Contains(name, "Password") || strings.Contains(name, "password") {
			displayValues[name] = "hidden"
		} else {
			displayValues[name] = value
		}
	}
	c.logger.Debug("install-release", lager.Data{
		"release-name": release,
		"chart-name": chartName,
		"version": version,
		"values": displayValues,
		"namespace": namespace,
	})

	var wait bool = false

	if acceptsIncomplete == false {
		wait = true
	}

	rawValues, err := utils.ConvertInterfaceToByte(values)
	if err != nil {
		return nil, fmt.Errorf("convert values to yaml values cause an error: %s", err)
	}

	chart, err := getChart(c.config, c.env, chartName, version, chartOffline,c.logger)

	if err != nil {
		return nil, fmt.Errorf("install release %s, cause an error: %s", chartName, err)
	}

	res, err := c.helm.InstallReleaseFromChart(chart, namespace, installOpts(release, wait, rawValues)...)
	if res == nil || res.Release == nil {
		rls ,err := c.getRelease(release)
		if err != nil {
			return nil, fmt.Errorf("get the release cause an error: %s", err)
		}
		if rls != nil {
			return rls, nil
		}
	}else {
		return res.Release, nil
	}
	return nil, err
}

func (c *Client) DeleteRelease(release string) error {
	_, err := c.helm.DeleteRelease(release, deleteOpts()...)
	if err != nil {
		return fmt.Errorf("delete release cause an error: %s", err)
	}
	return nil
}

func (c *Client) GetReleaseValues (release string) (map[string]string, error) {
	res , err := c.helm.ReleaseContent(release)
	if err != nil {
		return nil, err
	}
	cfg, err := chartutil.CoalesceValues(res.Release.Chart, res.Release.Config)
	content, err := cfg.YAML()
	if err != nil{
		return nil, err
	}
	node, err := yaml.Parse(bytes.NewReader([]byte(content)))
	if err != nil {
		return nil, err
	}
	properties := c.readYamlProperties(node, "")
	return properties, nil
}

func (c *Client) GetStatus(release string) (Status, error) {
	status, err := c.helm.ReleaseStatus(release)
	if  err != nil {
		return Status{}, err
	}
	name := status.GetName()
	namespace := status.GetNamespace()
	status_code := status.GetInfo().GetStatus().GetCode()
	resources := status.GetInfo().GetStatus().GetResources()

	loc, _ := time.LoadLocation("Local")
	lastDeploymentTime, _ := time.ParseInLocation(time.ANSIC, timeconv.String(status.Info.LastDeployed), loc)

	var deployed bool = false
	if status_code == 1 {
		deployed = true
	}
	s, err := convertByteToStatus(name, namespace, lastDeploymentTime, deployed, []byte(resources))
	if err != nil {
		return Status{}, err
	}
	return s, nil
}

func (c *Client) getRelease(release string) (*rspb.Release, error)  {
	releases , err := c.helm.ListReleases(listOpts(release)...)
	if err != nil {
		return nil, err
	}
	if releases.Count < 1 {
		return nil, nil
	} else if releases.Count >1 {
		return nil, fmt.Errorf("Error in multi releases exist for release %s", release)
	}
	return releases.Releases[0], nil
}

func getHelmEnvironment(config config.Config) environment.EnvSettings {
	var envs environment.EnvSettings
	envs.TillerHost = config.TillerConfig.Host
	envs.TillerNamespace = config.TillerConfig.Namespace
	envs.TillerConnectionTimeout = config.TillerConfig.ConnectionTimeout
	if config.TillerConfig.Home != "" {
		envs.Home = helmpath.Home(config.TillerConfig.Home)
	}
	return envs
}

func getHelmClient(config config.Config) (*helm.Client, error){
	var tillerHost string

	if config.TillerConfig.Host != "" {
		tillerHost = config.TillerConfig.Host
		return helm.NewClient(helm.Host(tillerHost)), nil
	}else {
		kubeclient, kubeconfig, err := kubectl.GetKubeClient()
		if err != nil {
			return nil, err
		}
		var tunnel *kube.Tunnel
		if config.TillerConfig.Namespace != "" {
			tunnel, err = portforwarder.New(config.TillerConfig.Namespace, kubeclient, kubeconfig)
			if err != nil {
				return nil, err
			}
		}else {
			tunnel, err = portforwarder.New("kube-system", kubeclient, kubeconfig)
			if err != nil {
				return nil, err
			}
		}
		tillerHost := fmt.Sprintf("127.0.0.1:%d", tunnel.Local)
		hclient := helm.NewClient(helm.Host(tillerHost), helm.ConnectTimeout(30))
		if hclient != nil {
			return hclient, nil
		} else {
			return nil, err
		}
		return hclient, nil
	}
}

func (c *Client) readYamlProperties(node yaml.Node, prefix string) map[string]string {
	values := map[string]string{}

	switch n := node.(type) {
	case yaml.Map:
		for mapKey, mapNode := range n {
			nodeName := prefix

			if len(nodeName) > 0 {
				nodeName += "."
			}

			nodeName += mapKey

			for key, value := range c.readYamlProperties(mapNode, nodeName) {
				values[key] = value
			}
		}
	case yaml.Scalar:
		value := n.String()
		values[prefix] = value
	}

	return values
}