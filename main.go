package main

import (
	"os"
	"log"
	"flag"
	"strings"

	"code.cloudfoundry.org/lager"

	"github.com/wdxxs2z/helmi/pkg/broker"
	helmi "github.com/wdxxs2z/helmi/pkg/helm"
	"github.com/wdxxs2z/helmi/pkg/catalog"
)

var (
	configpath 	string
	port           	string

	logLevels = map[string]lager.LogLevel{
		"DEBUG": lager.DEBUG,
		"INFO":  lager.INFO,
		"ERROR": lager.ERROR,
		"FATAL": lager.FATAL,
	}
)

func init() {
	flag.StringVar(&configpath, "config", "", "The helmi config path")
	flag.StringVar(&port, "port", "8080", "Listen port")
}

func buildLogger(logLevel string) lager.Logger {
	laggerLogLevel, ok := logLevels[strings.ToUpper(logLevel)]
	if !ok {
		log.Fatal("Invalid log level: ", logLevel)
	}

	logger := lager.NewLogger("helmi-service-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, laggerLogLevel))

	return logger
}

func main() {
	flag.Parse()
	config, err := LoadConfig(configpath)

	if err != nil {
		log.Fatalf("Error loading config file: %s", err)
	}

	logger := buildLogger(config.LogLevel)

	c, catalogErr := catalog.ParseDir(config.HelmiConfig.CatalogDir)
	if catalogErr != nil {
		log.Fatal(err)
	}

	os.Setenv("USERNAME", config.Username)
	os.Setenv("PASSWORD", config.Password)

	os.Setenv("CLUSTER_DNS", config.HelmiConfig.ClusterDnsName)

	helmClient := helmi.NewClient(config.HelmiConfig, logger)

	helmibroker := broker.New(config.HelmiConfig, c, helmClient, logger)

	helmibroker.Run(":" + port)
}