package config

import (
	"github.com/wdxxs2z/helmi/pkg/catalog"
)

type Config struct {
	AllowUserProvisionParameters bool    		`yaml:"allow_user_provision_parameters"`
	AllowUserUpdateParameters    bool    		`yaml:"allow_user_update_parameters"`
	AllowUserBindParameters      bool               `yaml:"allow_user_bind_parameters"`
	Catalog			     catalog.Catalog	`yaml:"catalog"`
}