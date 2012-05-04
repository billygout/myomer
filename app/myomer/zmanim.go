package myomer

import (
	"appengine"

	"zmanim"
	"fmt"
	"time"
	"timezone"
)

func getZmanimString(c appengine.Context, zipcode string) string { 
	if len(zipcode) == 0 {
		return ""
	}
	if zipMap == nil {
		err := loadZipMap(c)
		if err != nil {
			return ""
		}
	}
	
	zd, ok := zipMap[zipcode]
	if !ok {
		return ""
	} 

	utc := time.Now()
	local, _ := timezone.TimeInZone(&utc, zd.Timezone, zd.DST)

	z := zmanim.Zmanim{
	Time: local,
	Geolocation: zmanim.Geolocation{
		Longitude: zd.Longitude, 
		Latitude: zd.Latitude,
		},
	}	
	
	zenith := zmanim.GeometricZenith
	
	rise := z.GetUtcSunrise(zenith, false)
	sunrise := z.GetDateFromTime(rise)
	
	set := z.GetUtcSunset(zenith, false)
	sunset := z.GetDateFromTime(set)
	
	tzais := z.GetTzais()
	
	dateFormat := "3:04:05PM"

	s := ""
	
	s += fmt.Sprintf("Zmanim for %v, %v (%v) on %v\n", 
		zd.City, zd.State, zipcode, z.Time.Format("1/2/2006"))

	s += fmt.Sprintf("Sunrise: %v\n", sunrise.Format(dateFormat))
	s += fmt.Sprintf("Sunset:  %v\n", sunset.Format(dateFormat))
	s += fmt.Sprintf("Tzais:   %v\n", tzais.Format(dateFormat))
	
	return s
}

	