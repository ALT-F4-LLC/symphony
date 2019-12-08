package services

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/sirupsen/logrus"
)

// RegisterHandler : registers HTTP handler with logrus output
func RegisterHandler(h http.Handler) http.Handler {
	return handlers.LoggingHandler(logrus.New().Writer(), h)
}
