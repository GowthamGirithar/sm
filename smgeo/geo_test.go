package smgeo_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"gopkg.in/check.v1"
	"sm/smbroker"
	"sm/smbroker/mocks"
	"sm/smgeo"
	"sm/smlog"
	"testing"
)

type GeoSuite struct {
	t          *testing.T
	mockCtrl   *gomock.Controller
	ctx        context.Context
	geoService smgeo.GeoServiceI
	broker     *mocks.MockBrokerI
}

func TestGeoSuite(t *testing.T) {
	check.Suite(&GeoSuite{
		t: t,
	})
	check.TestingT(t)
}

func (ut *GeoSuite) SetUpSuite(c *check.C) {
	ut.mockCtrl = gomock.NewController(ut.t)
	ut.ctx = context.Background()
	ut.ctx = smlog.ContextWithValue(ut.ctx, smlog.NewLogger(ut.ctx, "TEST"))
	ut.broker = mocks.NewMockBrokerI(ut.mockCtrl)
}

func (ut *GeoSuite) TestUpdateCoordinates(c *check.C) {
	ut.broker.EXPECT().Response(ut.ctx, gomock.Any(), gomock.Any()).Return(nil).Times(1)
	geo := smgeo.GetGeoServiceInstance(ut.broker)
	geo.UpdateCoordinates(ut.ctx)
}

func (ut *GeoSuite) TearDownSuite(c *check.C) {
	ut.mockCtrl.Finish()
}

func (ut *GeoSuite) TestSend(c *check.C) {
	ch := make(chan smbroker.Message)
	ut.broker.EXPECT().Send(ut.ctx, gomock.Any(), gomock.Any()).Return(ch, nil).Times(1)
	geo := smgeo.GetGeoServiceInstance(ut.broker)
	geo.GetDistance(ut.ctx, -1.657, 2.789)
	//TODO: check distance
}
