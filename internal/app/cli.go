package app

import (
	"log"
	"os"
	"strconv"

	"github.com/olebedev/config"
	"gopkg.in/urfave/cli.v1"
)

type AbuseMeshApp struct {
	cli    *cli.App
	config *config.Config
}

func NewAbuseMeshCLI() *AbuseMeshApp {
	abuseMeshApp := &AbuseMeshApp{
		config: &config.Config{},
	}

	app := cli.NewApp()
	app.Name = "Abuse mesh"
	app.Commands = []cli.Command{
		{
			Name:   "start-server",
			Usage:  "Start the Abuse mesh server",
			Action: abuseMeshApp.runServer,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "host",
					Usage: "The hostname/IP address the Abuse Mesh server will listen on",
				},
				cli.StringFlag{
					Name:  "port",
					Usage: "The TCP the Abuse Mesh server will listen on",
				},
			},
		},
	}

	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Usage: "The config file which will be used by the Abuse Mesh server",
		},
	}

	app.Before = abuseMeshApp.loadConfig
	app.Flags = flags

	abuseMeshApp.cli = app

	return abuseMeshApp
}

//If the config flag is set and the file exists try to load the config
func (app *AbuseMeshApp) loadConfig(context *cli.Context) error {
	if context.IsSet("config") {
		file := context.String("config")
		if _, err := os.Stat(file); err == nil {
			config, err := config.ParseYamlFile(file)
			if err != nil {
				app.config = config
			} else {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (app *AbuseMeshApp) getSettingInt(context *cli.Context, name string, defaultValue int) int {
	if context.IsSet(name) {
		return context.Int(name)
	}

	if setting, err := app.config.Int(name); err == nil {
		return setting
	}

	return defaultValue
}

func (app *AbuseMeshApp) getSettingString(context *cli.Context, name string, defaultValue string) string {
	if context.IsSet(name) {
		return context.String(name)
	}

	if setting, err := app.config.String(name); err == nil {
		return setting
	}

	return defaultValue
}

func (app *AbuseMeshApp) runServer(context *cli.Context) error {
	server := NewServer()

	host := app.getSettingString(context, "host", "")
	port := app.getSettingInt(context, "port", 80)

	if port < 1 || port > 65535 {
		log.Fatal("Port number must be between 1 and 65535")
	}

	server.webserver.Addr = host + ":" + strconv.Itoa(port)

	log.Println("Starting Abuse Mesh server on " + server.webserver.Addr)

	return server.webserver.ListenAndServe()
}
