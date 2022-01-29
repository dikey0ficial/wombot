package main

import (
	"context"
	"fmt"
	gomux "github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func nMux() (mux *gomux.Router) {
	mux = gomux.NewRouter()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/ping", pingHandler)
	return mux
}

// Server for heroku
type Server struct {
	Mux          *gomux.Router
	SelfPingTime time.Duration
}

// NewServer returns new standart *Server
func NewServer() *Server {
	return &Server{
		Mux:          nMux(),
		SelfPingTime: time.Minute,
	}
}

func (s *Server) selfPing(ctx context.Context) {
	ticker := time.NewTicker(s.SelfPingTime)
	for {
		select {
		case <-ticker.C:
			resp, err := http.Get("http://wombatobot.herokuapp.com:80/ping")
			if err != nil {
				errl.Println("SelfPing:", err)
				continue
			}
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errl.Println("SelfPing:", err)
				resp.Body.Close()
				continue
			}
			resp.Body.Close()
			if string(b) != "pong" {
				errl.Println("SelfPing: got wrong ping")
			}
			resp.Body.Close()
		case <-ctx.Done():
			return
		}
	}
}

// Run _
func (s *Server) Run() error {
	if s == nil {
		return fmt.Errorf("Server haven't initialized")
	}
	go s.selfPing(ctx)
	return http.ListenAndServe(":80", mw(s.Mux))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		w.Header().Set("Allow", "GET")
		fmt.Fprint(w, "Method not allowed")
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<p>developing...</p>\n")
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	fmt.Fprint(w, "pong")
}

func mw(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
			}
			servl.Printf("FromAddr:%s;ReqAddr:%s;Meth:%s;\n", r.RemoteAddr, r.URL.Path, r.Method)
			next.ServeHTTP(w, r)
		},
	)
}
