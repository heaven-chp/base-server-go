package main

import (
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/heaven-chp/base-server-go/config"
	command_line_argument "github.com/heaven-chp/common-library-go/command-line-argument"
	"github.com/heaven-chp/common-library-go/grpc"
	"github.com/heaven-chp/common-library-go/grpc/sample"
	"github.com/heaven-chp/common-library-go/log"
)

type Main struct {
	server           grpc.Server
	grpcServerConfig config.GrpcServer
}

func (this *Main) Initialize() error {
	err := this.initializeFlag()
	if err != nil {
		return err
	}

	err = this.initializeConfig()
	if err != nil {
		return err
	}

	err = this.initializeLog()
	if err != nil {
		return err
	}

	err = this.initializeServer()
	if err != nil {
		return err
	}

	return nil
}

func (this *Main) Finalize() error {
	defer this.finalizeLog()

	return this.finalizeServer()
}

func (this *Main) initializeFlag() error {
	err := command_line_argument.Set([]command_line_argument.CommandLineArgumentInfo{
		{FlagName: "config_file", Usage: "config/GrpcServer.config", DefaultValue: string("")},
	})
	if err != nil {
		return nil
	}

	if flag.NFlag() != 1 {
		flag.Usage()
		return errors.New("invalid flag")
	}

	return nil
}

func (this *Main) initializeConfig() error {
	return config.Parsing(&this.grpcServerConfig, command_line_argument.Get("config_file").(string))
}

func (this *Main) initializeLog() error {
	level, err := log.ToIntLevel(this.grpcServerConfig.LogLevel)
	if err != nil {
		return err
	}

	return log.Initialize(level, this.grpcServerConfig.LogOutputPath, this.grpcServerConfig.LogFileNamePrefix)
}

func (this *Main) finalizeLog() error {
	return log.Finalize()
}

func (this *Main) initializeServer() error {
	err := this.server.Initialize(this.grpcServerConfig.Address, &sample.Server{})
	if err != nil {
		return err
	}

	go func() {
		err := this.server.Run()
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

func (this *Main) finalizeServer() error {
	return this.server.Finalize()
}

func (this *Main) Run() error {
	err := this.Initialize()
	if err != nil {
		return err
	}
	defer this.Finalize()

	log.Info("process start")
	defer log.Info("process end")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	log.Info("signal : (%s)", <-signals)

	return nil
}

func main() {
	main := Main{}

	err := main.Run()
	if err != nil {
		log.Error(err.Error())
	}
}