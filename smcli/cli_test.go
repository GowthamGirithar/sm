package smcli_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"gopkg.in/check.v1"
	"sm/smbroker/mocks"
	"sm/smcli"
	"sm/smlog"
	"testing"
)

type CLISuite struct {
	t          *testing.T
	mockCtrl   *gomock.Controller
	ctx        context.Context
	cliService smcli.CLIService
}

func TestCLISuite(t *testing.T) {
	check.Suite(&CLISuite{
		t: t,
	})
	check.TestingT(t)
}

func (ut *CLISuite) SetUpSuite(c *check.C) {
	ut.mockCtrl = gomock.NewController(ut.t)
	ut.ctx = context.Background()
	ut.ctx = smlog.ContextWithValue(ut.ctx, smlog.NewLogger(ut.ctx, "TEST"))
}

func (ut *CLISuite) TestUpdatePosition(c *check.C) {
	broker := mocks.NewMockBrokerI(ut.mockCtrl)
	broker.EXPECT().Broadcast(gomock.Any(), gomock.Any()).Return(nil)
	ut.cliService.Broker = broker
	ut.cliService.UpdatePosition(ut.ctx)
}

func (ut *CLISuite) TearDownSuite(c *check.C) {
	ut.mockCtrl.Finish()
}
