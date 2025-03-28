package controller

import (
	"encoding/json"
	"net/http"
	"slices"
)

type Router struct {
	Middleware []func(http.HandlerFunc) http.HandlerFunc
}

func NewRouter() *Router {
	return new(Router)
}

func (r *Router) Chain(handler ...func(http.HandlerFunc) http.HandlerFunc) *Router {
	r.Middleware = append(r.Middleware, handler...)
	return r
}
func (r *Router) Post(path string, handler http.HandlerFunc) *Router {
	r.Add(path, handler, http.MethodPost)
	return r
}
func (r *Router) Get(path string, handler http.HandlerFunc) *Router {
	r.Add(path, handler, http.MethodGet)
	return r
}
func (r *Router) Delete(path string, handler http.HandlerFunc) *Router {
	r.Add(path, handler, http.MethodDelete)
	return r
}
func (r *Router) Add(path string, handler http.HandlerFunc, method string) *Router {
	for i := len(r.Middleware) - 1; i >= 0; i-- {
		handler = r.Middleware[i](handler)
	}
	if len(method) > 0 {
		handler = r.MethodGuard(handler, []string{method})
	}
	http.HandleFunc(path, handler)
	return r
}

func (r *Router) MethodGuard(handler http.HandlerFunc, methods []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !slices.Contains(methods, r.Method) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"Response": "Unsupported method"})
			return
		}
		handler(w, r)
	}
}
