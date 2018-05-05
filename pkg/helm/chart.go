package helm

import (
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
	"github.com/wdxxs2z/helmi/pkg/config"
	"code.cloudfoundry.org/lager"
	"fmt"
)

func getChart(config config.Config, envs environment.EnvSettings, chartName string, chartVersion string, logger lager.Logger) (*chart.Chart, error) {
	logger.Debug("get-chart", lager.Data{
		"chart-name": chartName,
		"chart-version": chartVersion,
	})
	err := handingHelmDirectors(envs.Home)
	if err != nil {
		return nil, err
	}

	repos, err := getRepos(config, envs)
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

func getRepos(config config.Config, envs environment.EnvSettings) (map[string]string, error) {
	repos := config.TillerConfig.Repos
	rmap := make(map[string]string)
	for _, repo := range repos {
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
	var err error
	var exist bool = false
	for _, url := range repos {
		curl, err := repo.FindChartInRepoURL(url, chartName, chartVersion, "", "", "", getter.All(envs))
		if err != nil {
			log.Debugf("get the %s chart in %s, but meet an error: %s", chartName, url, err)
		} else {
			chartUrl = curl
			exist = true
		}
	}

	if exist {
		log.Debugf("Chart URL found: %s", chartUrl)
		return downloadChart(chartUrl, chartVersion, envs, logger)
	}else {
		if err != nil {
			log.Errorf("Cause an error: %s", err)
			return "", err
		}else {
			log.Errorf("The chart not found in all repos.")
			return "", fmt.Errorf("Chart %s not found in all repos", chartName)
		}
	}
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

	log.Debugf("Downloading chart %s to %s", url, envs.Home.Archive())
	filename, _, err := dl.DownloadTo(url, version, envs.Home.Archive())
	if err != nil {
		return "", err
	}
	log.Debugf("Downloaded chart from URL %s to %s",url, filename)
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

	log.Debugf("Loading chart from %s", lname)
	chartRequested, err := chartutil.Load(lname)
	if err != nil {
		return nil, err
	}
	log.Infof("Loaded chart from %s", lname)
	return chartRequested, nil
}