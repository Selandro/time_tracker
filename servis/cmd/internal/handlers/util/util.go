package util

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
)

type APIResponse struct {
	Surname    string `json:"surname"`
	Name       string `json:"name"`
	Patronymic string `json:"patronymic"`
	Address    string `json:"address"`
}

// getUserInfoFromAPI выполняет запрос к внешнему API для получения информации о пользователе
func GetUserInfoFromAPI(log *slog.Logger, passportSerie, passportNumber int) (APIResponse, error) {
	var apiResponse APIResponse
	url := fmt.Sprintf("http://localhost:8081/userinfo?passportSerie=%d&passportNumber=%d", passportSerie, passportNumber)

	log.Debug("Sending request to external API", slog.String("url", url))
	resp, err := http.Get(url)
	if err != nil {
		log.Error("Failed to get response from API", slog.String("url", url), slog.String("error", err.Error()))
		return apiResponse, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("Failed to get user info from API", slog.String("url", url), slog.Int("status_code", resp.StatusCode))
		return apiResponse, fmt.Errorf("failed to get user info, status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read API response body", slog.String("url", url), slog.String("error", err.Error()))
		return apiResponse, err
	}

	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		log.Error("Failed to unmarshal API response", slog.String("url", url), slog.String("error", err.Error()))
		return apiResponse, err
	}

	log.Debug("Successfully retrieved user info from API", slog.Any("apiResponse", apiResponse))
	return apiResponse, nil
}

// addUserToDB добавляет нового пользователя в базу данных и возвращает его ID
func AddUserToDB(log *slog.Logger, db *sql.DB, passportSerie, passportNumber int, apiResponse APIResponse) (int, error) {
	var userID int
	log.Debug("Adding user to database", slog.Any("apiResponse", apiResponse))
	err := db.QueryRow(`
		INSERT INTO users (passport_serie, passport_number, surname, name, patronymic, address)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, passportSerie, passportNumber, apiResponse.Surname, apiResponse.Name, apiResponse.Patronymic, apiResponse.Address).Scan(&userID)
	if err != nil {
		log.Error("Failed to add user to database", slog.Any("apiResponse", apiResponse), slog.String("error", err.Error()))
		return 0, err
	}
	log.Info("User successfully added to database", slog.Int("userID", userID))
	return userID, nil
}
