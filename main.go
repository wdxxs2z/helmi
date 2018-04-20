package main

import (
	"os"
	"log"
	"fmt"
	"flag"
	"strings"
	"net/http"

	"gopkg.in/yaml.v2"
	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"

	"github.com/wdxxs2z/helmi/pkg/broker"
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
	flag.StringVar(&port, "port", "3000", "Listen port")
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

	if config.Platform == "kubernetes" {
		runAsKubernetes(config)
	}

	if config.Platform == "cloudfoundry" {
		runAsCloudfoundry(config)
	}

}

func runAsKubernetes(config *Config) {
	a := App{}

	out ,err := yaml.Marshal(config.HelmiConfig.Catalog)

	if err != nil {
		log.Fatalf("Error parse helmi yaml config: %s", err)
	}

	a.Initialize(out)
	a.Run(":" + port)
}

func runAsCloudfoundry(config *Config) {

	logger := buildLogger(config.LogLevel)

	helmibroker := broker.New(config.HelmiConfig, logger)

	credentials := brokerapi.BrokerCredentials{
		Username: config.Username,
		Password: config.Password,
	}

	brokerApi := brokerapi.New(helmibroker, logger, credentials)
	http.Handle("/", brokerApi)

	fmt.Println("Helm Service Broker started on port " + port + "...")
	http.ListenAndServe(":"+port, nil)
}