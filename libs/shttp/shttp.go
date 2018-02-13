package shttp

import (
	"net/http"
	"log"
	"unsafe"
	"sync/atomic"
	"github.com/Packet-Clearing-House/DNSAuth/libs/utils"
	"os"
	"context"
)

type HttpServerConfig struct {
	Addr string	`cfg:"addr; required; netaddr"`
	Acl []string`cfg:"acl; [\"127.0.0.1\", \"::1\"]"`
	Password string `cfg:"password; """`
}
var server *http.Server
var mux *AuthMux = NewAuthMux("", "")

func Handle(pattern string, handler http.Handler) {
	mux.Handle(pattern, handler)
}

func Start(config HttpServerConfig) error {
	if acl, err := utils.ParseACLFromStrings(config.Acl); err != nil {
		return err
	} else {
		mux.ChangeCreds("", config.Password)
		newServer := http.Server{
			Handler: mux,
			ErrorLog: log.New(os.Stdout, "", log.LstdFlags),
		}
		atomic.SwapPointer((*unsafe.Pointer)((unsafe.Pointer)(&server)), unsafe.Pointer(&newServer))

		if listener, err := utils.ACLListen("tcp", config.Addr, acl); err != nil {
			return err
		} else {
			go server.Serve(listener)
		}
	}
	return nil
}

func Reload(config HttpServerConfig) {
	Stop()
	Start(config)
}


func Stop() {
	if err := server.Shutdown(context.Background()); err != nil {
		log.Panic(err)
	}
	if err := server.Close(); err != nil {
		log.Panic(err)
	}
}
