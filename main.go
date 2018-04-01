package main

import (
	"log"
	"monitor/client"
	"monitor/daemon"
	"os"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "monitorer"
	app.Usage = "Monitor websites from the CLI"

	app.Commands = []cli.Command{
		agent.Start,
		client.Show,
		client.Stop,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
