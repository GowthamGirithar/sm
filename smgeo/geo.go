package smgeo

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sm/smbroker"
	"sm/smlog"
	"sm/smrand"
)


//GeoService defines the service information
type GeoServiceImpl struct {
	//Name of the service
	Name string
	//broker in which is registered
	//has to be interface type which helps in mocking
	Broker smbroker.Broker
	//Coordinates of this service
	GeoCoordinates *Coordinates
}

//GetGeoServiceInstance return the instance
func GetGeoServiceInstance(broker smbroker.Broker) *GeoServiceImpl {
	//initialize the service
	geoSrvInstance := &GeoServiceImpl{
		Broker:         broker,
		Name:           GetGeoSrcName(),
		GeoCoordinates: GetNewCoordinates(),
	}
	return geoSrvInstance
}

//GetNewCoordinates get new random generated coordinates
func GetNewCoordinates() *Coordinates {
	return &Coordinates{
		Lat: rand.Float64(), Long: rand.Float64(),
	}
}

//Coordinates of the location
type Coordinates struct {
	Lat  float64
	Long float64
}

func (c *GeoServiceImpl) updateCoordinates(){
	c.GeoCoordinates=GetNewCoordinates()
}

//Service Name which will be in geo-xxx format
func GetGeoSrcName() string {
	return fmt.Sprintf("geo-%+v", smrand.RandomString(3))
}

//GeoServiceI has methods for the geo service functionalities
type GeoService interface {
	//GetDistance to calculate the Euclidian distance
	GetDistance(aInCtx context.Context, x, y float64) float64
	//Update the coordinates
	UpdateCoordinates(aInCtx context.Context)
}

func (c *GeoServiceImpl) GetDistance(aInCtx context.Context, lat, log float64) float64 {
	const PI float64 = 3.141592653589793

	radlat1 := PI * c.GeoCoordinates.Lat / 180
	radlat2 := PI * lat / 180

	theta := c.GeoCoordinates.Long - log
	radtheta := PI * theta / 180

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515
	return dist
}

func (c *GeoServiceImpl) UpdateCoordinates(aInCtx context.Context) {
	logger := smlog.MustFromContext(aInCtx)
	c.updateCoordinates()
	logger.Sugar().Debugf("Service coordinates changes to %+v", c.GeoCoordinates)
}
