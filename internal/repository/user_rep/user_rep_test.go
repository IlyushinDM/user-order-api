package user_rep_test

// import (
// 	"context"
// 	"testing"

// 	// Библиотека для мокинга SQL драйвера (альтернативный подход)
// 	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
// 	"github.com/IlyushinDM/user-order-api/internal/repository/user_rep" // Убедитесь, что путь импорта правильный
// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// 	"gorm.io/driver/sqlite" // Используем драйвер SQLite для in-memory тестов
// 	"gorm.io/gorm"
// 	"gorm.io/gorm/logger"
// )

// var (
// 	// db *gorm.DB // Можно использовать глобальную переменную для DB, если тесты не изолированы транзакциями
// 	mockLog *logrus.Logger // Логгер для тестов
// )

// // setupTestDB инициализирует in-memory SQLite базу данных для тестов.
// // Возвращает *gorm.DB и функцию очистки (cleanup).
// func setupTestDB(t *testing.T) (*gorm.DB, func()) {
// 	// Инициализируем логгер для тестов, чтобы не засорять консоль обычными логами GORM/приложения
// 	if mockLog == nil {
// 		mockLog = logrus.New()
// 		mockLog.SetOutput(nil)              // Отключаем вывод логов в консоль для тестов
// 		mockLog.SetLevel(logrus.PanicLevel) // Устанавливаем высокий уровень, чтобы ничего не логировалось
// 	}

// 	// Открываем in-memory SQLite базу данных
// 	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
// 		Logger: logger.Default.LogMode(logger.Silent), // Отключаем логирование GORM
// 	})
// 	if err != nil {
// 		t.Fatalf("Не удалось открыть in-memory SQLite базу данных: %v", err)
// 	}

// 	// Автомиграция моделей
// 	err = db.AutoMigrate(&user_model.User{})
// 	if err != nil {
// 		t.Fatalf("Не удалось выполнить автомиграцию: %v", err)
// 	}

// 	// Возвращаем подключение к БД и функцию очистки.
// 	// В данном случае, закрытие DB connection приведет к удалению in-memory базы.
// 	sqlDB, _ := db.DB()
// 	return db, func() {
// 		sqlDB.Close()
// 	}
// }

// // setupTestRepository настраивает тестовое окружение для репозитория.
// // Возвращает экземпляр UserRepository и функцию очистки.
// func setupTestRepository(t *testing.T) (user_rep.UserRepository, func()) {
// 	db, cleanup := setupTestDB(t)
// 	repo := user_rep.NewGormUserRepository(db, mockLog)
// 	return repo, cleanup
// }

// // TestCreateUser проверяет успешное создание пользователя.
// func TestCreateUser(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	user := &user_model.User{
// 		Name:  "Test User",
// 		Email: "test.user@example.com",
// 		Age:   30,
// 	}

// 	err := repo.Create(context.Background(), user)
// 	assert.NoError(t, err, "Create должен успешно создать пользователя")
// 	assert.NotZero(t, user.ID, "После создания ID пользователя должен быть установлен")

// 	// Проверяем, что пользователь действительно был сохранен
// 	fetchedUser, err := repo.GetByID(context.Background(), user.ID)
// 	assert.NoError(t, err, "GetByID должен найти созданного пользователя")
// 	assert.NotNil(t, fetchedUser, "Найденный пользователь не должен быть nil")
// 	assert.Equal(t, user.Email, fetchedUser.Email, "Email созданного и найденного пользователя должны совпадать")
// }

// // TestCreateUser_DuplicateEmail (пример, требует UNIQUE ограничения на Email в модели user_model.User)
// // func TestCreateUser_DuplicateEmail(t *testing.T) {
// // 	repo, cleanup := setupTestRepository(t)
// // 	defer cleanup()

// // 	user1 := &user_model.User{Name: "User 1", Email: "duplicate@example.com", Age: 25}
// // 	user2 := &user_model.User{Name: "User 2", Email: "duplicate@example.com", Age: 35}

// // 	err := repo.Create(context.Background(), user1)
// // 	assert.NoError(t, err, "Первый пользователь должен быть создан успешно")

