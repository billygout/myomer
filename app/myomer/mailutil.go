package myomer

import (

	"net/http"
	"net/mail"

	"appengine"
	appmail "appengine/mail"
	
	"io"
	"bufio"
	"errors"
	"fmt"
	"strings"
)

func addressListToStringSlice(list []*mail.Address) []string {
	sl := make([]string, len(list))

	for i, address := range list {
		sl[i] = address.Address
	}

	return sl
}

func parseBody(body io.Reader, is_multipart bool) (plain, html string, err error) {
	var str string
	var delim string = ""
	var in_plain bool = false
	var in_html bool = false
	var need_skip bool = false
	var is_complex = false
	var simple_text string = ""
	
	var line []byte
	var isPrefix bool

	plain = ""
	html = ""
	err = nil

	bufr := bufio.NewReader(body)
	for {
		if line, isPrefix, err = bufr.ReadLine(); err != nil || isPrefix  {
			if err != nil && err != io.EOF {
				return "", "", errors.New(fmt.Sprintf("error parsing body: %v", err))
			} else if isPrefix {
				return "", "", errors.New("error parsing body: parsed incomplete line")
			}
			
			break
		} else {
			str = (string(line))
			simple_text += str + "\n"
			switch {
			case delim == "":
				delim = str
			case strings.Contains(str, delim):
				in_plain = false
				in_html = false
			case strings.HasPrefix(str, "Content-Type:"):
				is_complex = true
				if strings.Contains(str, "text/plain") {
					in_plain = true
				} else if strings.Contains(str, "text/html") {
					in_html = true
				}
				need_skip = true
			case need_skip:
				need_skip = false
			default:
				if in_plain {
					// for plain section of a multipart message
					// omit all asterisks (*)
					if is_multipart {
						str = strings.Replace(str, "*", "", -1)
					}
					
					plain += str + "\n"
				} else if in_html {
					html += str + "\n"
				}
			}
			
		}
	}

	if !is_complex {
		plain = simple_text
		html = ""
	}
	
	err = nil
	return
}

func parseMessage(msg *mail.Message) (*appmail.Message, error) {

	var from string
	var to []*mail.Address
	var cc []*mail.Address
	var bcc []*mail.Address
	var body, html string
	var contenttype string
	var is_multipart bool
	var err error

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


	// "Content-Type"
	contenttype = msg.Header.Get("Content-Type")
	is_multipart = strings.Contains(contenttype, "multipart")

	/* optional fields ...
	 */
	
	replyto := msg.Header.Get("ReplyTo")
	subject := msg.Header.Get("Subject")
	
	// parse body
	if body, html, err = parseBody(msg.Body, is_multipart); err != nil {
		return nil, err
	} 
	
	return &appmail.Message{
	Sender: from,
	ReplyTo: replyto,
	To: addressListToStringSlice(to),
	Cc: addressListToStringSlice(cc),
	Bcc: addressListToStringSlice(bcc),
	Subject: subject,
	Body: body,
	HTMLBody: html,
	}, nil
}

func dumpMessage(c appengine.Context, msg *mail.Message) {
	iw := InfoWriter{c}
	
	c.Infof("--BEGIN MESSAGE")
	io.Copy(iw, msg.Body)
	c.Infof("END MESSAGE--")
}

func GetMessageFromRequest(r *http.Request) (*appmail.Message, error) {
	var err error
	var msg  *mail.Message
	var appmsg *appmail.Message
	
	
	if msg, err = mail.ReadMessage(r.Body); err != nil {
		return nil, err
	}
	
	/*
	c := appengine.NewContext(r)
	c.Infof("msg: %+v", msg)
	dumpMessage(c, msg)
	return nil, errors.New("just testing")
        */
	
	if appmsg, err = parseMessage(msg); err != nil {
		return nil, err
	} 
	
	return appmsg, nil
}

func SendMail(c appengine.Context, from, to, body string) error {
	msg := &appmail.Message{
        Sender:  from,
	ReplyTo: from,
        To:      []string{to},
        Body:    body,
        }
        
	return appmail.Send(c, msg)
}

