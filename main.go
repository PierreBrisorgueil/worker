package main

import (
	"context"
	"fmt"
	"os"

	"github.com/namsral/flag"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/waitfor/pkg/runner"
	waitrunner "github.com/ziflex/waitfor/pkg/runner"

	"github.com/MontFerret/worker/internal/controllers"
	"github.com/MontFerret/worker/internal/server"
	"github.com/MontFerret/worker/pkg/worker"
)

var (
	version string

	ferretVersion string

	port = flag.Uint64("port", 8080, "port to listen")

	chromeIP = flag.String("chrome-ip", "127.0.0.1", "Google Chrome remote IP address")

	chromeDebuggingPort = flag.Uint64("chrome-port", 9222, "Google Chrome remote debugging port")

	showVersion = flag.Bool(
		"version",
		false,
		"show version",
	)

	help = flag.Bool(
		"help",
		false,
		"show this list",
	)
)

func main() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
		return
	}

	if *showVersion {
		fmt.Println(fmt.Sprintf("Worker: %s", version))
		fmt.Println(fmt.Sprintf("Ferret: %s", ferretVersion))
		os.Exit(0)
		return
	}

	cdp := worker.CDPSettings{
		Host: *chromeIP,
		Port: *chromeDebuggingPort,
	}

	if err := waitForChrome(cdp); err != nil {
		log.Fatal().
			Err(err).
			Msg("wait for Chrome")
	}

	srv, err := server.New()

	if err != nil {
		panic(err)
	}

	if err := setupControllers(srv, cdp); err != nil {
		panic(err)
	}

	err = srv.Run(*port)

	log.Err(err).
		Timestamp().
		Msg("listen and server")
}

func waitForChrome(cdp worker.CDPSettings) error {
	return waitrunner.Test(context.Background(), []string{
		cdp.BaseURL(),
	}, runner.WithAttempts(10))
}

func setupControllers(server *server.Server, cdp worker.CDPSettings) error {
	workerCtl, err := controllers.NewWorker(cdp)

	if err != nil {
		return err
	}

	workerCtl.Use(server.Router())

	healthCtl, err := controllers.NewHealth(cdp)

	if err != nil {
		return err
	}

	healthCtl.Use(server.Router())

	infoCtl, err := controllers.NewInfo(controllers.InfoSettings{
		Version:       version,
		FerretVersion: ferretVersion,
		CDP:           cdp,
	})

	if err != nil {
		return err
	}

	infoCtl.Use(server.Router())

	return nil
}
