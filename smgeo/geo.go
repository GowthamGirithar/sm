package smgeo

import (
	"sm/smbroker"
)

var geoCoordinates Coordinates

type GeoService struct {
	//Name of the service
	Name string
	//broker in which is registered
	Broker *smbroker.Broker
}

//Coordinates of the location
type Coordinates struct {
	X int64
	Y int64
}

func GetGeoSrcName() string{
	//TODO: generate random string
	return ""
}

type GeoSrcI interface {
	GetDistance(X, Y int64) float64
}


func (c *Coordinates) GetDistance(X, Y int64) float64{
	//TODO; perform calculation
	return 0
}