// // 	err = repo.Create(context.Background(), user2)
// // 	// Ожидаем ошибку из-за нарушения UNIQUE ограничения
// // 	assert.Error(t, err, "Создание пользователя с дублирующимся email должно вернуть ошибку")
// // 	// Можно добавить проверку на конкретный тип ошибки или её текст, если GORM возвращает специфичную ошибку для UNIQUE
// // }

// // TestUpdateUser проверяет успешное обновление пользователя.
// func TestUpdateUser(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Сначала создаем пользователя
// 	user := &user_model.User{
// 		Name:  "Initial Name",
// 		Email: "initial@example.com",
// 		Age:   20,
// 	}
// 	err := repo.Create(context.Background(), user)
// 	assert.NoError(t, err, "Не удалось создать пользователя для обновления")
// 	initialID := user.ID

// 	// Обновляем поля пользователя
// 	user.Name = "Updated Name"
// 	user.Age = 21
// 	// Email нельзя менять просто так, если он UNIQUE, но в данном тесте мы его не трогаем.
// 	// user.Email = "updated@example.com" // Опасно менять UNIQUE поля в тестах без учета ограничений

// 	err = repo.Update(context.Background(), user)
// 	assert.NoError(t, err, "Update должен успешно обновить пользователя")

// 	// Проверяем, что пользователь был обновлен в БД
// 	fetchedUser, err := repo.GetByID(context.Background(), initialID)
// 	assert.NoError(t, err, "Не удалось найти обновленного пользователя по ID")
// 	assert.NotNil(t, fetchedUser, "Обновленный пользователь не должен быть nil")
// 	assert.Equal(t, "Updated Name", fetchedUser.Name, "Имя пользователя должно быть обновлено")
// 	assert.Equal(t, 21, fetchedUser.Age, "Возраст пользователя должен быть обновлен")
// 	assert.Equal(t, "initial@example.com", fetchedUser.Email, "Email пользователя не должен был измениться") // Проверяем, что Email остался прежним
// 	assert.Equal(t, initialID, fetchedUser.ID, "ID пользователя не должен был измениться")
// }

// // TestUpdateUser_NotFound проверяет обновление несуществующего пользователя.
// func TestUpdateUser_NotFound(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Попытка обновить пользователя с ID=999, которого нет в БД
// 	nonExistentUser := &user_model.User{
// 		ID:    999,
// 		Name:  "Non Existent",
// 		Email: "none@example.com",
// 		Age:   50,
// 	}

// 	err := repo.Update(context.Background(), nonExistentUser)
// 	// GORM Updates с .Model(user) не всегда возвращает ErrRecordNotFound,
// 	// если запись не найдена, так как RowsAffected будет 0.
// 	// В текущей реализации репозитория Update не возвращает ошибку при RowsAffected == 0,
// 	// только логирует предупреждение.
// 	// assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "Update несуществующего пользователя должен вернуть ErrRecordNotFound")
// 	assert.NoError(t, err, "Update несуществующего пользователя не должен возвращать ошибку в данной реализации, но RowsAffected будет 0")

// 	// Проверяем, что в логах есть предупреждение о 0 затронутых строках
// 	// Это требует мокинга логгера и проверки его вызовов, что выходит за рамки простого примера.
// 	// Альтернатива: изменить логику репозитория, чтобы он возвращал ошибку при RowsAffected == 0.
// }

// // TestDeleteUser проверяет успешное удаление пользователя (мягкое удаление).
// func TestDeleteUser(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем пользователя для удаления
// 	user := &user_model.User{
// 		Name:  "User to Delete",
// 		Email: "delete@example.com",
// 		Age:   40,
// 	}
// 	err := repo.Create(context.Background(), user)
// 	assert.NoError(t, err, "Не удалось создать пользователя для удаления")
// 	userID := user.ID

// 	// Удаляем пользователя
// 	err = repo.Delete(context.Background(), userID)
// 	assert.NoError(t, err, "Delete должен успешно удалить пользователя")

// 	// Проверяем, что пользователь помечен как удаленный (мягкое удаление)
// 	// GetByID по умолчанию не находит мягко удаленные записи.
// 	fetchedUser, err := repo.GetByID(context.Background(), userID)
// 	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "GetByID не должен найти мягко удаленного пользователя")
// 	assert.Nil(t, fetchedUser, "Мягко удаленный пользователь должен быть nil при поиске по ID")

