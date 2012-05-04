package myomer

import (
	"net/http"
	"appengine"
	"appengine/mail"
	"strings"
	"fmt"
)

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	slice := strings.SplitN(s, "\n", 2)
	if len(slice) == 0 {
		return ""
	}

	return slice[0]
}

func incomingMail(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var err error
	var msg *mail.Message

	// get email message from the request
	if msg, err = GetMessageFromRequest(r); err != nil {
		c.Errorf("%v", err)
		return
	}	

	c.Infof("Received mail: %+v", msg)
	
	var body string
	
	// get zmanim for given zipcode
	first_line := firstLine(msg.Body)
	if zmanim_string := getZmanimString(c, first_line); len(zmanim_string) > 0 {
		body = zmanim_string
	} else {
		body = fmt.Sprintf("failed to retrieve zmanim for \"%v\"", first_line)
	}
	
	if err = SendMail(c, msg.To[0], msg.Sender, body); err != nil {
		c.Errorf("%v", err)
		return
	}
}