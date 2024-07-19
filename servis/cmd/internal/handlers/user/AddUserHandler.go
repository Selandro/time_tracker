package user

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"log/slog"

	"main.go/cmd/internal/handlers/util"
	"main.go/cmd/internal/storage/cache"
	model "main.go/tracker_model"
)

type UserInput struct {
	PassportNumber string `json:"passportNumber"`
}

// @Summary Add a new user
// @Description Add a new user to the database
// @Tags User
// @Accept json
// @Produce json
// @Param user body UserInput true "User Input"
// @Success 201 {integer} int "User ID"
// @Failure 400 {string} string "Invalid input"
// @Failure 500 {string} string "Failed to add user"
// @Router /api/v1/users [post]
// addUserHandler обрабатывает запросы на добавление нового пользователя
func AddUserHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input UserInput

		// Декодирование JSON данных
		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			log.Error("Invalid input format", slog.String("error", err.Error()))
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		log.Debug("Received user input", slog.Any("input", input))

		// Разделение серии и номера паспорта
		parts := strings.Split(input.PassportNumber, " ")
		if len(parts) != 2 {
			log.Warn("Invalid passport number format", slog.String("passportNumber", input.PassportNumber))
			http.Error(w, "Invalid passport number format", http.StatusBadRequest)
			return
		}

		passportSerie, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Error("Invalid passport serie", slog.String("error", err.Error()), slog.String("passportSerie", parts[0]))
			http.Error(w, "Invalid passport serie", http.StatusBadRequest)
			return
		}

		passportNumber, err := strconv.Atoi(parts[1])
		if err != nil {
			log.Error("Invalid passport number", slog.String("error", err.Error()), slog.String("passportNumber", parts[1]))
			http.Error(w, "Invalid passport number", http.StatusBadRequest)
			return
		}

		log.Debug("Parsed passport details", slog.Int("passportSerie", passportSerie), slog.Int("passportNumber", passportNumber))

		// Получение информации о пользователе из внешнего API
		apiResponse, err := util.GetUserInfoFromAPI(log, passportSerie, passportNumber)
		if err != nil {
			log.Error("Failed to get user info from API", slog.String("error", err.Error()))
			http.Error(w, "Failed to get user info from API", http.StatusInternalServerError)
			return
		}

		log.Debug("Received API response", slog.Any("apiResponse", apiResponse))

		// Вставка нового пользователя в базу данных
		userID, err := util.AddUserToDB(log, db, passportSerie, passportNumber, apiResponse)
		if err != nil {
			log.Error("Failed to add user to database", slog.String("error", err.Error()))
			http.Error(w, "Failed to add user", http.StatusInternalServerError)
			return
		}

		log.Info("User added to database", slog.Int("userID", userID))

		// Обновление кэша
		cache.CacheUser(model.Users{
			UserID:         userID,
			PassportSerie:  passportSerie,
			PassportNumber: passportNumber,
			Surname:        apiResponse.Surname,
			Name:           apiResponse.Name,
			Patronymic:     apiResponse.Patronymic,
			Address:        apiResponse.Address,
		})

		log.Info("User cached", slog.Int("userID", userID))

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(userID)

		log.Info("Response sent", slog.Int("userID", userID))
	}
}
