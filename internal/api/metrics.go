package api

import (
	"net/http"
)

func (config *APIConfig) MiddlewareMetricsInc(realHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		config.FileserverHits.Add(1)
		realHandler.ServeHTTP(writer, request)
	})
}
