package smcli

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sm/smbroker"
	"sm/smlog"
)

//ShutdownChannel of type struct{}
type ShutdownChannel chan struct{}

var (
	//ShutdownChann to close all the channels of this service
	ShutdownChann = make(ShutdownChannel)
	osSignal      = make(chan os.Signal, 1)
)

func init() {
	fmt.Print("Initializing the CLI app")
	rootCtx := context.Background()

	//Get broker instance to register the broker
	broker := smbroker.GetBrokerInstance()

	//initialize the service
	cliSrv := CLIService{
		Broker: broker,
		Name:   "CLI",
	}

	logger := smlog.Init(rootCtx, cliSrv.Name)
	ctx := smlog.ContextWithValue(rootCtx, logger)

	//register the service to the broker
	chann, err := cliSrv.Broker.Register(ctx, cliSrv.Name)
	if err != nil {
		logger.With(zap.Error(err)).Error("Error in registering to the broker.")
		panic("Error in starting the service")
	}

	// trap SIGINT to trigger a shutdown.
	signal.Notify(osSignal, os.Interrupt)
	go func() {
		for {
			select {
			case <-osSignal:
				close(ShutdownChann)
				ShutdownChann = nil
				return
			}
		}
	}()

	//Process the requests
	go ProcessRequest(ctx, cliSrv, chann)
	//send the health status to broker
	go SendHealthStatus(ctx, cliSrv, chann)

}

//SendHealthStatus to ping to broker every 4 second
func SendHealthStatus(aInCtx context.Context, cliSrv CLIService, chann chan smbroker.Message) {
	for {
		// If shutdown already initiated return
		if ShutdownChann == nil {
			return
		}
		select {
		//send messages to broker every 2 second and target name is empty for the broker
		case <-healthTicker.C:
			cliSrv.Broker.Send(aInCtx, "", smbroker.Message{})
		//if broker closes the channel and make it to null, we will stop this goroutine
		case <-ShutdownChann:
			return
		}
	}
}

//ProcessRequest to process the request from request channel
func ProcessRequest(aInCtx context.Context, cliSrv CLIService, chann chan smbroker.Message) {
	logger := smlog.MustFromContext(aInCtx)
	// If shutdown already initiated return
	if ShutdownChann == nil {
		return
	}
	for {
		select {
		case _, ok := <-chann:
			if !ok {
				logger.Sugar().Info("Request channel closed")
				//close all the service owned channels
				close(ShutdownChann) // closing of channel will send message
				ShutdownChann = nil
				return
			}
		}
	}
}
