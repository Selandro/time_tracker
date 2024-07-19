package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"main.go/cmd/internal/storage/cache"
	model "main.go/tracker_model"
)

// UpdateUserHandler обрабатывает запросы на изменение данных пользователя.
// @Summary Update a user
// @Description Update user details
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body Users true "User details"
// @Success 200 {object} Users "Updated user details"
// @Failure 400 {string} string "Invalid user ID or input"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Failed to update user"
// @Router /api/v1/users/{id} [put]
func UpdateUserHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/update_user/")
		userID, err := strconv.Atoi(idStr)
		if err != nil {
			log.Error("Invalid user ID", slog.String("idStr", idStr), slog.String("error", err.Error()))
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var user model.Users
		err = json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			log.Error("Invalid input", slog.String("error", err.Error()))
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		user.UserID = userID

		log.Debug("Updating user", slog.Any("user", user))
		result, err := db.Exec(`
			UPDATE users
			SET passport_serie = $2, passport_number = $3, surname = $4, name = $5, patronymic = $6, address = $7
			WHERE id = $1
		`, user.UserID, user.PassportSerie, user.PassportNumber, user.Surname, user.Name, user.Patronymic, user.Address)
		if err != nil {
			log.Error("Failed to update user", slog.Any("user", user), slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Failed to update user: %v", err), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Error("Error checking rows affected", slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Error checking rows affected: %v", err), http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			log.Info("No user found with the given ID", slog.Int("userID", userID))
			http.Error(w, "No user found with the given ID", http.StatusNotFound)
			return
		}

		cache.UserCacheMutex.Lock()
		defer cache.UserCacheMutex.Unlock()

		if existingUser, ok := cache.UserCache[user.UserID]; ok {
			existingUser.PassportSerie = user.PassportSerie
			existingUser.PassportNumber = user.PassportNumber
			existingUser.Surname = user.Surname
			existingUser.Name = user.Name
			existingUser.Patronymic = user.Patronymic
			existingUser.Address = user.Address
			cache.UserCache[user.UserID] = existingUser
			log.Debug("Updated user in cache", slog.Any("user", existingUser))
		} else {
			log.Debug("User not found in cache", slog.Int("userID", user.UserID))
		}

		log.Info("User updated successfully", slog.Any("user", user))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}
