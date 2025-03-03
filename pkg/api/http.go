package api

import (
	"encoding/json"
	"net/http"

	"raft-kv-store/pkg/kvstore"
	"raft-kv-store/pkg/raft"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	store    *kvstore.Store
	raftNode *raft.RaftNode
}

func NewHTTPServer(store *kvstore.Store, raftNode *raft.RaftNode) *HTTPServer {
	return &HTTPServer{
		store:    store,
		raftNode: raftNode,
	}
}

func (s *HTTPServer) HandleGetKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := s.store.Get(key)
	if err != nil {
		if err == kvstore.ErrKeyNotFound {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"value": value})
}

func (s *HTTPServer) HandlePutKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value := r.URL.Query().Get("value")
	if value == "" {
		http.Error(w, "Value is required", http.StatusBadRequest)
		return
	}

	if err := s.store.Propose(key, value); err != nil {
		if err == raft.ErrNotLeader {
			leader := s.raftNode.GetLeader()
			http.Redirect(w, r, leader+"/key/"+key, http.StatusTemporaryRedirect)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *HTTPServer) HandleClusterStatus(w http.ResponseWriter, r *http.Request) {
	status := s.raftNode.GetClusterStatus()
	respondWithJSON(w, http.StatusOK, status)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
