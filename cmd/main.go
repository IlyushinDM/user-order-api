package main

import (
	"fmt"
	"net/http" // Пример добавления зависимости из стандартной библиотеки

	"github.com/IlyushinDM/user-order-api/internal/auth" // Импорт внутреннего пакета
)

func main() {
	// Пример использования внутреннего пакета
	user := auth.GetUser()
	fmt.Printf("Default user: %s\n", user)

	// Простой HTTP-сервер
	http.HandleFunc("/", helloHandler)
	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, Go Professional!")
}
