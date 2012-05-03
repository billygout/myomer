package zmanim

import "math"

/* from MathExntensions.cs */

func toDegree(angleRadians float64) float64 {
	return (180.0 * angleRadians / math.Pi)
}

func toRadians(angleDegree float64) float64 {
	return (math.Pi * angleDegree / 180.0)
}

/* from AstronomicalCalendar.cs */

const GeometricZenith float64 = 90

/* from AstronomicalCalculator.cs */

const solarRadius float64 = (16./60.)
const refraction float64 = (34./60.)

func getElevationAdjustment(elevation float64) float64 {
	earthRadius := float64(6356.9)
	
	elevationAdjustment := toDegree(
		math.Acos(earthRadius/(earthRadius + (elevation/1000))))
	
	return elevationAdjustment
}

func adjustZenith(zenith, elevation float64) float64 {
	if zenith == GeometricZenith {
		zenith += solarRadius + 
			refraction + 
			getElevationAdjustment(elevation)	
	}
		
	return zenith
}

/* from SunTimesCalculator.cs */

const (
	//zenith float64 = 90 + 50.0/60.0
	degPerHour float64 = 360.0/24.0
)

func sinDeg(deg float64) float64 {
        return math.Sin(deg*2.0*math.Pi/360.0)
}


func acosDeg(x float64) float64 {
        return math.Acos(x)*360.0/(2*math.Pi)
}

func asinDeg(x float64) float64 {
        return math.Asin(x)*360.0/(2*math.Pi);
}

func tanDeg(deg float64) float64 {
        return math.Tan(deg*2.0*math.Pi/360.0);
}

func cosDeg(deg float64) float64 {
	return math.Cos(deg*2.0*math.Pi/360.0);
}

func getApproxTimeDays(dayOfYear int, hoursFromMeridian float64, is_sunset bool) float64 {
	if !is_sunset {
		return float64(dayOfYear) + ((6.0 - hoursFromMeridian)/24)
	} else {
		return float64(dayOfYear) + ((18.0 - hoursFromMeridian)/24)
	}
	
	return 0
}

func getSunTrueLongitude(sunMeanAnomaly float64) float64 {
	l := sunMeanAnomaly + 
		(1.916*sinDeg(sunMeanAnomaly)) + 
		(0.020*sinDeg(2*sunMeanAnomaly)) + 
		282.634;

	if l >= 360.0 {
		l = l - 360.0
	}
	if l < 0 {
		l = l + 360.0
	}
	return l
}

func getSunRightAscensionHours(sunTrueLongitude float64) float64 {
        a := 0.91764*tanDeg(sunTrueLongitude)
        ra := 360.0/(2.0*math.Pi)*math.Atan(a)
	lQuadrant := math.Floor(sunTrueLongitude/90.0)*90.0
        raQuadrant := math.Floor(ra/90.0)*90.0
        ra = ra + (lQuadrant - raQuadrant)
	return ra/degPerHour
}

func getLocalMeanTime(localHour, sunRightAscensionHours, approxTimeDays float64) float64 {
        return localHour + sunRightAscensionHours - (0.06571*approxTimeDays) - 6.622
}

/* (*Zmanim) methods */

func (z *Zmanim) GetUtcSunrise(zenith float64, adjustForElevation bool) float64 {
	time := math.NaN()

	if adjustForElevation {
		zenith = adjustZenith(zenith, z.Elevation)
	} else {
		zenith = adjustZenith(zenith, 0)
	}

	time = z.getTimeUtc(zenith, false)

	return time	
}

func (z *Zmanim) GetUtcSunset(zenith float64, adjustForElevation bool) float64 {
	time := math.NaN()

	if adjustForElevation {
		zenith = adjustZenith(zenith, z.Elevation)
	} else {
		zenith = adjustZenith(zenith, 0)
	}

	time = z.getTimeUtc(zenith, true)

	return time	
}

func (z *Zmanim) getDayOfYear() int {
	y, month, d := z.Time.Date()
	m := int(month)

	n1 := 275*m/9
	n2 := (m + 9)/12
	n3 := (1 + ((y - 4*(y/4) + 2)/3))
	n := n1 - (n2*n3) + d - 30

	return n
}

func (z *Zmanim) getHoursFromMeridian() float64 {
	return z.Geolocation.Longitude / degPerHour
}

func (z *Zmanim) getMeanAnomaly(dayOfYear int, is_sunset bool) float64 {
	return (0.9856*getApproxTimeDays(dayOfYear, z.getHoursFromMeridian(), is_sunset)) - 3.289
}


func (z *Zmanim) getCosLocalHourAngle(sunTrueLongitude float64, zenith float64) float64 { 
	sinDec := 0.39782*sinDeg(sunTrueLongitude)
        cosDec := cosDeg(asinDeg(sinDec))

        cosH := (cosDeg(zenith) - (sinDec*sinDeg(z.Geolocation.Latitude))) /
		(cosDec*cosDeg(z.Geolocation.Latitude))
	
        return cosH;
}

func (z *Zmanim) getTimeUtc(zenith float64, is_sunset bool) float64 {
	dayOfYear := z.getDayOfYear()
        sunMeanAnomaly := z.getMeanAnomaly(dayOfYear, is_sunset)
        sunTrueLong :=  getSunTrueLongitude(sunMeanAnomaly)
        sunRightAscensionHours := getSunRightAscensionHours(sunTrueLong)
        cosLocalHourAngle := z.getCosLocalHourAngle(sunTrueLong, zenith)
	
        localHourAngle := float64(0)
        
	if !is_sunset {
                if cosLocalHourAngle > 1 {
	        }
                
		localHourAngle = 360.0 - acosDeg(cosLocalHourAngle)
        } else {
		if cosLocalHourAngle < -1 {
                }
                
		localHourAngle = acosDeg(cosLocalHourAngle)
        }
        
	localHour := localHourAngle/degPerHour

	hoursFromMeridian := z.getHoursFromMeridian()
	
	approxTimeDays := getApproxTimeDays(dayOfYear, hoursFromMeridian, is_sunset)

        localMeanTime := getLocalMeanTime(localHour, sunRightAscensionHours, approxTimeDays)
	
        pocessedTime := localMeanTime - hoursFromMeridian
        
	for pocessedTime < 0.0 {
                pocessedTime += 24.0;
        }
        
	for pocessedTime >= 24.0 {
                pocessedTime -= 24.0;
        }

        return pocessedTime;
} 

