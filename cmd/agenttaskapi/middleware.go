package main

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

func mwLogger(fn httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, rp httprouter.Params) {
		start := time.Now()
		defer func() {
			log.Infof("%s Completed %s for %s in %v", r.RemoteAddr, r.Method, r.URL.String(), time.Since(start))
		}()

		fn(w, r, rp)
	}
}

func httpRedirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	log.Infof("%s Redirected to %s with %d", r.RemoteAddr, url, code)
	http.Redirect(w, r, url, code)
}
