package smgeo

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"sm/smbroker"
	"sm/smlog"
	"strconv"
	"time"
)

//ShutdownChannel of type struct{}
type ShutdownChannel chan struct{}

var (
	//ShutdownChann to close all the channels of this service
	ShutdownChann = make(ShutdownChannel)
	osSignal      = make(chan os.Signal, 1)
	//to send health status to broker every 2 second
	healthTicker = time.NewTicker(2 * time.Second)
)

//if we use init() fn, it will be called during tests also.
func InitGeoService() {
	rootCtx := context.Background()

	//Get broker instance to register the broker
	broker := smbroker.GetBrokerInstance()

	//get the geo service instance
	geoSrv := GetGeoServiceInstance(broker)

	logger := smlog.Init(rootCtx, geoSrv.Name)
	ctx := smlog.ContextWithValue(rootCtx, logger)

	//register the service to the broker
	reqChan, err := geoSrv.Broker.Register(ctx, GetGeoSrcName())
	if err != nil {
		logger.With(zap.Error(err)).Error("Error in registering to the broker.")
		panic("Error in starting the service") //TODO: check panic or fatal
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
	go ProcessRequests(ctx, *geoSrv, reqChan)
	//send the health status to broker
	go SendHealthStatus(ctx, *geoSrv, reqChan)

	logger.Sugar().Infof("The service %v with coorindates %+v is initialized successfully", geoSrv.Name, geoSrv.GeoCoordinates)

}

//SendHealthStatus to ping to broker every 4 second
func SendHealthStatus(aInCtx context.Context, geoSrv GeoServiceImpl, chann chan smbroker.Message) {
	for {
		select {
		//send messages to broker every 2 second and target name is empty for the broker
		case <-healthTicker.C:
			geoSrv.Broker.Send(aInCtx, "", smbroker.Message{})
		//if broker closes the channel and make it to null, we will stop this goroutine
		case <-ShutdownChann:
			return
		}
	}
}

//ProcessRequest to process the request from request channel
func ProcessRequests(aInCtx context.Context, geoSrv GeoServiceImpl, reqChan chan smbroker.Message) {
	logger := smlog.MustFromContext(aInCtx)
	for {
		select {
		case msg, chOk := <-reqChan:
			if !chOk {
				logger.Sugar().Info("Request channel closed")
				//close all the service owned channels
				close(ShutdownChann) //
				ShutdownChann = nil
				return
			}
			//Get the implementation fn to process the request
			checkGetOverride := reflect.New(msg.RestStim.MoType.Elem())
			mo := checkGetOverride.Interface()
			v, lOk := mo.(GeoService)
			if lOk {
				switch msg.RestStim.Verb {
				case http.MethodGet:
					//parse the URL to get the input params
					parsedURL, err := url.Parse(msg.RestStim.RestUrl)
					if err != nil {
						logger.With(zap.Error(err)).Error("Error in input data")
						msg.RestStim.RespStatus = http.StatusBadRequest
					} else {
						x := parsedURL.Query().Get("Lat")
						y := parsedURL.Query().Get("Long")
						xVal, _ := strconv.ParseFloat(x, 64)
						yVal, _ := strconv.ParseFloat(y, 64)
						//calculate the distance
						output := v.GetDistance(aInCtx, xVal, yVal)
						//return the response
						msg.RestStim.RespBody = strconv.FormatFloat(output, 'f', 6, 64)
						msg.RestStim.RespStatus = http.StatusOK
					}
				case http.MethodPut:
					v.UpdateCoordinates(aInCtx)
					msg.RestStim.RespStatus = http.StatusOK
				}
			} else {
				msg.RestStim.RespStatus = http.StatusBadRequest
			}
			//update the src and target service name
			msg.TargetSrvName = msg.SrcSrvName
			msg.SrcSrvName = geoSrv.Name
			msg.RestStim.IsResponse = true
			//send the response to the broker to send
			err := geoSrv.Broker.Response(aInCtx, msg.SrcSrvName, msg)
			if err != nil {
				logger.With(zap.Error(err)).Error("Error in sending the response.")
			}

		}

	}

}
