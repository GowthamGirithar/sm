package smcli

import (
	"context"
	"fmt"
	"sm/smbroker"
	"time"
)

func init(){
     fmt.Print("Initializing the CLI app")

	ctx:=context.Background()
	//Get broker instance to register the broker
	broker:=smbroker.GetBrokerInstance()
	//initialize the service
	cliSrv:=CLIService{
		Broker: broker,
		Name: "CLI",
	}


	//register the service to the broker
	chann, err:=cliSrv.Broker.Register(ctx,cliSrv.Name)
	if err != nil{
		panic("Error in starting the service")
	}


	go ProcessRequest(ctx,cliSrv,chann)
	go SendHealthStatus(ctx,cliSrv, chann)
}

//SendHealthStatus to ping to broker every 4 second
func SendHealthStatus(ctx context.Context ,cliSrv CLIService, chann chan smbroker.Message) {
	for {
		select {
		case <- time.After(4*time.Second):
			cliSrv.Broker.Send(ctx,"", smbroker.Message{})
		}
		if chann == nil{
			break
		}
	}
}

//ProcessRequest to process the request from request channel
func ProcessRequest(aInCtx context.Context,cliSrv CLIService,chann chan smbroker.Message ) {
	for {
		select {
		case msg,ok := <- chann :
			if ok{
				fmt.Println(msg)
			}else{
				chann=nil
			}
		}
		if chann== nil{
			break
		}
	}

}




