package logger_util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/sirupsen/logrus"
)

// asyncWriter реализует io.Writer для асинхронной записи логов
type asyncWriter struct {
	writer io.Writer      // Основной Writer (например, os.Stdout)
	queue  chan []byte    // Канал для буферизации логов (байтовых срезов)
	wg     sync.WaitGroup // WaitGroup для ожидания завершения горутины
	done   chan struct{}  // Канал для сигнала завершения работы
}

// LoggerConfig содержит настройки, специфичные для логгера
type LoggerConfig struct {
	LogLevel string `env:"LOG_LEVEL" env-default:"info"`
}

func NewAsyncWriter(destWriter io.Writer, queueSize int) (*asyncWriter, error) {
	if destWriter == nil {
		return nil, fmt.Errorf("destination writer cannot be nil")
	}

	aw := &asyncWriter{
		writer: destWriter,
		queue:  make(chan []byte, queueSize),
		done:   make(chan struct{}),
	}

	// Запускаем фоновую горутину для обработки очереди и записи
	aw.wg.Add(1)
	go aw.processQueue()

	return aw, nil
}

// Write реализует метод io.Writer
func (aw *asyncWriter) Write(p []byte) (n int, err error) {
	pCopy := make([]byte, len(p))
	copy(pCopy, p)

	select {
	case aw.queue <- pCopy: // Пытаемся отправить данные в очередь
		return len(p), nil
	case <-aw.done: // Если получен сигнал о завершении работы
		return 0, io.ErrClosedPipe // Возвращаем ошибку
	default: // Если очередь заполнена
		return len(p), nil
	}
}

// processQueue - горутина, которая читает данные из очереди и записывает их в основной writer
func (aw *asyncWriter) processQueue() {
	defer aw.wg.Done() // Уменьшаем счетчик горутин при выходе
	defer func() {
		// Закрываем основной writer ТОЛЬКО если он поддерживает io.Closer И НЕ является os.Stdout
		if closer, ok := aw.writer.(io.Closer); ok && aw.writer != os.Stdout {
			closer.Close()
		}
	}()

	for {
		select {
		case data, ok := <-aw.queue:
			if !ok {
				// Канал закрыт (сигнал Close()), обрабатываем все оставшиеся элементы в очереди перед выходом
				for data := range aw.queue {
					aw.writer.Write(data) // Синхронно пишем оставшиеся логи
				}
				return // Завершаем горутину
			}
			// Получили данные из очереди, пишем их в основной writer.
			aw.writer.Write(data)
		case <-aw.done:
			// Получен сигнал завершения работы. Обрабатываем оставшиеся элементы в очереди.
			for {
				select {
				case data := <-aw.queue:
					aw.writer.Write(data)
				default:
					// Очередь пуста
					return // Завершаем горутину после обработки всего из очереди
				}
			}
		}
	}
}

// Close закрывает асинхронный writer.
// Отправляет сигнал завершения горутине и ждет, пока она обработает все оставшиеся логи в очереди и завершится.
func (aw *asyncWriter) Close() error {
	close(aw.done) // Отправляем сигнал завершения горутине
	aw.wg.Wait()   // Ждем завершения горутины (которая в defer закроет writer, если необходимо)
	return nil
}

// SetupLogger настраивает логгер Logrus для асинхронного вывода в терминал.
// Использует cleanenv для загрузки LOG_LEVEL из .env файла или переменных окружения.
func SetupLogger() (*logrus.Logger, func() error) {
	log := logrus.New()

	// Используем текстовый форматтер для терминала для лучшей читаемости.
	textFormatter := &logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true, // Попробуйте включить цвета
	}
	log.SetFormatter(textFormatter)

	// --- Настройка асинхронной записи в os.Stdout ---
	queueSize := 1000 // Размер буфера (очереди)

	// Создаем асинхронный writer, который будет писать в os.Stdout
	asyncWriter, err := NewAsyncWriter(os.Stdout, queueSize)
	if err != nil {
		// Если не удалось настроить асинхронный writer для stdout (маловероятно, но обрабатываем ошибку)
		log.WithError(err).Error("Не удалось настроить асинхронный writer для stdout. Логи будут выводиться синхронно в stdout.")
		log.SetOutput(os.Stdout) // Возвращаемся к синхронному выводу
		// Возвращаем nil функцию закрытия, т.к. асинхронный writer не был успешно создан
		return log, func() error { return nil }
	}

	// Устанавливаем созданный асинхронный writer как вывод для логгера Logrus.
	log.SetOutput(asyncWriter)

	// --- Загрузка уровня логирования с помощью cleanenv ---
	var loggerCfg LoggerConfig
	// cleanenv.ReadConfig попытается прочитать из .env и переопределить из окружения.
	err = cleanenv.ReadConfig(".env", &loggerCfg)

	// Проверяем ошибку загрузки
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn("Файл .env не найден для настроек логгера. LOG_LEVEL должен быть загружен из переменных окружения.")
		} else {
			log.WithError(err).Error("Ошибка загрузки LOG_LEVEL с помощью cleanenv. Используется уровень 'info'")
		}
	} else {
		log.Debugf("LOG_LEVEL загружен: %s", loggerCfg.LogLevel)
	}

	// Настройка уровня логирования из загруженного значения
	level, err := logrus.ParseLevel(loggerCfg.LogLevel)
	if err != nil {
		// Это сработает, если значение LOG_LEVEL из .env или окружения некорректно
		log.Warnf("Некорректное значение LOG_LEVEL '%s' загружено, используется уровень 'info'", loggerCfg.LogLevel)
		level = logrus.InfoLevel
	}
	log.SetLevel(level)
	// --- Конец загрузки уровня логирования ---

	closeFunc := func() error {
		return asyncWriter.Close()
	}

	return log, closeFunc
}

// logrusGormWriter адаптирует логгер Logrus к интерфейсу логгера GORM.
// Это часть исходного кода и не требует изменений.
type LogrusGormWriter struct {
	Logger *logrus.Logger
}

// Printf реализует интерфейс логгера GORM.
// Использует Tracef для логирования запросов GORM.
func (w *LogrusGormWriter) Printf(message string, data ...interface{}) {
	w.Logger.Tracef(message, data...)
}
