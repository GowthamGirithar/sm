package smcli

import (
	"context"
	"go.uber.org/zap"
	"math"
	"math/rand"
	"sm/smbroker"
	"sm/smgeo"
	"sm/smlog"
	"strconv"
	"strings"
	"sync"
	"time"
)

//CLIService define the service
type CLIService struct {
	Name   string
	Broker *smbroker.Broker
}

var (
	//to store the service name and the distance
	distanceOp map[string]float64
	//output count defines how many services have sent response
	opCount int
	//lock for opCount
	opCountLock sync.Mutex
	//lock for distanceOpLock
	distanceOpLock sync.Mutex
)

//CLIServiceI defines methods for user operations
type CLIServiceI interface {
	//Find service which is min to the CLIService
	CalculateMinDistances(aInCtx context.Context, coordinates smgeo.Coordinates) error
	//Send a request to update the geo location
	UpdatePosition(aInCtx context.Context) error
}

func (c *CLIService) CalculateMinDistances(aInCtx context.Context, coordinates smgeo.Coordinates) error {
	logger := smlog.MustFromContext(aInCtx)
	//get all the services
	services, err := c.Broker.GetServices(aInCtx)
	if err != nil {
		logger.With(zap.Error(err)).Error("Error in receiving the requests")
		return err
	}
	//request channel to send the coordinates to geo services
	reqCh := make(chan smbroker.Message, 3)
	//send messages
	go sendMessage(aInCtx, c, reqCh)
	//filter the services with templace name is in geo-xxx format
	//alternate way is to fetch based on type
	//make a call to those services
	var servicesCount int
	for _, v := range services {
		if strings.HasPrefix(v, "geo-") {
			msg := smbroker.Message{}
			msg.TargetSrvName = v
			stim := smbroker.RestStim{}
			//generate the correlation id
			corrId := string(rand.Int63())
			stim.CorrelationId = corrId
			msg.RestStim = smbroker.RestStim{}
			//TODO: Other parameteres
			reqCh <- msg
			servicesCount++
		}
	}
	//close the request channel
	close(reqCh)
	go PrintMinDistance(aInCtx, servicesCount)
	return nil
}

func PrintMinDistance(aInCtx context.Context, count int) {
	logger := smlog.MustFromContext(aInCtx)
	for {
		if count == opCount {
			//calculation to find min distance from the available ones
			minDist := math.MaxFloat64
			var minDistSrv string
			for serviceName, dist := range distanceOp {
				if dist < minDist {
					minDist = dist
					minDistSrv = serviceName
				}
			}
			logger.Sugar().Infof("The serice %v is the nearest one", minDistSrv)
			return
		}
		//retry after sometime
		time.Sleep(1 * time.Minute)
	}
}

//sendMessage to send the message to broker
func sendMessage(aInCtx context.Context, c *CLIService, reqCh <-chan smbroker.Message) {
	logger := smlog.MustFromContext(aInCtx)
	for {
		select {
		case msg, ok := <-reqCh:
			if !ok {
				logger.Sugar().Debug("Request channel is closed and no more requests are queued.")
				return
			}
			//timeout for this operations
			ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
			ch, err := c.Broker.Send(ctx, msg.TargetSrvName, msg)
			if err != nil {
				logger.With(zap.Error(err)).Error("Error in sending the message.")
			}

			//Process the channel to store the output for calculation
			//storeResults to store the results in the map for the processing
			go func(aInCtx context.Context, serviceName string, ch chan smbroker.Message) {
				//time out for the operation default value is 10 minute
				d := 10 * time.Minute
				if deadline, deadlineSet := ctx.Deadline(); deadlineSet {
					d = time.Until(deadline)
					if d <= time.Duration(0) {
						return
					}
				}
				for {
					select {
					case msg := <-ch:
						//response from geo service is available
						val := msg.RestStim.RespBody
						if val != "" {
							dis, err := strconv.ParseFloat(val, 64)
							if err != nil {
								logger.With(zap.Error(err)).Error("Invalid Response")
							} else {
								distanceOpLock.Lock()
								distanceOp[msg.SrcSrvName] = dis
								distanceOpLock.Unlock()
							}
						}
						opCountLock.Lock()
						opCount++
						opCountLock.Unlock()
					case <-time.After(d):
						close(ch)
						opCountLock.Lock()
						opCount++
						opCountLock.Unlock()
						logger.Sugar().Infof("The service %v havent sent the response yet and so we are discarding it after wait time", serviceName)
						return
					}
				}
			}(aInCtx, msg.TargetSrvName, ch)
		}
	}
}

func (c *CLIService) UpdatePosition(aInCtx context.Context) error {
	logger := smlog.MustFromContext(aInCtx)
	msg := smbroker.Message{}
	err := c.Broker.Broadcast(aInCtx, msg)
	if err != nil {
		logger.With(zap.Error(err)).Error("Error in sending the requests.")
		return err
	}
	return nil
}
