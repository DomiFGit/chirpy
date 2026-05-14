package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	type errorResp struct {
		Error string `json:"error"`
	}
	type params struct {
		Email string `json:"email"`
	}

	var p params
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		resp, _ := json.Marshal(errorResp{Error: "wrong payload format"})
		w.Write(resp)
		return
	}

	user, err := h.dbQueries.CreateUser(r.Context(), p.Email)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		resp, _ := json.Marshal(errorResp{Error: "server error creating user"})
		w.Write(resp)
		return
	}

	serializedUser, err := json.Marshal(user)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		resp, _ := json.Marshal(errorResp{Error: "response could not be serialized"})
		w.Write(resp)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(serializedUser)
}
