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
	broker     *mocks.MockBrokerI
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
	ut.broker = mocks.NewMockBrokerI(ut.mockCtrl)
	ut.broker.EXPECT().Broadcast(gomock.Any(), gomock.Any()).Return(nil)
	ut.cliService.Broker = ut.broker
	ut.cliService.UpdatePosition(ut.ctx)
}

func (ut *CLISuite) TearDownSuite(c *check.C) {
	ut.mockCtrl.Finish()
}
