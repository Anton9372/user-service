package metric

import (
	"Users/pkg/logging"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const (
	URL = "/api/heartbeat"
)

type Handler struct {
	Logger *logging.Logger
}

func NewHandler(logger *logging.Logger) *Handler {
	return &Handler{
		Logger: logger,
	}
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, URL, h.Heartbeat)
}

// Heartbeat
// @Summary 	Heartbeat
// @Description Checks that the server is up and running
// @Tags 		Heartbeat
// @Success 	204
// @Router 		/metric [get]
func (h *Handler) Heartbeat(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(204)
}