// 	// Можно проверить наличие записи с непустым DeletedAt напрямую через Unscoped
// 	// var deletedUser user_model.User
// 	// result := repo.(*user_rep.GormUserRepository).db.Unscoped().First(&deletedUser, userID)
// 	// assert.NoError(t, result.Error, "Unscoped должен найти мягко удаленного пользователя")
// 	// assert.NotNil(t, deletedUser.DeletedAt, "Поле DeletedAt должно быть установлено для мягко удаленного пользователя")
// }

// // TestDeleteUser_NotFound проверяет удаление несуществующего пользователя.
// func TestDeleteUser_NotFound(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Попытка удалить пользователя с ID=999, которого нет в БД
// 	err := repo.Delete(context.Background(), 999)
// 	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "Delete несуществующего пользователя должен вернуть ErrRecordNotFound")
// }

// // TestGetByID проверяет получение пользователя по ID.
// func TestGetByID(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем пользователя для получения
// 	user := &user_model.User{
// 		Name:  "Find Me",
// 		Email: "findme@example.com",
// 		Age:   28,
// 	}
// 	err := repo.Create(context.Background(), user)
// 	assert.NoError(t, err, "Не удалось создать пользователя для получения по ID")
// 	userID := user.ID

// 	// Получаем пользователя по ID
// 	fetchedUser, err := repo.GetByID(context.Background(), userID)
// 	assert.NoError(t, err, "GetByID должен успешно найти пользователя")
// 	assert.NotNil(t, fetchedUser, "Найденный пользователь не должен быть nil")
// 	assert.Equal(t, userID, fetchedUser.ID, "Полученный пользователь должен иметь правильный ID")
// 	assert.Equal(t, "Find Me", fetchedUser.Name, "Полученный пользователь должен иметь правильное имя")
// 	assert.Equal(t, "findme@example.com", fetchedUser.Email, "Полученный пользователь должен иметь правильный Email")
// 	assert.Equal(t, 28, fetchedUser.Age, "Полученный пользователь должен иметь правильный возраст")
// }

// // TestGetByID_NotFound проверяет получение несуществующего пользователя по ID.
// func TestGetByID_NotFound(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Попытка получить пользователя с ID=999, которого нет в БД
// 	fetchedUser, err := repo.GetByID(context.Background(), 999)
// 	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "GetByID несуществующего пользователя должен вернуть ErrRecordNotFound")
// 	assert.Nil(t, fetchedUser, "Найденный пользователь должен быть nil")
// }

// // TestGetByEmail проверяет получение пользователя по Email.
// func TestGetByEmail(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем пользователя для получения по Email
// 	user := &user_model.User{
// 		Name:  "Find By Email",
// 		Email: "email.search@example.com",
// 		Age:   35,
// 	}
// 	err := repo.Create(context.Background(), user)
// 	assert.NoError(t, err, "Не удалось создать пользователя для получения по Email")
// 	userEmail := user.Email

// 	// Получаем пользователя по Email
// 	fetchedUser, err := repo.GetByEmail(context.Background(), userEmail)
// 	assert.NoError(t, err, "GetByEmail должен успешно найти пользователя")
// 	assert.NotNil(t, fetchedUser, "Найденный пользователь не должен быть nil")
// 	assert.Equal(t, userEmail, fetchedUser.Email, "Полученный пользователь должен иметь правильный Email")
// 	assert.Equal(t, "Find By Email", fetchedUser.Name, "Полученный пользователь должен иметь правильное имя")
// }

// // TestGetByEmail_NotFound проверяет получение несуществующего пользователя по Email.
// func TestGetByEmail_NotFound(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Попытка получить пользователя по Email, которого нет в БД
// 	fetchedUser, err := repo.GetByEmail(context.Background(), "nonexistent@example.com")
// 	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "GetByEmail несуществующего пользователя должен вернуть ErrRecordNotFound")
// 	assert.Nil(t, fetchedUser, "Найденный пользователь должен быть nil")
// }

