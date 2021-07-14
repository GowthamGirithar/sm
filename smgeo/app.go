package smgeo

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"sm/smbroker"
	"time"
)

func init(){
	ctx:=context.Background()
    //Get broker instance to register the broker
	broker:=smbroker.GetBrokerInstance()
	//initialize the service
	geoSrv:=GeoService{
		Broker: broker,
		Name: GetGeoSrcName(),
	}

	//generate coorfinates
	geoCoordinates= Coordinates{
		X: rand.Int63(),Y: rand.Int63(),
	}

	//register the service to the broker
	chann, err:=geoSrv.Broker.Register(ctx,GetGeoSrcName())
	if err != nil{
		panic("Error in starting the service")
	}


	go ProcessRequest(ctx,geoSrv,chann)
	go SendHealthStatus(ctx,geoSrv, chann)

}

//SendHealthStatus to ping to broker every 4 second
func SendHealthStatus(ctx context.Context ,geoSrv GeoService, chann chan smbroker.Message) {
	for {
		select {
		case <- time.After(4*time.Second):
			geoSrv.Broker.Send(ctx,"", smbroker.Message{})
		}
		if chann == nil{
			break
		}
	}
}

//ProcessRequest to process the request from request channel
func ProcessRequest(aInCtx context.Context,geoSrv GeoService,chann chan smbroker.Message ) {
	for {
		select {
		case msg,ok := <- chann :
			if ok{
				checkGetOverride := reflect.New(msg.RestStim.MoType.Elem())
				mo := checkGetOverride.Interface()
				v, ok := mo.(GeoSrcI)
				if ok{
					//TODO: get the input from msg
					output:=v.GetDistance(10,10)
					//todo: form the o/p msg
					fmt.Print(output)
				}
				//TODO: update source and target name
				err:=geoSrv.Broker.Response(aInCtx,msg.SrcSrvName,msg)
				if err != nil{
					//TODO: LOG
				}
			}else{
				chann=nil
			}
		}
		if chann== nil{
			break
		}
	}

}




