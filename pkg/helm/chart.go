package helm

import (
	"os"
	"path/filepath"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
	"github.com/wdxxs2z/helmi/pkg/config"
	"code.cloudfoundry.org/lager"
	"fmt"
	"strings"
)

func getChart(config config.Config, envs environment.EnvSettings, chartName string, chartVersion string, chartOffline string, logger lager.Logger) (*chart.Chart, error) {
	logger.Debug("get-chart", lager.Data{
		"chart-name": chartName,
		"chart-version": chartVersion,
	})
	err := handingHelmDirectors(envs.Home)
	if err != nil {
		return nil, fmt.Errorf("get chart cause an error: %s", err)
	}

	exist, chartfile, err := findChartWithPath(chartName, chartVersion, chartOffline, envs, logger)
	if exist {
		helmChart, err := loadChart(chartfile, logger)
		if err != nil {
			return nil, err
		}
		return helmChart, nil
	} else {
		logger.Error("find-chart-from-local", err, lager.Data{
			"start-download-chart-from-remote-repo": true,
		})
		repos, err := getRepos(config, envs, logger)
		if err != nil {
			return nil, err
		}
		chartFile, err := writeChart(config, envs, chartName, chartVersion, repos, logger)
		if err != nil {
			return nil, err
		}
		helmChart, err := loadChart(chartFile, logger)
		if err != nil {
			return nil, err
		}
		return helmChart, nil
	}
}

func getRepos(config config.Config, envs environment.EnvSettings, logger lager.Logger) (map[string]string, error) {
	logger.Debug("chart-get-repos", lager.Data{
		"repos": config.TillerConfig.Repos,
	})
	repos := config.TillerConfig.Repos
	rmap := make(map[string]string)
	for _, repo := range repos {
		err := addrepo(repo.Name, repo.Url, envs, logger)
		if err != nil {
			return nil, err
		}
		rmap[repo.Name] = repo.Url
	}
	return rmap, nil
}

func writeChart(config config.Config,
		envs environment.EnvSettings,
		chartName string,
		chartVersion string,
		repos map[string]string,
		logger lager.Logger) (string, error) {

	logger.Debug("write-chart", lager.Data{
		"chart-name": chartName,
		"chart-version": chartVersion,
		"repos": repos,
	})

	var chartUrl string
	var exist bool = false

	if strings.Contains(chartName, "/") {
		repoName := strings.Split(chartName, "/")[0]
		realChart := strings.Split(chartName, "/")[1]
		for name, url := range repos {
			if repoName == name {
				chartExist, curl, err := findChartWithUrl(realChart, chartVersion, url, envs)
				if err != nil {
					return "", err
				} else if chartExist == true {
					chartUrl = curl
					exist = true
				}
			}
		}
	} else {
		return "", fmt.Errorf("the chart name must contain repo name and chart name, such as stabel/%s", chartName)
	}

	if exist {
		return downloadChart(chartUrl, chartVersion, envs, logger)
	} else {
		return "", fmt.Errorf("Chart %s not found in all repos", chartName)
	}
}

func findChartWithUrl(chartName, chartVersion, repoUrl string, envs environment.EnvSettings) (bool, string, error) {
	curl, err := repo.FindChartInRepoURL(repoUrl, chartName, chartVersion, "", "", "", getter.All(envs))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, "", nil
		} else {
			return false, "", err
		}
	} else {
		return true, curl, nil
	}
}

func findChartWithPath(chartName, chartVersion, chartfile string, envs environment.EnvSettings, logger lager.Logger) (bool, string, error) {
	logger.Debug("find-local-chart", lager.Data{
		"chart-name": chartName,
		"chart-version": chartVersion,
		"chart-file": chartfile,
	})
	var imagefile string
	if chartfile == "" {
		imagefile = fmt.Sprintf("%s/%s-%s.tgz", envs.Home.Archive(), parseChartName(chartName), chartVersion)
	} else {
		if (strings.Contains(chartfile, fmt.Sprintf("%s-%s.tgz", parseChartName(chartName), chartVersion))) {
			imagefile = chartfile
		} else {
			imagefile = fmt.Sprintf("%s/%s-%s.tgz", envs.Home.Archive(), parseChartName(chartName), chartVersion)
		}
	}
	_, err := os.Stat(imagefile)
	if err != nil {
		return false, "", err
	}
	return true, imagefile, nil
}

func downloadChart(url string, version string, envs environment.EnvSettings, logger lager.Logger) (string, error) {
	logger.Debug("download-chart", lager.Data{
		"chart-url": url,
		"char-version": version,
	})

	dl := downloader.ChartDownloader{
		HelmHome: envs.Home,
		Out:      os.Stdout,
		Getters:  getter.All(envs),
		Verify:   downloader.VerifyIfPossible,
	}

	filename, _, err := dl.DownloadTo(url, version, envs.Home.Archive())
	if err != nil {
		return "", err
	}
	logger.Debug("download-chart-success",lager.Data{"filename": filename})
	return filename, nil
}

func loadChart(filename string, logger lager.Logger) (*chart.Chart, error) {
	logger.Debug("load-chart", lager.Data{
		"filename": filename,
	})

	lname, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	chartRequested, err := chartutil.Load(lname)
	if err != nil {
		return nil, fmt.Errorf("load chart cause error: %s", err)
	}
	logger.Debug("load-chart-success", lager.Data{"filename": filename})
	return chartRequested, nil
}

func parseChartName(name string) string {
	s := strings.Split(name, "/")
	if len(s) == 1 {
		return s[0]
	} else {
		return s[1]
	}
}