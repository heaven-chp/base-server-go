package main

import (
	"errors"
	"flag"
	net_http "net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/heaven-chp/base-server-go/config"
	"github.com/heaven-chp/base-server-go/http-server/handler"
	"github.com/heaven-chp/base-server-go/http-server/log"
	"github.com/heaven-chp/base-server-go/http-server/swagger_docs"
	command_line_flag "github.com/heaven-chp/common-library-go/command-line/flag"
	"github.com/heaven-chp/common-library-go/http"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Main struct {
	server           http.Server
	httpServerConfig config.HttpServer
}

func (this *Main) initialize() error {
	if err := this.parseFlag(); err != nil {
		return err
	} else if err := this.setConfig(); err != nil {
		return err
	} else {
		log.Initialize(this.httpServerConfig)
		this.setSwaggerInfo()
		this.setHandler()

		return nil
	}
}

func (this *Main) parseFlag() error {
	flagInfos := []command_line_flag.FlagInfo{
		{FlagName: "config_file", Usage: "config/HttpServer.config", DefaultValue: string("")},
	}

	if err := command_line_flag.Parse(flagInfos); err != nil {
		return nil
	} else if flag.NFlag() != 1 {
		flag.Usage()
		return errors.New("invalid flag")
	} else {
		return nil
	}
}

func (this *Main) setConfig() error {
	fileName := command_line_flag.Get[string]("config_file")

	if httpServerConfig, err := config.Get[config.HttpServer](fileName); err != nil {
		return err
	} else {
		this.httpServerConfig = httpServerConfig
		return nil
	}
}

func (this *Main) setSwaggerInfo() {
	swagger_docs.SwaggerInfo.Version = "1.0"
	swagger_docs.SwaggerInfo.Host = this.httpServerConfig.SwaggerAddress
	swagger_docs.SwaggerInfo.BasePath = ""
	swagger_docs.SwaggerInfo.Title = "http server"
	swagger_docs.SwaggerInfo.Description = ""
}

func (this *Main) setHandler() {
	this.server.AddPathPrefixHandler(this.httpServerConfig.SwaggerUri, httpSwagger.WrapHandler)

	this.server.AddHandler("/v1/test/{id:[a-z,A-Z][a-z,A-Z,0-9,--,_,.]+}", net_http.MethodGet, handler.Get)
	this.server.AddHandler("/v1/test", net_http.MethodPost, handler.Post)
	this.server.AddHandler("/v1/test/{id:[a-z,A-Z][a-z,A-Z,0-9,--,_,.]+}", net_http.MethodDelete, handler.Delete)
}

func (this *Main) startServer() error {
	listenAndServeFailureFunc := func(err error) { log.Server.Error(err.Error()) }
	return this.server.Start(this.httpServerConfig.ServerAddress, listenAndServeFailureFunc)
}

func (this *Main) stopServer() error {
	return this.server.Stop(this.httpServerConfig.ShutdownTimeout)
}

func (this *Main) Run() error {
	defer log.Server.Flush()

	if err := this.initialize(); err != nil {
		return err
	}

	log.Server.Info("process start")
	defer log.Server.Info("process end")

	if err := this.startServer(); err != nil {
		return err
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	log.Server.Info("signal", "kind", <-signals)

	return this.stopServer()
}

func main() {
	if err := (&Main{}).Run(); err != nil {
		log.Server.Error(err.Error())
		log.Server.Flush()
	}
}
