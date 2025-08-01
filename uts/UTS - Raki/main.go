// main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql" // Blank import to register the MySQL driver
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/artichys/uts-raki/handlers"
	"github.com/artichys/uts-raki/middleware"
	"github.com/artichys/uts-raki/repository"
	"github.com/artichys/uts-raki/utils"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// Initialize database connection
	utils.InitDB()
	defer utils.CloseDB() // Ensure DB connection is closed when main exits

	// Initialize repositories
	userRepo := repository.NewUserRepository(utils.DB)
	questionRepo := repository.NewQuestionRepository(utils.DB)
	answerRepo := repository.NewAnswerRepository(utils.DB)
	sessionRepo := repository.NewSessionRepository(utils.DB) // NEW: Initialize SessionRepository

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo)
	userHandler := handlers.NewUserHandler(userRepo)
	questionHandler := handlers.NewQuestionHandler(questionRepo, userRepo)
	answerHandler := handlers.NewAnswerHandler(answerRepo, questionRepo, userRepo)

	router := mux.NewRouter()

	// --- Public Routes (No Authentication Needed) ---
	router.HandleFunc("/api/register", authHandler.RegisterUser).Methods("POST")
	router.HandleFunc("/api/login", authHandler.LoginUser).Methods("POST")
	router.HandleFunc("/api/forgot-password/initiate", authHandler.InitiatePasswordReset).Methods("POST")
	router.HandleFunc("/api/forgot-password/reset", authHandler.ResetPassword).Methods("POST")
	router.HandleFunc("/api/questions", questionHandler.GetAllQuestions).Methods("GET") // Public, but promoted questions appear first
	router.HandleFunc("/api/questions/{id}", questionHandler.GetQuestionByID).Methods("GET")
	router.HandleFunc("/api/questions/{question_id}/answers", answerHandler.GetAnswersByQuestionID).Methods("GET")


	// --- Authenticated Routes (Requires Bearer Token) ---
	authRouter := router.PathPrefix("/api").Subrouter()
	authRouter.Use(func(next http.Handler) http.Handler {
        return middleware.AuthMiddleware(sessionRepo, next)
    })

	// User Routes
	authRouter.HandleFunc("/users/me", userHandler.GetMyProfile).Methods("GET")
	authRouter.HandleFunc("/users/{id}", userHandler.GetUserByID).Methods("GET")

	// Question Routes
	authRouter.HandleFunc("/questions", questionHandler.CreateQuestion).Methods("POST")
	authRouter.HandleFunc("/questions/my", questionHandler.GetMyQuestions).Methods("GET")
	authRouter.HandleFunc("/questions/{id}", questionHandler.UpdateQuestion).Methods("PUT")
	authRouter.HandleFunc("/questions/{id}", questionHandler.DeleteQuestion).Methods("DELETE")

	// Answer Routes
	authRouter.HandleFunc("/questions/{question_id}/answers", answerHandler.CreateAnswer).Methods("POST")
	authRouter.HandleFunc("/questions/{question_id}/answers/{answer_id}", answerHandler.UpdateAnswer).Methods("PUT")
	authRouter.HandleFunc("/questions/{question_id}/answers/{answer_id}", answerHandler.DeleteAnswer).Methods("DELETE")

    // NEW: Logout Route
    authRouter.HandleFunc("/logout", authHandler.LogoutUser).Methods("POST")


	// --- Premium Only Routes (Requires Premium User Type) ---
	premiumRouter := router.PathPrefix("/api").Subrouter()
	premiumRouter.Use(func(next http.Handler) http.Handler {
        return middleware.AuthMiddleware(sessionRepo, next)
    })
	premiumRouter.Use(middleware.AuthorizeMiddleware("premium"))

	// Premium Feature: Promote Question
	premiumRouter.HandleFunc("/premium/questions/{id}/promote", questionHandler.PromoteQuestion).Methods("POST")


	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}