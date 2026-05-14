package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

func filterChirp(text string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	pattern := "(?i)" + strings.Join(badWords, "|")
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(text, "****")
}

func (h *Handler) ValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}
	type errorResp struct {
		Error string `json:"error"`
	}
	type validResp struct {
		CleanedBody string `json:"cleaned_body"`
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	clean := filterChirp(c.Body)
	resp, _ := json.Marshal(validResp{CleanedBody: clean})
	w.Write(resp)
}
