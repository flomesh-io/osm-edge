/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package main implements osm fd back.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	passfd "github.com/ftrvxmtrx/fd"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	cniSockName           = "/host/var/run/osm-cni.sock"
	transferFdSockName    = "/tmp/bpf-transfer-fd.sock"
	bpfBackServer         = "/host/var/run/bpf-back-server.sock"
	fdServerTransferFdURL = "/v1/transfer-fds"
	fdServerStandbyURL    = "/v1/standby"
	cniTransferFdStartURL = "/v1/cni/transfer-fd"
)

var listener net.Listener

type server struct {
	sync.Mutex
	unixSockPath string
	listeners    []net.Listener
}

func newServer(unixSockPath string) *server {
	if unixSockPath == "" {
		unixSockPath = bpfBackServer
	}
	return &server{
		unixSockPath: unixSockPath,
	}
}

func (s *server) run() (err error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGABRT)

	if err := os.RemoveAll(s.unixSockPath); err != nil {
		log.Fatal(err)
	}
	listener, err = net.Listen("unix", s.unixSockPath)

	if err != nil {
		log.Fatal("listen error:", err)
	}
	r := mux.NewRouter()
	r.Path(fdServerTransferFdURL).
		Methods("POST").
		HandlerFunc(s.transferFd)

	r.Path(fdServerStandbyURL).
		Methods("POST").
		HandlerFunc(s.transferFdBack)

	srv := http.Server{
		Handler:      r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		log.Println("server starting")
		if err := srv.Serve(listener); err != nil {
			log.Fatalf("listenAndServe failed: %v", err)
		}
	}()
	fmt.Println("server started")

	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", cniSockName)
			},
		},
	}
	body := bytes.NewReader([]byte{})
	stopLoop := false

	go func() {
		for {
			_, err = httpc.Post("http://osm-cni"+cniTransferFdStartURL, "application/json", body)
			if err != nil {
				log.Errorf("start transfer fd %s error", err)
				if stopLoop {
					break
				}
				time.Sleep(time.Second * 3)
			} else {
				break
			}
		}
	}()

	<-quit
	stopLoop = true

	// gracefully stop server
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("server stopped")
	return nil
}

func (s *server) transferFd(w http.ResponseWriter, req *http.Request) {
	unixconn, err := net.Dial("unix", transferFdSockName)
	if err != nil {
		log.Errorf("dial unix %s error", err)
		w.WriteHeader(500)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	go func() {
		defer func(unixconn net.Conn) {
			_ = unixconn.Close()
		}(unixconn)
		s.listeners = make([]net.Listener, 0)
		for {
			files, err := passfd.Get(unixconn.(*net.UnixConn), 1, nil)
			if err != nil {
				log.Errorf("passfd get err: %v", err)
				w.WriteHeader(500)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			f := files[0]
			tcpln, err := net.FileListener(f)
			tcpln.Addr()
			if err != nil {
				log.Errorf("listening fd(%v) err: %v", f, err)
				continue
			}
			s.listeners = append(s.listeners, tcpln)
			go func() {
				for {
					_, err := tcpln.Accept()
					if err != nil {
						// only break loop if error.
						break
					}
				}
			}()
			_ = f.Close()
		}
	}()
	w.WriteHeader(200)
	_, _ = w.Write([]byte(""))
}

func (s *server) transferFdBack(w http.ResponseWriter, req *http.Request) {
	log.Debugf("start transferring %d fds...", len(s.listeners))
	if len(s.listeners) > 0 {
		_ = os.Remove(transferFdSockName)
		unix, err := net.Listen("unix", transferFdSockName)
		if err != nil {
			panic(err)
		}

		go func() {
			unixconn, err := unix.Accept()
			if err != nil {
				log.Infof("unix get conn err: %v", err)
			}
			for _, l := range s.listeners {
				tcpln := l.(*net.TCPListener)
				f, err := tcpln.File()
				if err != nil {
					log.Infof("parse listen err: %v", err)
				}
				err = passfd.Put(unixconn.(*net.UnixConn), f)
				if err != nil {
					log.Errorf("passfd put fd err: %v", err)
				}
				_ = l.Close()
				_ = f.Close()
			}
			_ = unixconn.Close()
		}()
		log.Debugf("complete %d fds transfers", len(s.listeners))
		w.WriteHeader(200)
		_, _ = w.Write([]byte(""))
	} else {
		w.WriteHeader(204)
		_, _ = w.Write([]byte(""))
	}
}

func main() {
	flag.Parse()
	s := newServer("")
	var err error
	log.Print("start listening ...")
	err = s.run()
	if err != nil {
		log.Fatalf("listener error: %v", err)
	}
	log.Printf("server stop")
}
