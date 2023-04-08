package handler

import (
	"net/http"

	"github.com/joshDamian/web3storage-service/app"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	app.App().ServeHTTP(w, r)
}
