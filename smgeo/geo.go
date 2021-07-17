package smgeo

import (
	"context"
	"math"
	"math/rand"
	"sm/smbroker"
	"sm/smlog"
	"sm/smrand"
)

var geoSrvInstance *GeoService

//GeoService defines the service information
type GeoService struct {
	//Name of the service
	Name string
	//broker in which is registered
	Broker *smbroker.Broker
	//Coordinates of this service
	GeoCoordinates *Coordinates
}

//GetGeoServiceInstance return the instance
func GetGeoServiceInstance(broker *smbroker.Broker) *GeoService {
	if geoSrvInstance != nil {
		return geoSrvInstance
	}
	//initialize the service
	geoSrvInstance = &GeoService{
		Broker:         broker,
		Name:           GetGeoSrcName(),
		GeoCoordinates: GetNewCoordinates(),
	}
	return geoSrvInstance
}

//GetNewCoordinates get new random generated coordinates
func GetNewCoordinates() *Coordinates {
	return &Coordinates{
		lat: rand.Float64(), long: rand.Float64(),
	}
}

//Coordinates of the location
type Coordinates struct {
	lat  float64
	long float64
}

//Service Name which will be in geo-xxx format
func GetGeoSrcName() string {
	return "geo-" + smrand.RandomString(3)
}

//GeoServiceI has methods for the geo service functionalities
type GeoServiceI interface {
	//GetDistance to calculate the Euclidian distance
	GetDistance(aInCtx context.Context, x, y float64) float64
	//Update the coordinates
	UpdateCoordinates(aInCtx context.Context)
}

func (c *GeoService) GetDistance(aInCtx context.Context, lat, log float64) float64 {
	const PI float64 = 3.141592653589793

	radlat1 := PI * c.GeoCoordinates.lat / 180
	radlat2 := PI * lat / 180

	theta := c.GeoCoordinates.long - log
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

func (c *GeoService) UpdateCoordinates(aInCtx context.Context) {
	logger := smlog.MustFromContext(aInCtx)
	geoSrvInstance.GeoCoordinates = GetNewCoordinates()
	logger.Sugar().Infof("Service coordinates changes to %+v", geoSrvInstance.GeoCoordinates)
}
