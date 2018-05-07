package config

import (
	"github.com/wdxxs2z/helmi/pkg/catalog"
)

type Config struct {
	AllowUserProvisionParameters bool    		`yaml:"allow_user_provision_parameters"`
	AllowUserUpdateParameters    bool    		`yaml:"allow_user_update_parameters"`
	AllowUserBindParameters      bool               `yaml:"allow_user_bind_parameters"`
	TillerConfig                 TillerSet		`yaml:"tille_config"`
	Catalog			     catalog.Catalog	`yaml:"catalog"`
}

type TillerSet struct {
	Host			string			`yaml:"host"`
	Namespace		string          	`yaml:"namespace"`
	ConnectionTimeout 	int64			`yaml:"connection_timeout"`
	Repos           	[]Repository		`yaml:"repos"`
}

type Repository struct {
	Name	string		`yaml:"name"`
	Url     string		`yaml:"url"`
}