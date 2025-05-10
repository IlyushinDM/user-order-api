package logger_util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv" // Добавляем импорт cleanenv
	"github.com/sirupsen/logrus"
)

// asyncWriter реализует io.Writer для асинхронной записи логов.
// Использует канал для буферизации записей и горутину для фоновой записи
// в заданный io.Writer.
type asyncWriter struct {
	writer io.Writer      // Основной Writer (например, os.Stdout)
	queue  chan []byte    // Канал для буферизации логов (байтовых срезов)
	wg     sync.WaitGroup // WaitGroup для ожидания завершения горутины
	done   chan struct{}  // Канал для сигнала завершения работы
}

// LoggerConfig содержит настройки, специфичные для логгера.
// Значения будут загружены cleanenv.
type LoggerConfig struct {
	LogLevel string `env:"LOG_LEVEL" env-default:"info"` // Описываем переменную LOG_LEVEL для cleanenv
}

// NewAsyncWriter создает и запускает новый асинхронный Writer для заданного io.Writer.
// destWriter - io.Writer, куда будут записываться логи (например, os.Stdout).
// queueSize - размер буфера (канала) для логов.
func NewAsyncWriter(destWriter io.Writer, queueSize int) (*asyncWriter, error) {
	// Простая проверка на nil writer
	if destWriter == nil {
		return nil, fmt.Errorf("destination writer cannot be nil")
	}

	aw := &asyncWriter{
		writer: destWriter,
		queue:  make(chan []byte, queueSize),
		done:   make(chan struct{}),
	}

	// Запускаем фоновую горутину для обработки очереди и записи.
	aw.wg.Add(1)
	go aw.processQueue()

	return aw, nil
}

// Write реализует метод io.Writer.
// Отправляет данные в очередь. Не блокируется, если очередь заполнена (теряет лог).
func (aw *asyncWriter) Write(p []byte) (n int, err error) {
	pCopy := make([]byte, len(p))
	copy(pCopy, p)

	select {
	case aw.queue <- pCopy: // Пытаемся отправить данные в очередь
		// Успешно добавлено в очередь, возвращаем количество байт.
		return len(p), nil
	case <-aw.done: // Если получен сигнал о завершении работы
		return 0, io.ErrClosedPipe // Возвращаем ошибку
	default: // Если очередь заполнена
		// При полной очереди мы предпочитаем пропустить лог.
		// logrus.Errorf("Log queue is full, dropping log entry") // Пример логирования потери (может вызвать рекурсию!)
		return len(p), nil // Возвращаем len(p), чтобы вызывающий считал запись "успешной" по количеству байт.
	}
}

// processQueue - горутина, которая читает данные из очереди и записывает их в основной writer.
func (aw *asyncWriter) processQueue() {
	defer aw.wg.Done() // Уменьшаем счетчик горутин при выходе
	defer func() {
		// Закрываем основной writer ТОЛЬКО если он поддерживает io.Closer И НЕ является os.Stdout.
		// os.Stdout не следует закрывать.
		if closer, ok := aw.writer.(io.Closer); ok && aw.writer != os.Stdout {
			closer.Close()
		}
	}()

	for {
		select {
		case data, ok := <-aw.queue:
			if !ok {
				// Канал закрыт (сигнал Close()), обрабатываем все оставшиеся элементы в очереди перед выходом.
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
// Возвращает:
//   - *logrus.Logger: настроенный экземпляр логгера.
//   - func() error: функция для корректного завершения работы асинхронного writer'а. Эту функцию необходимо вызвать при завершении работы приложения (например, в main перед os.Exit).
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
		// Если ошибка - файл .env не найден, это не критично.
		// cleanenv должен был попытаться загрузить из переменных окружения в первом вызове.
		if errors.Is(err, os.ErrNotExist) {
			log.Warn("Файл .env не найден для настроек логгера. LOG_LEVEL должен быть загружен из переменных окружения.")
			// Нет необходимости явно вызывать ReadEnv здесь, т.к. ReadConfig уже это делает.
			// Если обязательные поля были бы в LoggerConfig, cleanenv.ReadConfig
			// вернул бы ошибку, если их нет ни в .env, ни в env.
		} else {
			// Если это любая другая ошибка cleanenv (не связанная с отсутствием файла),
			// это указывает на проблему с загрузкой или валидацией из env.
			log.WithError(err).Error("Ошибка загрузки LOG_LEVEL с помощью cleanenv. Используется уровень 'info'")
			// loggerCfg останется со значением по умолчанию ("info") благодаря `env-default`.
		}
	} else {
		// Если cleanenv.ReadConfig завершился без ошибки (нашел .env или загрузил из env)
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

	// Возвращаем логгер и функцию для его корректного закрытия.
	// Вызывающая сторона ДОЛЖНА вызвать closeFunc перед завершением работы,
	// чтобы гарантировать обработку всех оставшихся в буфере логов.
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
