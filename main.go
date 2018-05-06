package main

import (
	"os"
	"log"
	"flag"
	"strings"

	"code.cloudfoundry.org/lager"

	"github.com/wdxxs2z/helmi/pkg/broker"
	helmi "github.com/wdxxs2z/helmi/pkg/helm"
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
	flag.StringVar(&port, "port", "5000", "Listen port")
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

	os.Setenv("USERNAME", config.Username)
	os.Setenv("PASSWORD", config.Password)

	helmClient := helmi.NewClient(config.HelmiConfig, logger)

	if helmClient == nil {
		log.Fatalf("please check your internet and cache the repo index, or the kubernetes tiller server is ok.")
		return
	}

	helmibroker := broker.New(config.HelmiConfig, helmClient, logger)

	helmibroker.Run(":" + port)
}