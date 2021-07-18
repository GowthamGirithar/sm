package smgeo_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"gopkg.in/check.v1"
	"sm/smbroker/mocks"
	"sm/smgeo"
	"sm/smlog"
	"testing"
)

type GeoSuite struct {
	t          *testing.T
	mockCtrl   *gomock.Controller
	ctx        context.Context
	geoService *smgeo.GeoService
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
	ut.geoService = smgeo.GetGeoServiceInstance(ut.broker)
}

func (ut *GeoSuite) TestUpdateCoordinates(c *check.C) {
	ut.geoService.UpdateCoordinates(ut.ctx)
	//TODO: COMPARE OLD VALUES AND NEW VALUES
}

func (ut *GeoSuite) TestGetDistance(c *check.C) {
	ut.broker = mocks.NewMockBrokerI(ut.mockCtrl)
	geo := smgeo.GetGeoServiceInstance(ut.broker)
	geo.GetDistance(ut.ctx, -1.657, 2.789)
	//TODO: CHECK VALUE
}

func (ut *GeoSuite) TearDownSuite(c *check.C) {
	ut.mockCtrl.Finish()
}
