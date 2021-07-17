package smbroker

import (
	"context"
	"go.uber.org/zap"
	"sm/smlog"
	"sm/smrand"
	"sync"
	"time"
)

const (
	MaxQueueSize = 100
)

var (
	//brokerInstance
	brokerInstance *Broker
	//requestChan in which request is delivered
	requestChan = make(map[string]chan Message)
	//healthChan in which health ping is received
	healthChan = make(map[string]chan Message)
	//healthLock is for healthChan
	healthLock sync.RWMutex
	//reqLock is for healthChan
	reqLock sync.RWMutex
	//services registered in this broker
	services []Service
)

//BrokerI defines method for the broker related communications
type BrokerI interface {
	//Register to register the services and which returns channels to deliver requests
	Register(ctx context.Context, srvName string) (chan Message, error)
	//Broadcast to broadcast messages to all the services
	Broadcast(ctx context.Context, msg Message) error
	//Send to send the message to the target services
	Send(ctx context.Context, targetSrvName string, msg Message) (chan Message, error)
	//Response to respond the request
	Response(ctx context.Context, targetSrvName string, msg Message) error
	//GetServices to get all the services registered in this broker
	GetServices(ctx context.Context) ([]string, error)
}

//Service contains service information
type Service struct {
	Name           string
	RegisteredDate time.Time
	Type           ServiceType
}

//Broker contains broker instance details
type Broker struct {
	//name of the broker
	name string
}

func (b *Broker) Register(ctx context.Context, srvName string) (chan Message, error) {
	logger := smlog.MustFromContext(ctx)

	//request channel to process the requests
	reqCh := make(chan Message)
	reqLock.Lock()
	requestChan[srvName] = reqCh
	reqLock.Unlock()

	//health channel to check services health
	//if it dont receive status for >4 sec, it will be off-boarded
	healthCh := make(chan Message)
	healthLock.Lock()
	healthChan[srvName] = healthCh
	healthLock.Unlock()

	logger.Sugar().Debugf("Service with name %v is registered", srvName)

	//go routine to check the service health
	go checkSrvHealth(ctx, srvName, healthCh)

	//add services to registered list
	srv := Service{
		Name:           srvName,
		RegisteredDate: time.Now(),
	}
	services = append(services, srv)

	//return the request channel
	return reqCh, nil
}

func (b *Broker) Broadcast(aInCtx context.Context, msg Message) error {
	logger := smlog.MustFromContext(aInCtx)

	//get all the subscribed services
	srvNames := GetServiceNamesByType(services, msg.SrvType)
	logger.Sugar().Debug("The service names are %v", srvNames)
	for _, v := range srvNames {
		msg.TargetSrvName = v
		_, err := b.Send(aInCtx, v, msg)
		if err != nil {
			logger.With(zap.Error(err)).Error("Error in sending tghe data to service", zap.String("ServiceName", v))
		}
	}
	return nil
}

func (b *Broker) Send(ctx context.Context, targetSrvName string, msg Message) (chan Message, error) {
	logger := smlog.MustFromContext(ctx)
	// the target service name is empty for broker message
	if targetSrvName == "" {

		healthLock.Lock()
		c := healthChan[msg.SrcSrvName]
		healthLock.Unlock()

		//send the msg to health channel of that service which is maintained by broker
		c <- msg
	} else {

		//response channel- can queue upto MaxQueueSize
		//if we dont give any, it gets blocked unless any one read a request
		resChan := make(chan Message, MaxQueueSize)

		if msg.RestStim.CorrelationId == "" {
			//generate the correlation id
			corrId := smrand.RandomString(8)
			msg.RestStim.CorrelationId = corrId
		}

		//for sync message
		if msg.Sync {
			gClientSyncMapLock.Lock()
			syncChan[msg.RestStim.CorrelationId] = resChan
			gClientSyncMapLock.Unlock()
		} else {
			//for async message
			gClientAsncMapLock.Lock()
			aSyncChan[msg.RestStim.CorrelationId] = resChan
			gClientAsncMapLock.Unlock()
		}
		logger.Sugar().Debugf("The message is sent to %v from %v", targetSrvName, msg.SrcSrvName)
		return resChan, nil
	}
	return nil, nil
}

func (b *Broker) Response(ctx context.Context, targetSrvName string, msg Message) error {
	logger := smlog.MustFromContext(ctx)
	//Pass the msg to the corresponding response channel for which the request is listening
	if msg.RestStim.IsResponse && msg.RestStim.CorrelationId != "" {
		logger.Sugar().Debugf("Response received from %v is sent to %v", msg.SrcSrvName, targetSrvName)
		if msg.Sync {
			if ch, ok := syncChan[msg.RestStim.CorrelationId]; ok {
				ch <- msg
			}
		} else {
			if ch, ok := aSyncChan[msg.RestStim.CorrelationId]; ok {
				ch <- msg
			}
		}
	}

	return nil
}
func (b *Broker) GetServices(ctx context.Context) ([]string, error) {
	logger := smlog.MustFromContext(ctx)

	var serviceNames []string
	//healthChan map contains all the services
	for k, _ := range healthChan {
		serviceNames = append(serviceNames, k)
	}
	logger.Sugar().Debugf("The registered service names are %+v", serviceNames)
	return serviceNames, nil
}

//GetBrokerInstance to return the broker instance and create if not present
func GetBrokerInstance() *Broker {
	if brokerInstance != nil {
		return brokerInstance
	}

	brokerInstance = &Broker{name: "Mq"}
	return brokerInstance
}

//checkSrvHealth checks the health of all the services
func checkSrvHealth(ctx context.Context, srvName string, healthCh chan Message) {
	logger := smlog.MustFromContext(ctx)
	for {
		select {
		case <-healthCh:
			logger.Sugar().Debugf("Service %v is healthy", srvName)
		case <-time.After(4 * time.Second):
			reqLock.Lock()
			srvChan := requestChan[srvName]
			close(srvChan)
			reqLock.Unlock()
			//unregister the app
			unregisterService(srvName)
			return
		}
	}
}

//unregister clears the health and request channel of the registered services
func unregisterService(name string) {
	healthLock.Lock()
	defer healthLock.Unlock()
	reqLock.Lock()
	defer reqLock.Unlock()
	delete(requestChan, name)
	delete(healthChan, name)
	removeRegService(name)

}

//removeRegService will remove the name from the registered list
func removeRegService(name string) {
	var index int
	for i, v := range services {
		if v.Name == name {
			index = i
		}
	}
	l := len(services)
	// Remove the element at index.
	services[index] = services[l-1]
	services = services[:l-1]
}
