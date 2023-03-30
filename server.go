package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

type APIServer struct {
	h      *http.Server
	addr   string
	port   string
	read   func(key string) (string, error)
	write  func(key, value string) error
	delete func(key string) error
}

func InitServer(ni *NodeInfo, Read func(key string) (string, error), Write func(key, value string) error, Delete func(key string) error) *APIServer {
	var s APIServer
	s.read = Read
	s.write = Write
	s.delete = Delete
	s.addr = ni.Addr
	s.port = ni.APIPort
	return &s
}

func (s *APIServer) Start() error {
	s.h = &http.Server{
		Addr: s.addr + ":" + s.port,
	}

	http.HandleFunc("/read", s.readHandler)
	http.HandleFunc("/write", s.writeHandler)
	http.HandleFunc("/delete", s.deleteHandler)

	log.Info("Starting server at " + s.addr + ":" + s.port)

	err := s.h.ListenAndServe()
	if err != nil {
		println(err.Error())
		return err
	}
	return nil
}

func (s *APIServer) readHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	log.Infof("Server processing read request for key=%s", key)
	value, err := s.read(key)
	if err != nil {
		if err.Error() == KEY_NOT_FOUND {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

func (s *APIServer) writeHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	log.Infof("Server processing write request for key=%s", key)
	value := r.URL.Query().Get("value")
	err := s.write(key, value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) deleteHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	log.Infof("Server processing delete request for key=%s", key)
	s.delete(key)
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) Stop() {
	s.h.Close()
}
