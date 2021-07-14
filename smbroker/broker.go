package smbroker

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type BrokerI interface {
	//Register to register the services and which returns channels to deliver requests
	Register(ctx context.Context,srvName string) (chan Message, error)
	//Broadcast to broadcast messages to all the services
	Broadcast (ctx context.Context,msg Message)error
	//Send to send the message to the target services
	Send (ctx context.Context,targetSrvName string, msg Message) (chan Message, error)
	//Response to respond the request
	Response(ctx context.Context,targetSrvName string, msg Message)error
	//GetServices to get all the services registered in this broker
	GetServices(ctx context.Context) ([]string, error)
}

var (
	//brokerInstance
	brokerInstance *Broker
	//requestChan in which request is delivered
	requestChan map[string]chan Message
	//healthChan in which health ping is received
	healthChan map[string]chan Message
	//healthLock is for healthChan
	healthLock sync.RWMutex
	//reqLock is for healthChan
	reqLock sync.RWMutex
)


type Broker struct {
	//name of the broker
	name string
}

func (b *Broker) Register(ctx context.Context,srvName string) (chan Message, error){
	//request channel to process the requests
	reqCh:=make(chan Message)
	reqLock.Lock()
	requestChan[srvName]=reqCh
	reqLock.Unlock()
	//health channel to check services health
	healthCh:=make(chan Message)
	healthLock.Lock()
	healthChan[srvName]=healthCh
	healthLock.Unlock()
	//go routine to check the service health
	go checkSrvHealth(srvName,healthCh)
	//return the request channel
	return reqCh,nil
}

func (b *Broker) Broadcast(aInCtx context.Context,msg Message) error {
	//get all the subscribed services
    srvNames , err:=b.GetServices(aInCtx)
    if err != nil{
    	return err
	}
	//broadcast messages
	fmt.Println(srvNames)
	// TODO:

	return nil
}

func (b *Broker) Send(ctx context.Context,targetSrvName string, msg Message) (chan Message, error){
	// the target service name is empty for broker message
	if targetSrvName == ""{
		healthLock.Lock()
		c:=healthChan[msg.SrcSrvName]
		healthLock.Unlock()
		//send the msg to health channel of that service which is maintained by broker
		c<- msg
	}else{
		//response channel
		resChan:=make(chan Message)
		//generate the correlation id
		corrId:=string(rand.Int63())
		msg.RestStim.CorrelationId=corrId
		//for sync message
		if msg.Sync {
			gClientSyncMapLock.Lock()
			syncChan[corrId]= resChan
			gClientSyncMapLock.Unlock()
		}else{
		//for async message
			gClientAsncMapLock.Lock()
			aSyncChan[corrId]= resChan
			gClientAsncMapLock.Unlock()
		}
		return resChan,nil
	}
	return nil,nil
}

func (b *Broker) Response(ctx context.Context,targetSrvName string, msg Message) error {
	//Pass the msg to the corresponding response channel for which the request is listening
	if msg.RestStim.CorrelationId != "" && msg.Sync{
		syncChan[msg.RestStim.CorrelationId] <- msg
	}
	return nil
}
func (b *Broker) GetServices(ctx context.Context) ([]string , error){
	var serviceNames []string
	//healthChan map contains all the services
	for k,_:= range healthChan{
		serviceNames=append(serviceNames,k)
	}
	return serviceNames,nil
}

//GetBrokerInstance to return the broker instance and create if not present
func GetBrokerInstance() *Broker{
	if brokerInstance != nil{
		return brokerInstance
	}
	brokerInstance=&Broker{name:"Mq"}
	return brokerInstance
}

//checkSrvHealth checks the health of all the services
func checkSrvHealth(srvName string,healthCh chan Message) {
	for {
		select {
		case m:=<- healthCh:
			fmt.Print(m)
		case <-time.After(4*time.Second):
			reqLock.Lock()
			srvChan:=requestChan[srvName]
			close(srvChan)
			reqLock.Unlock()
			healthCh=nil
			//unregister the app
			unregister(srvName)
		}
		if healthCh == nil{
			break
		}
	}
}

//unregister clears the health and request channel of the registered services
func unregister(name string) {
	healthLock.Lock()
	reqLock.Lock()
	delete(healthChan,name)
	delete(requestChan,name)
	reqLock.Unlock()
	healthLock.Unlock()
}