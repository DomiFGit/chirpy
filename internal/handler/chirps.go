package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"github.com/domifgit/chirpy/internal/database"
)

func filterChirp(text string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	pattern := "(?i)" + strings.Join(badWords, "|")
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(text, "****")
}

func (h *Handler) InsertChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
		User string `json:"user_id"`
	}
	type errorResp struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(r.Body)
	var c chirp
	if err := decoder.Decode(&c); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		resp, _ := json.Marshal(errorResp{Error: "Could not decode response"})
		w.Write(resp)
		return
	}
	if len(c.Body) > 140 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		resp, _ := json.Marshal(errorResp{Error: "Chirp is too long"})
		w.Write(resp)
		return
	}

	cleanedBody := filterChirp(c.Body)
	userID, err := uuid.Parse(c.User)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		resp, _ := json.Marshal(errorResp{Error: "Invalid user ID"})
		w.Write(resp)
		return
	}
	createChirpParams := database.CreateChirpParams{UserID: userID, Body: cleanedBody}
	createdChirp, err := h.dbQueries.CreateChirp(r.Context(), createChirpParams)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		resp, _ := json.Marshal(errorResp{Error: "Could not create chirp"})
		w.Write(resp)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(createdChirp)
}
