package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type APIServer struct {
	h      *http.Server
	addr   string
	port   string
	read   func(key string) (string, error)
	write  func(key, value string) error
	delete func(key string) error
	repair func(otherNode string) error
	node   *Node
}

// TODO: Refactor long argument list
func InitServer(n *Node, Read func(key string) (string, error), Write func(key, value string) error, Delete func(key string) error, Repair func(otherNode string) error) *APIServer {
	var s APIServer
	s.read = Read
	s.write = Write
	s.delete = Delete
	s.repair = Repair
	s.addr = n.Info.Addr
	s.port = n.Info.APIPort
	s.node = n
	return &s
}

func (s *APIServer) Start() error {
	s.h = &http.Server{
		Addr: s.addr + ":" + s.port,
	}

	http.HandleFunc("/read", s.readHandler)
	http.HandleFunc("/write", s.writeHandler)
	http.HandleFunc("/delete", s.deleteHandler)
	http.HandleFunc("/repair", s.repairHandler)
	http.HandleFunc("/graph/add-edge", s.addEdgeHandler)
	http.HandleFunc("/graph/remove-edge", s.removeEdgeHandler)
	http.HandleFunc("/graph/get-neighbours", s.getNeighboursHandler)
	http.HandleFunc("/graph/get-degrees", s.getDegreesBetweenHandler)
	http.HandleFunc("/graph/get-mutual", s.getMutualVerticesHandler)
	http.HandleFunc("/graph/recon", s.requestGraphRecon)
	http.HandleFunc("/image/store", s.storeImageHandler)
	http.HandleFunc("/image/get", s.getImageHandler)

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
	err := s.delete(key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) repairHandler(w http.ResponseWriter, r *http.Request) {
	otherNode := r.URL.Query().Get("node")
	log.Infof("Server processing repair request with node=%s", otherNode)
	err := s.repair(otherNode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) addEdgeHandler(w http.ResponseWriter, r *http.Request) {
	v1 := r.URL.Query().Get("v1")
	v2 := r.URL.Query().Get("v2")
	log.Infof("Server processing add edge request for v1=%s, v2=%s", v1, v2)
	err := s.addDirectedEdge(v1, v2)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = s.addDirectedEdge(v2, v1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// s.node.RequestGraphRecon()

	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) addDirectedEdge(v1 string, v2 string) error {
	s.node.mu.Lock()
	if _, ok := s.node.Graph.batchedOps[v1]; !ok {
		s.node.Graph.batchedOps[v1] = make([]string, 0)
	}
	s.node.Graph.batchedOps[v1] = append(s.node.Graph.batchedOps[v1], v2)
	s.node.Graph.numBatchedOps++
	s.node.mu.Unlock()
	if s.node.Graph.numBatchedOps >= BATCH_SIZE {
		s.node.Graph.ApplyBatchedOps()
	}
	return nil
}

func (s *APIServer) getNeighboursHandler(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query().Get("v")
	log.Infof("Server processing get neighbours request for v=%s", v)
	v1Neighbours, err := s.read(v)
	if err != nil {
		if err.Error() == KEY_NOT_FOUND {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(v1Neighbours))
}

func (s *APIServer) removeEdgeHandler(w http.ResponseWriter, r *http.Request) {
	v1 := r.URL.Query().Get("v1")
	v2 := r.URL.Query().Get("v2")
	log.Infof("Server processing remove edge request for v1=%s, v2=%s", v1, v2)
	err := s.removeDirectedEdge(v1, v2)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = s.removeDirectedEdge(v2, v1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// s.node.RequestGraphRecon()

	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) removeDirectedEdge(v1 string, v2 string) error {
	v1Neighbours, err := s.read(v1)
	var vn VertexNeighbours
	if err != nil && err.Error() == KEY_NOT_FOUND {
		return nil
	} else if err != nil {
		return err
	} else {
		err = json.Unmarshal([]byte(v1Neighbours), &vn)
		if err != nil {
			return err
		}
	}
	var newNeighbours []string
	for _, n := range vn.Neighbours {
		if n != v2 {
			newNeighbours = append(newNeighbours, n)
		}
	}
	vn.Neighbours = newNeighbours
	v1NeighboursBytes, err := json.Marshal(vn)
	if err != nil {
		return err
	}
	err = s.write(v1, string(v1NeighboursBytes))
	if err != nil {
		return err
	}
	return nil
}

func (s *APIServer) getDegreesBetweenHandler(w http.ResponseWriter, r *http.Request) {
	v1 := r.URL.Query().Get("v1")
	v2 := r.URL.Query().Get("v2")
	log.Infof("Server processing get degrees between request for v1=%s, v2=%s", v1, v2)
	degrees := s.node.Graph.FindDegreeBetween(v1, v2)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.Itoa(degrees)))
}

func (s *APIServer) getMutualVerticesHandler(w http.ResponseWriter, r *http.Request) {
	v1 := r.URL.Query().Get("v1")
	v2 := r.URL.Query().Get("v2")
	log.Infof("Server processing get mutual vertices request for v1=%s, v2=%s", v1, v2)
	mutualVertices := s.node.Graph.GetMutualVertices(v1, v2)
	mutualVerticesBytes, err := json.Marshal(mutualVertices)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(mutualVerticesBytes)
}

func (s *APIServer) requestGraphRecon(w http.ResponseWriter, r *http.Request) {
	log.Infof("Server processing request graph recon request")
	if s.node.Graph.numBatchedOps > 0 {
		s.node.Graph.ApplyBatchedOps()
	}
	s.node.RequestGraphRecon()
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) storeImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("Server processing store image request")
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	name := r.URL.Query().Get("name")

	bufstr := base64.StdEncoding.EncodeToString(buf)

	err = s.write(name, bufstr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) getImageHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("Server processing get image request")
	name := r.URL.Query().Get("name")
	img, err := s.read(name)
	if err != nil {
		if err.Error() == KEY_NOT_FOUND {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("[]"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	imgBytes, err := base64.StdEncoding.DecodeString(img)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write(imgBytes)
}

func (s *APIServer) Stop() {
	s.h.Close()
}

type VertexNeighbours struct {
	Neighbours []string `json:"neighbours"`
}

const (
	BATCH_SIZE = 10000
)
