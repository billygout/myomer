package myomer

import (
	"net/http"
	"appengine"
	"appengine/mail"
	"fmt"
)

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
	
	body := fmt.Sprintf("You said: \"%s\"\n", msg.Body)
	
	// send reply
	if err = SendMail(c, msg.To[0], msg.Sender, body); err != nil {
		c.Errorf("%v", err)
		return
	}
}