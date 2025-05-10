# Функция для настройки базы данных: проверка существования и создание при необходимости
setup_database() {
    echo ">> Проверка и настройка базы данных '$DB_NAME' на $DB_HOST:$DB_PORT..."

    # Формируем базовую команду psql в виде массива (без PGPASSWORD в массиве)
    local psql_base_cmd=(psql -U "$DB_USER" -h "$DB_HOST" -p "$DB_PORT")

    # Переменная PGPASSWORD должна быть установлена непосредственно перед вызовом команды psql
    # для того, чтобы она была видна только этой команде и не сохранялась в окружении
    # Используем двойные кавычки вокруг значения $DB_PASSWORD на случай спецсимволов

    echo "   Использование PGPASSWORD для аутентификации." # Сообщение для ясности

    # Проверяем, существует ли база данных с таким именем
    # Используем -lqt для списка баз, -c для выполнения команды и -t для вывода только строк
    # Грепаем точное имя базы данных, учитывая возможные пробелы вокруг имени в выводе psql
    # PGPASSWORD="$DB_PASSWORD" ставится перед вызовом команды
    if PGPASSWORD="$DB_PASSWORD" "${psql_base_cmd[@]}" -lqt | cut -d \| -f 1 | grep -q "^ *$DB_NAME *$"; then
        echo "<< База данных '$DB_NAME' уже существует."
    else
        echo ">> База данных '$DB_NAME' не найдена. Выполняется создание..."
        # Создаем базу данных, подключаясь к стандартной БД 'postgres'
        # Экранируем имя базы данных в SQL команде двойными кавычками
        echo "   Выполнение: psql ... -d postgres -c \"CREATE DATABASE \"$DB_NAME\"\" (используя PGPASSWORD)"
        # PGPASSWORD="$DB_PASSWORD" ставится перед вызовом команды создания
        if PGPASSWORD="$DB_PASSWORD" "${psql_base_cmd[@]}" -d postgres -c "CREATE DATABASE \"$DB_NAME\""; then
            echo "<< База данных '$DB_NAME' успешно создана."
        else
            echo "!! Ошибка при создании базы данных '$DB_NAME'." >&2
            exit 1
        fi
    fi
}

# Остальные функции остаются без изменений

