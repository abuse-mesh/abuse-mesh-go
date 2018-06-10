package app

import (
	"log"
	"os"
)

func Main() {
	app := NewAbuseMeshCLI()

	err := app.cli.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
