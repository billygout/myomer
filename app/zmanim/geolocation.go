package zmanim

import "fmt"

type Geolocation struct {
	Latitude  float64
	Longitude float64
	Elevation float64
}

func (g *Geolocation) String() string {
	return fmt.Sprintf("%v/%v/%v", g.Latitude, g.Longitude, g.Elevation)
}
