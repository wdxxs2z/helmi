package main

import (

)
import (
	"github.com/wdxxs2z/helmi/pkg/kubectl"
	"log"
	"os"
)

func main() {
	info, err := kubectl.CheckVersion()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	log.Fatalf("Server Info: %s", info)
	os.Exit(0)
}
