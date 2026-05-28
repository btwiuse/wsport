//go:build !js || !wasm

package main

import (
	"net/http"

	term "github.com/webteleport/ufo/apps/term/handler"
)

func termHandler() http.Handler {
	return term.Handler()
}
