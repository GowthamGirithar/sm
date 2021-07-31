package smbroker_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"gopkg.in/check.v1"
	"sm/smbroker"
	"sm/smlog"
	"testing"
)

type BrokerSuite struct {
	t        *testing.T
	mockCtrl *gomock.Controller
	ctx      context.Context
}

func TestBrokerSuite(t *testing.T) {
	check.Suite(&BrokerSuite{
		t: t,
	})
	check.TestingT(t)
}

func (ut *BrokerSuite) SetUpSuite(c *check.C) {
	ut.mockCtrl = gomock.NewController(ut.t)
	ut.ctx = context.Background()
	ut.ctx = smlog.ContextWithValue(ut.ctx, smlog.NewLogger(ut.ctx, "TEST"))
}

func (ut *BrokerSuite) TearDownSuite(c *check.C) {
	ut.mockCtrl.Finish()
}

func (ut *BrokerSuite) TestRegister(c *check.C) {
	//register
	broker := smbroker.BrokerImpl{}
	ch, err := broker.Register(ut.ctx, "TEST")
	c.Assert(err, check.IsNil)
	c.Assert(ch, check.NotNil)

	//send the data
	msg := smbroker.Message{
		RestStim: smbroker.RestStim{
			CorrelationId: "123",
		},
		Sync:       true,
		SrcSrvName: "TEST",
	}
	ch, err = broker.Send(ut.ctx, "SRVB", msg)
	c.Assert(err, check.IsNil)
	c.Assert(ch, check.NotNil)

	//respond from SRVB
	msg.SrcSrvName = "SRVB"
	msg.TargetSrvName = "TEST"
	msg.RestStim.IsResponse = true
	err = broker.Response(ut.ctx, "TEST", msg)
	c.Assert(err, check.IsNil)
	data, ok := <-ch
	if !ok {
		c.Fail()
	}
	c.Assert(data, check.NotNil)

}
