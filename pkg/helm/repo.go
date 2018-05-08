package helm

import (
	"os"

	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/getter"
	"code.cloudfoundry.org/lager"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/helm/environment"
	"github.com/wdxxs2z/helmi/pkg/config"
	"fmt"
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
		return fmt.Errorf("helm repository must not null, please config the tillerconfig repos.")
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
		return err
	}
	return handingRepos(repo, entry, env, logger)
}

func handingRepos(r *repo.ChartRepository, e repo.Entry, env environment.EnvSettings, logger lager.Logger) error {
	logger.Debug("handing-repos", lager.Data{
		"repofile": env.Home.RepositoryFile(),
	})

	downloadErr := r.DownloadIndexFile("")
	if downloadErr != nil {
		return fmt.Errorf("download repo index file cause an error: %s", downloadErr)
	}
	_, err := os.Stat(env.Home.RepositoryFile())
	if err != nil {
		adderr := addRepoFile(env.Home.RepositoryFile(), e)
		if adderr != nil{
			return adderr
		}
	}
	return updateRepoFile(env.Home.RepositoryFile(), e, logger)
}

func addRepoFile(file string, e repo.Entry) error {
	f := repo.NewRepoFile()
	f.Add(&e)
	return f.WriteFile(file, 0644)
}

func updateRepoFile(file string, e repo.Entry, logger lager.Logger) error {
	logger.Debug("update-repo-file", lager.Data{
		"file": file,
	})
	f, err := repo.LoadRepositoriesFile(file)
	if err != nil {
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
			return fmt.Errorf("handing helm directors cause an error: %s", err)
		}
	}
	return nil
}

func handingDirectory(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0744)
		if err != nil {
			return fmt.Errorf("handing director %s cause an error: %s", dir, err)
		}
	}
	return nil
}