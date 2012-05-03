package timezone

import (
	"time"
)

func secondSundayInMarch(t *time.Time) int {
	
	// first day in march
	day := time.Date(t.Year(), time.March, 1, 0, 0, 0, 0, t.Location())

	// duration of one day
	oneday := time.Duration(24 * time.Hour)

	sundayCount := 0
	for {
		if day.Weekday() == time.Sunday {
			sundayCount++
			
			if sundayCount == 2 {
				break
			}
		}

		day = day.Add(oneday)
	}

	return day.Day()
}

func firstSundayInNovember(t *time.Time) int {
	
	// first day in november
	day := time.Date(t.Year(), time.November, 1, 0, 0, 0, 0, t.Location())

	// duration of one day
	oneday := time.Duration(24 * time.Hour)

	sundayCount := 0
	for {
		if day.Weekday() == time.Sunday {
			sundayCount++
			
			if sundayCount == 1 {
				break
			}
		}

		day = day.Add(oneday)
	}

	return day.Day()
}

func dstInEffect(t *time.Time) bool {

	dstStartTime := time.Date(t.Year(), time.March, secondSundayInMarch(t), 
		2, 0, 0, 0, t.Location())

	dstEndTime := time.Date(t.Year(), time.November, firstSundayInNovember(t), 
		2, 0, 0, 0, t.Location())

	return !t.Before(dstStartTime) && t.Before(dstEndTime)
}

func TimeInZone(t *time.Time, tz, dst int) (time.Time, bool){
	offset := int(3600 * tz)
	location := time.FixedZone("", offset)
	
	duration := time.Duration(time.Hour * time.Duration(tz))
	
	adj := t.Add(duration)
	
	local := time.Date(
		adj.Year(), 
		adj.Month(), 
		adj.Day(),
		adj.Hour(),
		adj.Minute(),
		adj.Second(),
		adj.Nanosecond(),
		location)

	// handle shifting for DST
	if dst != 0 {
		effect := dstInEffect(&local)
		if effect {
			local = local.Add(time.Hour)
			
			location = time.FixedZone("", offset+3600)
			local = time.Date(
				local.Year(), 
				local.Month(), 
				local.Day(),
				local.Hour(),
				local.Minute(),
				local.Second(),
				local.Nanosecond(),
				location)
		}
		
		return local, effect
	}

	return local, false
}

