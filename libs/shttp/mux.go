package shttp

import (
	"log"
	"net/http"
	"sync/atomic"
	"unsafe"
)

func NewAuthMux(login, pass string) *AuthMux {

	return &AuthMux{
		*http.NewServeMux(),
		&login,
		&pass,
	}
}

type AuthMux struct {
	http.ServeMux
	login, pass *string
}

func (m *AuthMux) ChangeCreds(login, pass string) {
	atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.login)), unsafe.Pointer(&login))
	atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&m.pass)), unsafe.Pointer(&pass))
}

func (m *AuthMux) Handle(pattern string, handler http.Handler) {
	m.ServeMux.Handle(pattern, m.HTTPBasicAuth(handler))
}

func (m *AuthMux) HTTPBasicAuth(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		login, password, _ := r.BasicAuth()

		if *m.login != login || password != *m.pass {
			log.Println("[ERROR] Unauthorized HTTP access: wrong login/password!")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	}
}
