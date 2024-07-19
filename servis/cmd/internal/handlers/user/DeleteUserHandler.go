package user

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
)

// @Summary Delete a user
// @Description Delete a user and their tasks from the database
// @Tags User
// @Accept json
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {string} string "Success message"
// @Failure 400 {string} string "Invalid user_id parameter"
// @Failure 500 {string} string "Failed to delete user"
// @Router /api/v1/users [delete]
func DeleteUserHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			log.Error("Missing user_id parameter")
			http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			log.Error("Invalid user_id parameter", slog.String("error", err.Error()), slog.String("user_id", userIDStr))
			http.Error(w, "Invalid user_id parameter", http.StatusBadRequest)
			return
		}

		log.Debug("Deleting user's tasks", slog.Int("userID", userID))
		_, err = db.Exec("DELETE FROM users_tasks WHERE user_id = $1", userID)
		if err != nil {
			log.Error("Failed to delete user's tasks", slog.String("error", err.Error()), slog.Int("userID", userID))
			http.Error(w, fmt.Sprintf("Failed to delete user's tasks: %v", err), http.StatusInternalServerError)
			return
		}

		log.Debug("Deleting user", slog.Int("userID", userID))
		_, err = db.Exec("DELETE FROM users WHERE id = $1", userID)
		if err != nil {
			log.Error("Failed to delete user", slog.String("error", err.Error()), slog.Int("userID", userID))
			http.Error(w, fmt.Sprintf("Failed to delete user: %v", err), http.StatusInternalServerError)
			return
		}

		log.Info("User and their tasks deleted", slog.Int("userID", userID))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "User with ID %s and their tasks have been deleted", userIDStr)
	}
}
