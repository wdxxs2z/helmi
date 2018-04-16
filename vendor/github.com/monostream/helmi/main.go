package main

import (
	"os"
	"path/filepath"
)

func main() {
	a := App{}

	path, _ := filepath.Abs("./catalog.yaml")

	port := os.Getenv("PORT")

	if len(port) == 0 {
		port = "5000"
	}

	a.Initialize(path)
	a.Run(":" + port)
}
