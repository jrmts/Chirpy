package api

import (
	"strings"
)

func badWordReplace(chirp string) string {
	splitedchirp := strings.Split(chirp, " ")
	var sliceCleanedChirp []string
	var cleanedChirp string
	for _, word := range splitedchirp {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			sliceCleanedChirp = append(sliceCleanedChirp, "****") // "****"
		} else {
			sliceCleanedChirp = append(sliceCleanedChirp, word)
		}
	}
	cleanedChirp = strings.Join(sliceCleanedChirp, " ")
	return cleanedChirp
}
