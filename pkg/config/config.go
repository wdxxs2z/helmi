package config

type Config struct {
	AllowUserProvisionParameters bool    		`yaml:"allow_user_provision_parameters"`
	AllowUserUpdateParameters    bool    		`yaml:"allow_user_update_parameters"`
	AllowUserBindParameters      bool               `yaml:"allow_user_bind_parameters"`
	Db                           Database		`yaml:"db"`
	ClusterDnsName		     string             `yaml:"cluster_dns_name"`
	TillerConfig                 TillerSet		`yaml:"tille_config"`
	CatalogDir                   string             `yaml:"catalog_dir"`
}

type Database struct {
	Endpoints               []string                `yaml:"db_endpoints"`
	DialTimeout             int                     `yaml:"db_dial_timeout"`
	DbName                  string                  `yaml:"db_name"`
	CaCert                  string                  `yaml:"db_ca_cert"`
	TLSKey                  string                  `yaml:"db_tls_key"`
	TLSCert                 string                  `yaml:"db_tls_cert"`
}

type TillerSet struct {
	Host			string			`yaml:"host"`
	Namespace		string          	`yaml:"namespace"`
	Home			string			`yaml:"home"`
	ConnectionTimeout       int64               	`yaml:"connection_timeout"`
	Repos           	[]Repository		`yaml:"repos"`
}

type Repository struct {
	Name	string		`yaml:"name"`
	Url     string		`yaml:"url"`
}