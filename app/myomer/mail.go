package myomer

import (
	"net/http"
	"appengine"
	"io"
	appmail "appengine/mail"
	"net/mail"
	"errors"
)

type InfoWriter struct {
	appengine.Context
}

func (i InfoWriter) Write(p []byte) (n int, err error){
	i.Context.Infof("%s", p)
	return len(p), nil
}

func addressListToStringSlice(list []*mail.Address) []string {
	sl := make([]string, len(list))

	for i, address := range list {
		sl[i] = address.Address
	}

	return sl
}

func parseMessage(msg *mail.Message) (*appmail.Message, error) {

	var from string
	var to []*mail.Address
	var cc []*mail.Address
	var bcc []*mail.Address

	// required: "From"
	if from = msg.Header.Get("From"); from == "" {
		return nil, errors.New("failed to parse \"From\"")
	}

	// required: one of "To", "Cc", "Bcc"
	to, _  = msg.Header.AddressList("To")
	cc, _  = msg.Header.AddressList("Cc")
	bcc, _ = msg.Header.AddressList("Bcc")

	if to == nil && cc == nil && bcc == nil {
		return nil, errors.New("failed to parse any of \"To\", \"Cc\", and \"Bcc\"")
	}

	/* optional fields ...
	 */
	
	replyto := msg.Header.Get("ReplyTo")
	subject := msg.Header.Get("Subject")
	
	return &appmail.Message{
	Sender: from,
	ReplyTo: replyto,
	To: addressListToStringSlice(to),
	Cc: addressListToStringSlice(cc),
	Bcc: addressListToStringSlice(bcc),
	Subject: subject,
	}, nil
}

func sendMail(c appengine.Context, from, to, body string) error {
	msg := &appmail.Message{
        Sender:  from,
	ReplyTo: from,
        To:      []string{to},
        Body:    body,
        }
        
	return appmail.Send(c, msg)
}

func incomingMail(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
        //defer r.Body.Close()
        
	var err error
	var msg  *mail.Message
	var appmsg *appmail.Message
	
	if msg, err = mail.ReadMessage(r.Body); err != nil {
		c.Infof("Error: %v", err)
		return
	}

	if appmsg, err = parseMessage(msg); err != nil {
		c.Infof("error parsing message: %v", err)
		return
	} 

	c.Infof("Received mail: %+v", appmsg)

	
	iw := InfoWriter{c}
	
	if _, err = io.Copy(iw, msg.Body); err != nil {
		c.Infof("Error: %v", err)
	} 
        

	// sent reply
	if err = sendMail(c, appmsg.To[0], appmsg.Sender, "Automatic Reply..."); err != nil {
		c.Infof("Error sending reply: %v", err) 
	}
}