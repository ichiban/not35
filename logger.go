package main

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()

		next.ServeHTTP(w, r)

		logrus.WithFields(logrus.Fields{
			"proto":   r.Proto,
			"from":    r.RemoteAddr,
			"url":     r.RequestURI,
			"elapsed": int64(time.Since(t) / time.Millisecond),
		}).Info(r.Method)
	})
}