# Функция для загрузки переменных из .env файла
load_environment_variables() {
    local env_file=".env"
    echo ">> Попытка загрузки переменных из файла: $env_file"

    if [ -f "$env_file" ]; then
        # Читаем файл построчно
        while IFS='=' read -r line; do
            # Удаляем ведущие и завершающие пробелы из строки перед обработкой
            local trimmed_line=$(printf "%s" "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

            # Игнорируем комментарии (строки, начинающиеся с # после обрезки пробелов) и пустые строки
            if [[ "$trimmed_line" =~ ^#.* ]] || [[ -z "$trimmed_line" ]]; then
                continue
            fi

            # Находим позицию первого символа '=' для разделения имени и значения
            # Если строка не содержит '=', игнорируем ее (или выдаем предупреждение)
            if [[ "$trimmed_line" =~ ^[^=]*= ]]; then
                # Разделяем строку на имя переменной (до первого =) и значение (после первого =)
                local var_name="${trimmed_line%%=*}"
                local var_value="${trimmed_line#*=}"

                # Удаляем ведущие и завершающие пробелы из имени переменной
                var_name=$(printf "%s" "$var_name" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

                # Удаляем ведущие и завершающие пробелы из значения переменной
                local value_trimmed_spaces=$(printf "%s" "$var_value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')

                # Проверяем и удаляем ведущие/завершающие двойные или одинарные кавычки из значения, если они есть
                local final_value="$value_trimmed_spaces"

                # Если значение начинается и заканчивается на двойные кавычки
                if [[ "$final_value" =~ ^\".*\"$ ]]; then
                    # Удаляем только внешние кавычки
                    final_value="${final_value#\"}"
                    final_value="${final_value%\"}"
                # Если значение начинается и заканчивается на одинарные кавычки
                elif [[ "$final_value" =~ ^\'.*\'$ ]]; then
                    # Удаляем только внешние кавычки
                    final_value="${final_value#\'}"
                    final_value="${final_value%\'}"
                fi

                # Экспортируем переменную в текущее окружение скрипта
                export "$var_name=$final_value"

            else
                # Строка без знака '=' (после фильтрации комментариев и пустых строк)
                echo "!! Предупреждение: В файле .env пропущена строка без знака '=': $trimmed_line" >&2
            fi

        done < "$env_file"
        echo "<< Переменные из .env загружены."
    else
        echo "!! Ошибка: Файл .env не найден в текущей директории '$PWD'." >&2
        exit 1
    fi
}

# Функция для проверки наличия обязательных переменных окружения
validate_essential_variables() {
    echo ">> Проверка наличия обязательных переменных окружения..."

    local required_vars=("DB_USER" "DB_PASSWORD" "DB_HOST" "DB_PORT" "DB_NAME")
    local missing_vars=()

    for var_name in "${required_vars[@]}"; do
        if [ -z "${!var_name:-}" ]; then
            missing_vars+=("$var_name")
        fi
    done

    if [ ${#missing_vars[@]} -ne 0 ]; then
        echo "!! Ошибка: Следующие обязательные переменные не установлены или пусты: ${missing_vars[*]}" >&2
        echo "!! Проверка переменных завершена с ошибками. Исправьте файл .env или установите переменные окружения." >&2
        exit 1
    fi

    # Установка значений по умолчанию для необязательных переменных, если они не установлены или пусты
    local db_sslmode_default="disable"
    if [ -z "${DB_SSLMODE:-}" ]; then
        DB_SSLMODE="$db_sslmode_default"
        export DB_SSLMODE
        echo "   Переменная DB_SSLMODE не задана или пуста, установлено значение по умолчанию: $DB_SSLMODE"
    fi

    local migrations_path_default="./migrations"
    if [ -z "${MIGRATIONS_PATH:-}" ]; then
         MIGRATIONS_PATH="$migrations_path_default"
         export MIGRATIONS_PATH
         echo "   Переменная MIGRATIONS_PATH не задана или пуста, установлен путь по умолчанию: $MIGRATIONS_PATH"
    fi


    echo "<< Обязательные переменные найдены и не пусты. Необязательные переменные установлены (если не были заданы):"
    # Выводим значения, маскируя конфиденциальные данные
    echo "   DB_USER: ${DB_USER:0:3}..."
    echo "   DB_PASSWORD: ${DB_PASSWORD:0:3}..."
    echo "   DB_HOST: $DB_HOST"
    echo "   DB_PORT: $DB_PORT"
    echo "   DB_NAME: $DB_NAME"
    echo "   DB_SSLMODE: $DB_SSLMODE"
    echo "   MIGRATIONS_PATH: $MIGRATIONS_PATH"
}


# Функция для установки или обновления утилиты миграции migrate
install_migration_tool() {
    echo ">> Проверка и установка/обновление CLI утилиты migrate..."

    # Проверяем, установлен ли Go
    if ! command -v go &> /dev/null; then
        echo "!! Ошибка: Команда 'go' не найдена. Убедитесь, что Go установлен и доступен в PATH." >&2
        exit 1
    fi

    # Устанавливаем или обновляем утилиту migrate с поддержкой postgres
    echo "   Выполнение: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    if go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; then
        echo "<< Утилита migrate успешно установлена/обновлена."
    else
        echo "!! Ошибка при установке/обновлении утилиты migrate." >&2
        echo "   Убедитесь в наличии интернет-соединения и корректной настройке Go proxy." >&2
        exit 1
    fi

    # Добавляем каталог бинарников Go в PATH для текущей сессии, если он еще не там
    # Это необходимо, чтобы команда migrate была доступна сразу после установки.
    local go_bin_path
    go_bin_path="$(go env GOPATH)/bin"
    # Проверяем, существует ли каталог GOPATH/bin и не находится ли он уже в PATH
    if [ -d "$go_bin_path" ] && [[ ":$PATH:" != *":$go_bin_path:"* ]]; then
        export PATH="$PATH:$go_bin_path"
        echo ">> Добавлен Go бинарный путь ($go_bin_path) в PATH для текущей сессии."
    elif [ ! -d "$go_bin_path" ]; then
        echo "!! Предупреждение: Каталог бинарников Go '$go_bin_path' не найден. migrate может быть недоступен." >&2
         echo "   Убедитесь, что GOPATH правильно установлен или Go бинарники компилируются в другой каталог." >&2
    fi
}

# Функция для выполнения миграций базы данных
apply_database_migrations() {
    echo ">> Применение миграций базы данных из каталога: $MIGRATIONS_PATH"

    # Проверяем, доступна ли команда migrate после установки и добавления в PATH
    if ! command -v migrate &> /dev/null; then
        echo "!! Ошибка: Команда 'migrate' не найдена." >&2
        echo "   Убедитесь, что она успешно установлена и $(go env GOPATH)/bin или $HOME/go/bin находится в вашей переменной окружения PATH." >&2
        exit 1
    fi

    # Формируем URL для подключения к базе данных
    local database_url="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
    # Выводим URL, скрывая пароль для безопасности в логах
    local display_url="postgres://${DB_USER}:******@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
    echo ">> Используемый URL базы данных (пароль скрыт): $display_url"

    # Проверяем существование каталога с миграциями
    if [ ! -d "$MIGRATIONS_PATH" ]; then
        echo "!! Ошибка: Каталог с миграциями '$MIGRATIONS_PATH' не найден." >&2
        echo "   Убедитесь, что переменная MIGRATIONS_PATH указана верно и каталог существует." >&2
        exit 1
    fi

    # Выполняем миграции
    echo "   Выполнение: migrate -path \"$MIGRATIONS_PATH\" -database \"<masked_url>\" up"
    # Передаем URL в двойных кавычках, чтобы избежать проблем с пробелами или спецсимволами (кроме тех, что в пароле и требуют URL-кодирования)
    if migrate -path "$MIGRATIONS_PATH" -database "$database_url" up; then
        echo "<< Миграции успешно применены."
    else
        echo "!! Ошибка при выполнении миграций." >&2
        echo "   Проверьте логи команды migrate выше для получения подробностей." >&2
        exit 1
    fi
}

# Функция для загрузки Go модулей
download_go_modules() {
    echo ">> Загрузка Go модулей..."
     # Проверяем, установлен ли Go
    if ! command -v go &> /dev/null; then
        echo "!! Ошибка: Команда 'go' не найдена. Убедитесь, что Go установлен и доступен в PATH." >&2
        exit 1
    fi

    # Проверяем наличие файла go.mod, который является маркером Go модуля
    if [ ! -f go.mod ]; then
        echo "!! Ошибка: Файл go.mod не найден в текущей директории '$PWD'." >&2
        echo "   Убедитесь, что вы запускаете скрипт из корневой директории Go проекта." >&2
        exit 1
    fi

    echo "   Выполнение: go mod download"
    if go mod download; then
        echo "<< Go модули успешно загружены."
    else
        echo "!! Ошибка при загрузке Go модулей." >&2
         echo "   Проверьте логи go mod download выше для получения подробностей." >&2
        exit 1
    fi
}

# Функция для сборки Go приложения
build_go_application() {
    echo ">> Сборка Go приложения..."
     # Проверяем, установлен ли Go
    if ! command -v go &> /dev/null; then
        echo "!! Ошибка: Команда 'go' не найдена. Убедитесь, что Go установлен и доступен в PATH." >&2
        exit 1
    fi

    local source_file="cmd/main.go"
    # Проверяем наличие основного исходного файла приложения
     if [ ! -f "$source_file" ]; then
        echo "!! Ошибка: Основной исходный файл '$source_file' не найден." >&2
        echo "   Убедитесь, что вы запускаете скрипт из корневой директории Go проекта и путь к main.go верен." >&2
        exit 1
    fi

    local output_binary_name="run_api"
    local output_binary_path="./$output_binary_name"
    # Определяем расширение для исполняемого файла в зависимости от ОС
    if [[ "$OSTYPE" == "linux-gnu"* || "$OSTYPE" == "darwin"* ]]; then
        # Linux или macOS без расширения
        output_binary_path="./$output_binary_name"
    elif [[ "$OSTYPE" == "cygwin" || "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
        # Windows с расширением .exe
        output_binary_path="./$output_binary_name.exe"
    else
        echo "!! Предупреждение: Неизвестная операционная система '$OSTYPE'. Бинарник будет назван без расширения." >&2
         output_binary_path="./$output_binary_name"
    fi

    echo "   Выполнение: go build -o \"$output_binary_path\" \"$source_file\""
    # Выполняем сборку
    if go build -o "$output_binary_path" "$source_file"; then
        echo "<< Сборка завершена успешно. Исполняемый файл: $output_binary_path"
    else
        echo "!! Ошибка при сборке Go приложения." >&2
        echo "   Проверьте логи go build выше для получения подробностей." >&2
        exit 1
    fi
}

# --- Основная логика скрипта ---

echo "--- Запуск скрипта подготовки среды и сборки ---"

# Включаем режим немедленного выхода при ошибке и при использовании неустановленных переменных
set -e
set -u

# Вызываем функции в нужной последовательности
load_environment_variables
validate_essential_variables
setup_database # <-- Исправлена эта функция
install_migration_tool
apply_database_migrations
download_go_modules
build_go_application

echo "--- Все этапы подготовки и сборки завершены успешно ---"

exit 0