// // TestGetAll проверяет получение всех пользователей с пагинацией и фильтрами.
// func TestGetAll(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем несколько тестовых пользователей
// 	users := []*user_model.User{
// 		{Name: "Alice", Email: "alice@example.com", Age: 25},
// 		{Name: "Bob", Email: "bob@example.com", Age: 30},
// 		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
// 		{Name: "David", Email: "david@example.com", Age: 40},
// 		{Name: "Eve", Email: "eve@example.com", Age: 28},
// 	}
// 	for _, u := range users {
// 		err := repo.Create(context.Background(), u)
// 		assert.NoError(t, err, "Не удалось создать пользователя для GetAll")
// 	}

// 	// Тест 1: Получение всех без пагинации и фильтров (page=1, limit=100)
// 	fetchedUsers, total, err := repo.GetAll(context.Background(), 1, 100, nil)
// 	assert.NoError(t, err, "GetAll без фильтров и пагинации должен успешно работать")
// 	assert.Len(t, fetchedUsers, 5, "Должно быть получено 5 пользователей")
// 	assert.Equal(t, int64(5), total, "Общее количество пользователей должно быть 5")

// 	// Тест 2: Пагинация - Страница 1, лимит 2
// 	fetchedUsers, total, err = repo.GetAll(context.Background(), 1, 2, nil)
// 	assert.NoError(t, err, "GetAll с пагинацией (стр 1, лимит 2) должен успешно работать")
// 	assert.Len(t, fetchedUsers, 2, "Должно быть получено 2 пользователя на первой странице")
// 	assert.Equal(t, int64(5), total, "Общее количество пользователей должно быть 5")
// 	// Проверяем, что получены правильные пользователи (зависит от порядка сортировки GORM, по умолчанию по ID)
// 	assert.Equal(t, "Alice", fetchedUsers[0].Name)
// 	assert.Equal(t, "Bob", fetchedUsers[1].Name)

// 	// Тест 3: Пагинация - Страница 2, лимит 2
// 	fetchedUsers, total, err = repo.GetAll(context.Background(), 2, 2, nil)
// 	assert.NoError(t, err, "GetAll с пагинацией (стр 2, лимит 2) должен успешно работать")
// 	assert.Len(t, fetchedUsers, 2, "Должно быть получено 2 пользователя на второй странице")
// 	assert.Equal(t, int64(5), total, "Общее количество пользователей должно быть 5")
// 	assert.Equal(t, "Charlie", fetchedUsers[0].Name)
// 	assert.Equal(t, "David", fetchedUsers[1].Name)

// 	// Тест 4: Пагинация - Страница 3, лимит 2 (остаток)
// 	fetchedUsers, total, err = repo.GetAll(context.Background(), 3, 2, nil)
// 	assert.NoError(t, err, "GetAll с пагинацией (стр 3, лимит 2) должен успешно работать")
// 	assert.Len(t, fetchedUsers, 1, "Должен быть получен 1 пользователь на третьей странице")
// 	assert.Equal(t, int64(5), total, "Общее количество пользователей должно быть 5")
// 	assert.Equal(t, "Eve", fetchedUsers[0].Name)

// 	// Тест 5: Пагинация - Страница 4, лимит 2 (пустая страница)
// 	fetchedUsers, total, err = repo.GetAll(context.Background(), 4, 2, nil)
// 	assert.NoError(t, err, "GetAll с пагинацией (стр 4, лимит 2) должен успешно работать")
// 	assert.Len(t, fetchedUsers, 0, "На четвертой странице не должно быть пользователей")
// 	assert.Equal(t, int64(5), total, "Общее количество пользователей должно быть 5")

// 	// Тест 6: Фильтр по возрасту (min_age)
// 	filters := map[string]interface{}{"min_age": 30}
// 	fetchedUsers, total, err = repo.GetAll(context.Background(), 1, 100, filters)
// 	assert.NoError(t, err, "GetAll с фильтром min_age должен успешно работать")
// 	assert.Len(t, fetchedUsers, 3, "Должно быть получено 3 пользователя с возрастом >= 30") // Bob, Charlie, David
// 	assert.Equal(t, int64(3), total, "Общее количество пользователей с возрастом >= 30 должно быть 3")

