package InfoHandler

import (
	"net/http"

	"github.com/go-chi/render"
)

type Response struct {
	Status int `json:"status"`
}

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		render.JSON(w, r, Response{Status: http.StatusOK})
		return
	}
}
