package main

import (
	"Calculator3.0/Internal/DataBase"
	"Calculator3.0/Internal/Middleware"
	. "Calculator3.0/Internal/Orchestrator"
	"Calculator3.0/Internal/User"
	"fmt"
	"net/http"
	"os"
)

func main() {
	DataBase.InitDB()
	go StartGRPCServer()
	go CreateTasks()

	http.HandleFunc("/api/v1/calculate", Middleware.AuthMiddleware(AddExpressionHandler))
	http.HandleFunc("/api/v1/expressions", Middleware.AuthMiddleware(GetExpressionsHandler))
	http.HandleFunc("/api/v1/expressions/", Middleware.AuthMiddleware(GetExpressionByIDHandler))
	http.HandleFunc("/api/v1/register", User.RegisterHandler)
	http.HandleFunc("/api/v1/login", User.LoginHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Оркестратор запущен на порту", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("HTTP server failed: %v\n", err)
	}
}
