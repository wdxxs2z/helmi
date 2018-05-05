package helm

import (
	"os"
	"errors"

	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/getter"
	"code.cloudfoundry.org/lager"
	"k8s.io/helm/pkg/helm/helmpath"
	log "github.com/Sirupsen/logrus"
	"k8s.io/helm/pkg/helm/environment"
	"github.com/wdxxs2z/helmi/pkg/config"
)

type Repository struct {
	Name	string
	Url     string
}

func initRepos(env environment.EnvSettings, logger lager.Logger, config config.Config) error {
	logger.Debug("init-helm-repos", lager.Data{
		"repositorys": config.TillerConfig.Repos,
	})
	repositorys := config.TillerConfig.Repos
	if repositorys == nil {
		log.Debug("repository is null.")
		return errors.New("helm repository must not null")
	}
	for _, repository := range repositorys {
		err := addrepo(repository.Name, repository.Url, env, logger)
		if err != nil {
			return err
		}
	}
	return nil
}

func addrepo(name, url string, env environment.EnvSettings, logger lager.Logger) error {
	entry := repo.Entry{
		Name: 	name,
		URL:  	url,
		Cache:  env.Home.CacheIndex(name),
	}
	logger.Debug("add-helm-repos", lager.Data{
		"entrys": entry,
	})
	repo, err := repo.NewChartRepository(&entry, getter.All(env))
	if err != nil {
		log.Errorf("adding repository error: %s", err)
		return err
	}
	return handingRepos(repo, entry, env, logger)
}

func handingRepos(r *repo.ChartRepository, e repo.Entry, env environment.EnvSettings, logger lager.Logger) error {
	logger.Debug("handing-repos", lager.Data{
		"repofile": env.Home.RepositoryFile(),
	})
	err := r.DownloadIndexFile("")
	if err != nil {
		log.Errorf("downloading repository error: %s", err)
		return err
	}
	_, err = os.Stat(env.Home.RepositoryFile())
	if err != nil {
		err = addRepoFile(env.Home.RepositoryFile(), e)
		if err != nil{
			log.Debugf("add repository file error: %s", err)
			return err
		}
	}
	return updateRepoFile(env.Home.RepositoryFile(), e)
}

func addRepoFile(file string, e repo.Entry) error {
	f := repo.NewRepoFile()
	f.Add(&e)
	log.Debugf("Writing repository file %s", file)
	return f.WriteFile(file, 0644)
}

func updateRepoFile(file string, e repo.Entry) error {
	f, err := repo.LoadRepositoriesFile(file)
	if err != nil {
		log.Errorf("updating the repo file err: %s", err)
		return err
	}
	f.Update(&e)
	return f.WriteFile(file, 0644)
}

func handingHelmDirectors(home helmpath.Home) error{
	helmDirectories := []string{
		home.Repository(),
		home.Plugins(),
		home.Starters(),
		home.Cache(),
		home.Archive(),
	}
	for _, dir := range helmDirectories {
		err := handingDirectory(dir)
		if err != nil {
			log.Errorf("handing the %s director error: %s", dir, err)
			return err
		}
	}
	return nil
}

func handingDirectory(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0744)
		if err != nil {
			return err
		}
	}
	return nil
}