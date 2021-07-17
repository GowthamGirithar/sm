package smgeo_test

import (
	check ".glide/cache/src/https-github.com-CiscoM31-check"
	"context"
	"github.com/golang/mock/gomock"
	"sm/smgeo"
	"testing"
)

type GeoSuite struct {
	t        *testing.T
	mockCtrl *gomock.Controller
	ctx      context.Context
}

func TestBrokerSuite(t *testing.T) {
	check.Suite(&GeoSuite{
		t: t,
	})
	check.TestingT(t)
}

func (ut *GeoSuite) SetUpSuite(c *check.C) {
	ut.mockCtrl = gomock.NewController(ut.t)
}

func getGeoInstance() smgeo.GeoServiceI {
	geo := &smgeo.GeoService{}
	return geo
}

func (ut *GeoSuite) TestUpdateCoordinates(c *check.C) {
	//mockBroker:=mocks.NewMockBrokerI(ut.mockCtrl)
	//g:=smgeo.GetGeoServiceInstance(mockBroker)
	//g.UpdateCoordinates(ut.ctx)

}

func (ut *GeoSuite) TearDownSuite(c *check.C) {
	ut.mockCtrl.Finish()
}

func (ut *GeoSuite) TestSend(c *check.C) {

}
