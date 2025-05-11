package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IlyushinDM/user-order-api/internal/core"
	"github.com/IlyushinDM/user-order-api/internal/utils/config_util"
	"github.com/IlyushinDM/user-order-api/internal/utils/logger_util"

	_ "github.com/IlyushinDM/user-order-api/docs"
)

// @title User Order API
// @version 1.0
// @description Сервер для управления пользователями и их заказами.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Напиши "Bearer", пробел и JWT токен. Пример: "Bearer {token}"
func main() {
	// 1. Инициализация логгера.
	logger, cleanupLogger := logger_util.SetupLogger()

	defer func() {
		if err := cleanupLogger(); err != nil {
			log.Printf("Ошибка при закрытии логгера: %v\n", err)
		}
	}()

	// 2. Загрузка конфигурации.
	cfg, err := config_util.LoadConfig(logger)
	if err != nil {
		logger.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// 3. Инициализация приложения с передачей логгера и конфигурации.
	app, err := core.NewApp(logger, cfg)
	if err != nil {
		logger.Fatalf("Ошибка инициализации приложения: %v", err)
	}

	// 4. Запуск и управление жизненным циклом сервера.
	if err := runApp(app); err != nil {
		logger.Fatalf("Ошибка во время выполнения приложения: %v", err)
	}

	app.Logger.Info("Приложение завершило работу.")
}

// runApp настраивает маршрутизатор, запускает HTTP сервер и обрабатывает graceful shutdown
func runApp(app *core.App) error {
	// 1. Установка маршрутизатора.
	app.Router = core.SetupRouter(app)

	// 2. Создание экземпляра http.Server с таймаутами из конфигурации.
	srv := &http.Server{
		Addr:    ":" + app.Config.Port,
		Handler: app.Router,
		// Используем сконфигурированные таймауты сервера
		ReadTimeout:    time.Duration(app.Config.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(app.Config.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(app.Config.IdleTimeout) * time.Second,
		MaxHeaderBytes: app.Config.MaxHeaderBytes,
	}

	// 3. Запуск сервера в отдельной горутине.
	serverErr := make(chan error, 1)
	go func() {
		app.Logger.Infof("Сервер запускается на порту %s в режиме %s...", app.Config.Port, app.Config.GinMode)
		app.Logger.Infof("API доступен по адресу: http://localhost:%s", app.Config.Port)
		app.Logger.Infof("Ссылка для перехода в документацию: http://localhost:%s/swagger/index.html", app.Config.Port)

		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			app.Logger.Errorf("Ошибка запуска сервера: %v", err)
			serverErr <- fmt.Errorf("ошибка запуска сервера: %w", err)
		} else if err == http.ErrServerClosed {
			app.Logger.Info("HTTP сервер успешно остановлен после graceful shutdown.")
			serverErr <- nil
		} else {
			serverErr <- nil
		}
		close(serverErr)
	}()

	// 4. Ожидание сигналов операционной системы для graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 5. Ожидание сигнала остановки или критической ошибки сервера.
	select {
	case err := <-serverErr:
		app.Logger.Errorf("Сервер завершил работу с ошибкой до получения сигнала остановки: %v", err)
		return fmt.Errorf("сервер завершился с ошибкой: %w", err)
	case <-quit:
		app.Logger.Info("Получен сигнал остановки, запускается graceful shutdown...")

		// Создаем контекст с таймаутом, используя сконфигурированное значение.
		ctx, cancel := context.WithTimeout(context.Background(), app.Config.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			app.Logger.Errorf("Ошибка при graceful shutdown сервера: %v", err)
			return fmt.Errorf("ошибка при graceful shutdown: %w", err)
		}

		app.Logger.Info("Graceful shutdown завершен. Ожидание завершения горутины сервера...")
		<-serverErr
		app.Logger.Info("Горутина сервера завершилась.")
	}

	return nil
}
