package zipcode

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

type ZipCodeData struct {
	City string
	State string
	Latitude float64
	Longitude float64
	Timezone int
	DST int
}

func (z *ZipCodeData) String() string {
	return fmt.Sprintf("%+v", *z)
}

func GetMap(r io.Reader) (map[string]ZipCodeData, error) {
	cr := csv.NewReader(r)
	
	// read all records
	records, err := cr.ReadAll()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	
	// discard header record
	records = records[1:]

	zipmap := make(map[string]ZipCodeData)
	
	for _, record := range records {
		
		zip := record[0]
		city := record[1]
		state:= record[2]
		
		latitude, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return nil, err
		}

		longitude, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			return nil, err
		}

		timezone, err := strconv.ParseInt(record[5], 10, 0)
		if err != nil {
			return nil, err
		}

		dst, err := strconv.ParseInt(record[6], 10, 0)
		if err != nil {
			return nil, err
		}
		
		zipmap[zip] = ZipCodeData{
			city, state, latitude, longitude, int(timezone), int(dst),
		}
	}

	return zipmap, nil
}

