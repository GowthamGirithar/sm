package smbroker

import (
	"net/http"
	"reflect"
	"sync"
)

var (
	//syncChan for sync response channel with correlation id
	syncChan map[string]chan Message
	//aSyncChan for sync response channel with correlation id
	aSyncChan map[string]chan Message
	gClientSyncMapLock     sync.RWMutex
	gClientAsncMapLock     sync.RWMutex
)

//RestStim contains data about the REST
type RestStim struct {
	RequestId string
	Verb string
	RestUrl string
	TargetMoName string
	RestHeaders http.Header
	RestBody string
	RespStatus int
	RespHeaders http.Header
	RespBody string
	IsResponse bool
	CorrelationId string
	MoName string
	MoType reflect.Type
}

//Message contains communication endpoints details
type Message struct {
	TargetSrvName string
	SrcSrvName string
	Sync bool
	RestStim RestStim
}

type Executer interface {
	Execute() error
}

func (r RestStim)Execute() error{
	return nil
}