// 	// Тест 7: Фильтр по возрасту (max_age)
// 	filters = map[string]interface{}{"max_age": 30}
// 	fetchedUsers, total, err = repo.GetAll(context.Background(), 1, 100, filters)
// 	assert.NoError(t, err, "GetAll с фильтром max_age должен успешно работать")
// 	assert.Len(t, fetchedUsers, 3, "Должно быть получено 3 пользователя с возрастом <= 30") // Alice, Bob, Eve
// 	assert.Equal(t, int64(3), total, "Общее количество пользователей с возрастом <= 30 должно быть 3")

// 	// Тест 8: Фильтры по возрасту (min_age и max_age)
// 	filters = map[string]interface{}{"min_age": 30, "max_age": 35}
// 	fetchedUsers, total, err = repo.GetAll(context.Background(), 1, 100, filters)
// 	assert.NoError(t, err, "GetAll с фильтрами min_age и max_age должен успешно работать")
// 	assert.Len(t, fetchedUsers, 2, "Должно быть получено 2 пользователя с возрастом от 30 до 35") // Bob, Charlie
// 	assert.Equal(t, int64(2), total, "Общее количество пользователей с возрастом от 30 до 35 должно быть 2")

// 	// Тест 9: Фильтр по имени (частичное совпадение, без учета регистра)
// 	filters = map[string]interface{}{"name": "bo"} // Должен найти Bob
// 	fetchedUsers, total, err = repo.GetAll(context.Background(), 1, 100, filters)
// 	assert.NoError(t, err, "GetAll с фильтром по имени должен успешно работать")
// 	assert.Len(t, fetchedUsers, 1, "Должен быть получен 1 пользователь по части имени 'bo'")
// 	assert.Equal(t, int64(1), total, "Общее количество пользователей по части имени 'bo' должно быть 1")
// 	assert.Equal(t, "Bob", fetchedUsers[0].Name)

// 	// Тест 10: Комбинированные фильтры (имя и возраст)
// 	filters = map[string]interface{}{"name": "e", "min_age": 28} // Должен найти Eve
// 	fetchedUsers, total, err = repo.GetAll(context.Background(), 1, 100, filters)
// 	assert.NoError(t, err, "GetAll с комбинированными фильтрами должен успешно работать")
// 	assert.Len(t, fetchedUsers, 1, "Должен быть получен 1 пользователь по имени 'e' и возрасту >= 28")
// 	assert.Equal(t, int64(1), total, "Общее количество пользователей по комбинированным фильтрам должно быть 1")
// 	assert.Equal(t, "Eve", fetchedUsers[0].Name)

// 	// Тест 11: Фильтры, не находящие никого
// 	filters = map[string]interface{}{"name": "xyz", "min_age": 100}
// 	fetchedUsers, total, err = repo.GetAll(context.Background(), 1, 100, filters)
// 	assert.NoError(t, err, "GetAll с фильтрами, не находящими никого, должен успешно работать")
// 	assert.Len(t, fetchedUsers, 0, "Не должно быть получено пользователей по фильтрам, не находящим никого")
// 	assert.Equal(t, int64(0), total, "Общее количество пользователей по фильтрам, не находящим никого, должно быть 0")

// 	// Тест 12: Пустая база данных
// 	repoEmpty, cleanupEmpty := setupTestRepository(t) // Новая пустая база
// 	defer cleanupEmpty()
// 	fetchedUsersEmpty, totalEmpty, errEmpty := repoEmpty.GetAll(context.Background(), 1, 100, nil)
// 	assert.NoError(t, errEmpty, "GetAll на пустой базе данных должен успешно работать")
// 	assert.Len(t, fetchedUsersEmpty, 0, "На пустой базе данных не должно быть получено пользователей")
// 	assert.Equal(t, int64(0), totalEmpty, "На пустой базе данных общее количество пользователей должно быть 0")
// }

// // GormUserRepository - приводим тип для доступа к приватному полю db, если нужно для специфических проверок
// type gormUserRepository struct {
// 	db  *gorm.DB
// 	log *logrus.Logger
// }
