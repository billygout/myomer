package zmanim

import "math"
import "time"

const (
	candleLightingOffset float64 = 18
	zenith16Point1 float64 = GeometricZenith + 16.1
	zenith8Point5  float64 = GeometricZenith + 8.5
)

/* from AstronomicalCalendar.cs */

func adjustSunsetDate(sunset *time.Time, sunrise *time.Time) *time.Time {
        if sunset == nil || sunrise == nil || sunrise.Before(*sunset) {
                return sunset
	}
	
	temp := sunset.AddDate(0, 0, 1)
	return &temp
}

func (z *Zmanim) GetOffsetTime(t float64) float64 {
	_, offset := z.Time.Zone()
	offset /= 3600

	return t + float64(offset) 
}

func (z *Zmanim) GetDateFromTime(t float64) *time.Time {
	if math.IsNaN(t) {
		return nil
	}

	t = z.GetOffsetTime(t)

	t = math.Mod((t + 240), 24)
	
	hours := int(t)
	t -= float64(hours)
	t *= 60
	minutes := int(t)
	t -= float64(minutes)
	t *= 60
	seconds := int(t)
	t -= float64(seconds)
	
	date := time.Date(z.Time.Year(), z.Time.Month(), z.Time.Day(),
		hours, minutes, seconds, int(t * 1E9), z.Time.Location())

	return &date
}

func (z *Zmanim) getSunriseOffsetByDegrees(offsetZenith float64) *time.Time {
	alos := z.GetUtcSunrise(offsetZenith, true)
	
	return z.GetDateFromTime(alos)
}

func (z *Zmanim) getSunsetOffsetByDegrees(offsetZenith float64) *time.Time {
	sunset := z.GetUtcSunset(offsetZenith, true)
	if math.IsNaN(sunset) {
		return nil
	}
	
	return adjustSunsetDate(z.GetDateFromTime(sunset), 
		z.getSunriseOffsetByDegrees(offsetZenith))
}

/* from ZmanimCalendar.cs */

func (z *Zmanim) GetTzais() *time.Time {
	return z.getSunsetOffsetByDegrees(zenith8Point5)
}