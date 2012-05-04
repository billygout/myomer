package myomer

import (
	"appengine"
)

type InfoWriter struct {
	appengine.Context
}

func (i InfoWriter) Write(p []byte) (n int, err error){
	i.Context.Infof("%s", p)
	return len(p), nil
}
