package myomer

import (
	"zipcode"
	"zmanim"
	"timezone"

	"fmt"	
        "html/template"
        "io"
        "net/http"
	"appengine"
        "appengine/blobstore"
	"appengine/datastore"
	"errors"
	"time"
)

var zipMap map[string]zipcode.ZipCodeData

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
        w.WriteHeader(http.StatusInternalServerError)
        w.Header().Set("Content-Type", "text/plain")
        io.WriteString(w, "Internal Server Error")
        c.Errorf("%v", err)
}

func findFileBlob(c appengine.Context, filename string) (appengine.BlobKey, error) {
	q := datastore.NewQuery("__BlobInfo__")
	pll := []datastore.PropertyList{}
	keys, err := q.GetAll(c, &pll)
	if err != nil {
		return "", err
	}
	
	for i, pl := range pll {
		for _, p := range pl {
			if p.Name == "filename" {
				if p.Value == filename {
					return appengine.BlobKey(keys[i].StringID()), nil
				}
			}
		}
	}
	
	return "", errors.New("blob for \"" + filename + "\" not found")
}


func loadZipMap(c appengine.Context) error {
	b, err := findFileBlob(c, "zipcode.csv")
	if err != nil {
		return err
	} 
	
	reader := blobstore.NewReader(c, b)
	zipMap, err = zipcode.GetMap(reader)
	if err != nil {
		return err
	}
	
	return nil
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

const zmanimHTML = `
<html><body>
<form action="/postzmanim" method="POST" enctype="multipart/form-data">
Zip Code: <input type="text" name="zipcode"><br>
<input type="submit" name="submit" value="Submit">
</form></body></html>
`

func handleZmanim(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	if zipMap == nil {
		err := loadZipMap(c)
		if err != nil {
			serveError(c, w, err)
			return
		}

		c.Infof("Loaded zipMap")
	}

	fmt.Fprintf(w, zmanimHTML)
}

func handlePostZmanim(w http.ResponseWriter, r *http.Request) {
	
	zipcode := r.FormValue("zipcode")
	
	if zipcode == "" {
		fmt.Fprintf(w, "Empty zipcode specified")
		return
	}

	zd, ok := zipMap[zipcode]
	if !ok {
		fmt.Fprintf(w, "zipcode \"%v\" is not in the database\n", zipcode)
		return
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
	
	dateFormat := "1/2/2006 3:04:05PM (MST)"

	fmt.Fprintf(w, "<html><body>")
	
	fmt.Fprintf(w, "Zmanim for %v, %v (%v) on %v <br>", 
		zd.City, zd.State, zipcode, z.Time.Format("1/2/2006"))
	
	//fmt.Fprintf(w, "(Timezone Offset: %v DST? %v)<br>", zd.Timezone, zd.DST == 1)
	
	fmt.Fprintf(w, "<table>")

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, "<td>")
	fmt.Fprintf(w, "Sunrise:")
	fmt.Fprintf(w, "</td>")
	fmt.Fprintf(w, "<td>")
	fmt.Fprintf(w, "%v", sunrise.Format(dateFormat))
	fmt.Fprintf(w, "</td>")
	fmt.Fprintf(w, "</tr>")
	
	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, "<td>")
	fmt.Fprintf(w, "Sunset:")
	fmt.Fprintf(w, "</td>")
	fmt.Fprintf(w, "<td>")
	fmt.Fprintf(w, "%v", sunset.Format(dateFormat))
	fmt.Fprintf(w, "</td>")
	fmt.Fprintf(w, "</tr>")

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, "<td>")
	fmt.Fprintf(w, "Tzais:")
	fmt.Fprintf(w, "</td>")
	fmt.Fprintf(w, "<td>")
	fmt.Fprintf(w, "%v", tzais.Format(dateFormat))
	fmt.Fprintf(w, "</td>")
	fmt.Fprintf(w, "</tr>")
	
	fmt.Fprintf(w, "</table>")
	
	fmt.Fprintf(w, "</body></html>")
}

const uploadTemplateHTML = `
<html><body>
<form action="{{.}}" method="POST" enctype="multipart/form-data">
Upload File: <input type="file" name="file"><br>
<input type="submit" name="submit" value="Submit">
</form></body></html>
`
var uploadTemplate = template.Must(template.New("upload").Parse(uploadTemplateHTML))

func handleUpload(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
  
	uploadURL, err := blobstore.UploadURL(c, "/postupload", nil)
        if err != nil {
                serveError(c, w, err)
                return
        }
        w.Header().Set("Content-Type", "text/html")
        err = uploadTemplate.Execute(w, uploadURL)
        if err != nil {
                c.Errorf("%v", err)
        }
}

func handlePostUpload(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        blobs, _, err := blobstore.ParseUpload(r)
        if err != nil {
                serveError(c, w, err)
                return
        }
        file := blobs["file"]
        if len(file) == 0 {
                c.Errorf("no file uploaded")
                http.Redirect(w, r, "/upload", http.StatusFound)
                return
        }
        
	http.Redirect(w, r, "/upload", http.StatusFound)
}

func init() {
        http.HandleFunc("/", handleRoot)
        http.HandleFunc("/zmanim", handleZmanim)
	http.HandleFunc("/postzmanim", handlePostZmanim)
        http.HandleFunc("/upload", handleUpload)
        http.HandleFunc("/postupload", handlePostUpload)
	http.HandleFunc("/_ah/mail/", incomingMail)
}