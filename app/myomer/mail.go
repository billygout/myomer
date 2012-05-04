package myomer

import (
	"net/http"
	"appengine"
	"io"
	appmail "appengine/mail"
	"net/mail"
	"errors"
	"fmt"
	"bufio"
	"strings"
)

var mail_seqnum = int64(0)

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

func parseBody(body io.Reader) (plain, html string, err error) {
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
	} else {
		plain = strings.TrimSpace(plain)
		if plain[0] == '*' {
			plain = plain[1:]
		}
		if plain[len(plain)-1] == '*' {
			plain = plain[0:len(plain)-1]
		}
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

	/* optional fields ...
	 */
	
	replyto := msg.Header.Get("ReplyTo")
	subject := msg.Header.Get("Subject")
	
	// parse body
	if body, html, err = parseBody(msg.Body); err != nil {
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

	/*
	iw := InfoWriter{c}

	c.Infof("BEGIN BODY...")
	io.Copy(iw, msg.Body)
	c.Infof("...END BODY")
	 */
	
	 if appmsg, err = parseMessage(msg); err != nil {
		c.Infof("%v", err)
		return
	} 
	
	c.Infof("Received mail: %+v", appmsg)
	
	body := fmt.Sprintf("You said: \"%s\"\n(seqnum: %v)\n", appmsg.Body, mail_seqnum)
	
	// sent reply
	if err = sendMail(c, appmsg.To[0], appmsg.Sender, body); err != nil {
		c.Infof("Error sending reply: %v", err) 
	}
        
	mail_seqnum++
}