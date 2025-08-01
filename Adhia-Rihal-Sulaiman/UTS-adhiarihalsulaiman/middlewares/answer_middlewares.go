package middlewares

import (
	"uts_adhia/database"
	"uts_adhia/utils"

	"database/sql"
	"net/http"
)

func OwnsAnswerMiddleware(next http.Handler, answerID string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUserID := r.Context().Value(utils.UserIDKey)
		userID, ok := ctxUserID.(string)
		if !ok || userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if answerID == "" {
			http.Error(w, "answer_id is null", http.StatusBadRequest)
			return
		}

		var answerOwnerID string
		err := database.DB.QueryRow("SELECT user_id FROM answers WHERE answer_id = ? AND deleted_at IS NULL", answerID).Scan(&answerOwnerID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Answer not found", http.StatusNotFound)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		if answerOwnerID != userID {
			http.Error(w, "Forbidden - the answer isn't yours", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
