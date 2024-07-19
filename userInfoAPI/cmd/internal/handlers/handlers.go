package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	userstorage "main.go/cmd/userStorage"
)

func GetUserInfo(w http.ResponseWriter, r *http.Request) {

	passportSerieStr := r.URL.Query().Get("passportSerie")
	passportNumberStr := r.URL.Query().Get("passportNumber")

	if passportSerieStr == "" || passportNumberStr == "" {
		http.Error(w, "Missing passportSerie or passportNumber", http.StatusBadRequest)
		return
	}

	passportSerie, err := strconv.Atoi(passportSerieStr)
	if err != nil {
		http.Error(w, "Invalid passportSerie", http.StatusBadRequest)
		return
	}

	passportNumber, err := strconv.Atoi(passportNumberStr)
	if err != nil {
		http.Error(w, "Invalid passportNumber", http.StatusBadRequest)
		return
	}

	userinfo, err := userstorage.UserInfo(passportNumber, passportSerie)
	if err != nil {
		http.Error(w, "Error retrieving user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Установка заголовков и кодирования ответа в JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Установим статус 200 OK
	json.NewEncoder(w).Encode(userinfo)
}
