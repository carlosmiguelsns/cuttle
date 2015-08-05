package main

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/cloudflare/conf"
	"github.com/elazarl/goproxy"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{})

	config, err := conf.ReadConfigFile("cuttle.conf")
	if err != nil {
		log.Error("Failed to load config from 'cuttle.conf'.")
		log.Fatal(err)
	}

	// Config limit controller.
	var controller LimitController
	control := config.GetString("limitcontrol", "rps")
	if control == "rps" {
		limit := config.GetUint("rps-limit", 2)
		controller = &RPSController{
			Limit: limit,
		}
	} else {
		log.Fatal("Unknown limit control: ", control)
	}

	// Config proxy.
	addr := config.GetString("addr", ":8123")
	verbose := config.GetUint("verbose", 0)
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = verbose == 1

	// Starts now.
	controller.Start()

	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			// Acquire permission to forward request to downstream.
			controller.Acquire()

			return r, nil // Forward request.
		})

	log.Fatal(http.ListenAndServe(addr, proxy))
}
