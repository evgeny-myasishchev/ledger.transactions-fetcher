package router

import (
	"net/http"

	"goji.io"
	"goji.io/pat"
)

type gojiRouter struct {
	mux *goji.Mux
}

func (g *gojiRouter) Handle(method string, pattern string, handler http.Handler) {
	g.mux.Handle(pat.NewWithMethods(pattern, method), handler)
}

func (g *gojiRouter) Use(mw MiddlewareFunc) {
	g.mux.Use(func(h http.Handler) http.Handler { return mw(h) })
}

func (g *gojiRouter) pathParam(r *http.Request, name string) string {
	return pat.Param(r, name)
}

func (g *gojiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func createGojiRouter() Router {
	mux := goji.NewMux()
	return &gojiRouter{mux: mux}
}
