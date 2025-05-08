package order_db_test

// import (
// 	"context"
// 	"testing"

// 	"github.com/IlyushinDM/user-order-api/internal/models/order_model"
// 	"github.com/IlyushinDM/user-order-api/internal/models/user_model"   // Нужна модель пользователя для внешнего ключа
// 	"github.com/IlyushinDM/user-order-api/internal/repository/order_db" // Убедитесь, что путь импорта правильный
// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// 	"gorm.io/gorm/logger"
// )

// var (
// 	mockLog *logrus.Logger // Логгер для тестов
// )

// // setupTestDB инициализирует in-memory SQLite базу данных для тестов, включая модели User и Order.
// // Возвращает *gorm.DB и функцию очистки (cleanup).
// func setupTestDB(t *testing.T) (*gorm.DB, func()) {
// 	// Инициализируем логгер для тестов
// 	if mockLog == nil {
// 		mockLog = logrus.New()
// 		mockLog.SetOutput(nil)              // Отключаем вывод логов в консоль
// 		mockLog.SetLevel(logrus.PanicLevel) // Устанавливаем высокий уровень
// 	}

// 	// Открываем in-memory SQLite базу данных
// 	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
// 		Logger: logger.Default.LogMode(logger.Silent), // Отключаем логирование GORM
// 	})
// 	if err != nil {
// 		t.Fatalf("Не удалось открыть in-memory SQLite базу данных: %v", err)
// 	}

// 	// Автомиграция моделей User и Order
// 	err = db.AutoMigrate(&user_model.User{}, &order_model.Order{})
// 	if err != nil {
// 		t.Fatalf("Не удалось выполнить автомиграцию: %v", err)
// 	}

// 	// Возвращаем подключение к БД и функцию очистки.
// 	sqlDB, _ := db.DB()
// 	return db, func() {
// 		sqlDB.Close()
// 	}
// }

// // setupTestRepository настраивает тестовое окружение для репозитория заказов.
// // Возвращает экземпляр OrderRepository и функцию очистки.
// func setupTestRepository(t *testing.T) (order_db.OrderRepository, func()) {
// 	db, cleanup := setupTestDB(t)
// 	repo := order_db.NewGormOrderRepository(db, mockLog)
// 	return repo, cleanup
// }

// // TestCreateOrder проверяет успешное создание заказа.
// func TestCreateOrder(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем пользователя, к которому будет привязан заказ
// 	user := &user_model.User{Name: "Order User", Email: "order.user@example.com", Age: 25}
// 	err := repo.(*order_db.GormOrderRepository).DB.Create(user).Error // Используем DB напрямую для создания пользователя
// 	assert.NoError(t, err, "Не удалось создать пользователя для теста заказа")

// 	order := &order_model.Order{
// 		UserID:      user.ID,
// 		ProductName: "Test Product",
// 		Quantity:    2,
// 		Price:       100.50,
// 	}

// 	err = repo.Create(context.Background(), order)
// 	assert.NoError(t, err, "Create должен успешно создать заказ")
// 	assert.NotZero(t, order.ID, "После создания ID заказа должен быть установлен")
// 	assert.Equal(t, user.ID, order.UserID, "UserID в созданном заказе должен совпадать с ID пользователя")

// 	// Проверяем, что заказ действительно был сохранен и привязан к пользователю
// 	fetchedOrder, err := repo.GetByID(context.Background(), order.ID, user.ID) // Ищем по ID заказа и ID пользователя
// 	assert.NoError(t, err, "GetByID должен найти созданный заказ для правильного пользователя")
// 	assert.NotNil(t, fetchedOrder, "Найденный заказ не должен быть nil")
// 	assert.Equal(t, order.ProductName, fetchedOrder.ProductName, "Имя продукта должно совпадать")
// 	assert.Equal(t, order.UserID, fetchedOrder.UserID, "UserID в найденном заказе должен совпадать")
// }

// // TestUpdateOrder проверяет успешное обновление заказа для правильного пользователя.
// func TestUpdateOrder(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем пользователя и заказ для него
// 	user := &user_model.User{Name: "Update User", Email: "update.user@example.com", Age: 30}
// 	err := repo.(*order_db.GormOrderRepository).DB.Create(user).Error
// 	assert.NoError(t, err, "Не удалось создать пользователя для теста обновления заказа")

// 	order := &order_model.Order{UserID: user.ID, ProductName: "Old Product", Quantity: 1, Price: 50.00}
// 	err = repo.Create(context.Background(), order)
// 	assert.NoError(t, err, "Не удалось создать заказ для теста обновления")
// 	orderID := order.ID

// 	// Обновляем поля заказа
// 	order.ProductName = "New Product"
// 	order.Quantity = 3
// 	order.Price = 150.00
// 	// Важно: UserID не меняется в объекте обновления, чтобы проверить, что репозиторий учитывает его в WHERE
// 	// order.UserID = 999 // НЕ ДОЛЖНО СЛУЧИТЬСЯ

// 	err = repo.Update(context.Background(), order) // Передаем объект с обновленными полями и тем же ID/UserID
// 	assert.NoError(t, err, "Update должен успешно обновить заказ для правильного пользователя")

// 	// Проверяем, что заказ был обновлен в БД и принадлежит тому же пользователю
// 	fetchedOrder, err := repo.GetByID(context.Background(), orderID, user.ID)
// 	assert.NoError(t, err, "Не удалось найти обновленный заказ по ID для правильного пользователя")
// 	assert.NotNil(t, fetchedOrder, "Обновленный заказ не должен быть nil")
// 	assert.Equal(t, "New Product", fetchedOrder.ProductName, "Имя продукта должно быть обновлено")
// 	assert.Equal(t, 3, fetchedOrder.Quantity, "Количество должно быть обновлено")
// 	assert.Equal(t, 150.00, fetchedOrder.Price, "Цена должна быть обновлена")
// 	assert.Equal(t, user.ID, fetchedOrder.UserID, "UserID обновленного заказа не должен был измениться")
// }

// // TestUpdateOrder_NotFound проверяет обновление несуществующего заказа.
// func TestUpdateOrder_NotFound(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем пользователя, чтобы проверить, что даже с существующим пользователем,
// 	// обновление несуществующего заказа не проходит.
// 	user := &user_model.User{Name: "User For Not Found Update", Email: "notfound.update@example.com", Age: 30}
// 	err := repo.(*order_db.GormOrderRepository).DB.Create(user).Error
// 	assert.NoError(t, err, "Не удалось создать пользователя для теста обновления несуществующего заказа")

// 	// Попытка обновить заказ с ID=999, которого нет в БД
// 	nonExistentOrder := &order_model.Order{
// 		ID:          999,
// 		UserID:      user.ID, // Привязываем к существующему пользователю
// 		ProductName: "Non Existent Product",
// 		Quantity:    1,
// 		Price:       10.00,
// 	}

// 	err = repo.Update(context.Background(), nonExistentOrder)
// 	// Ожидаем ErrRecordNotFound, как реализовано в репозитории при RowsAffected == 0 после проверки существования записи.
// 	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "Update несуществующего заказа должен вернуть ErrRecordNotFound")
// }

// // TestUpdateOrder_WrongUser проверяет обновление заказа, принадлежащего другому пользователю.
// func TestUpdateOrder_WrongUser(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем двух пользователей
// 	user1 := &user_model.User{Name: "User 1 Update", Email: "user1.update@example.com", Age: 30}
// 	user2 := &user_model.User{Name: "User 2 Update", Email: "user2.update@example.com", Age: 35}
// 	err1 := repo.(*order_db.GormOrderRepository).DB.Create(user1).Error
// 	err2 := repo.(*order_db.GormOrderRepository).DB.Create(user2).Error
// 	assert.NoError(t, err1, "Не удалось создать пользователя 1 для теста")
// 	assert.NoError(t, err2, "Не удалось создать пользователя 2 для теста")

// 	// Создаем заказ для первого пользователя
// 	orderUser1 := &order_model.Order{UserID: user1.ID, ProductName: "User 1 Product", Quantity: 1, Price: 10.00}
// 	err := repo.Create(context.Background(), orderUser1)
// 	assert.NoError(t, err, "Не удалось создать заказ для пользователя 1")
// 	orderUser1ID := orderUser1.ID

// 	// Попытка обновить заказ пользователя 1, используя ID пользователя 2
// 	// Создаем объект обновления, как если бы пользователь 2 пытался обновить этот заказ.
// 	updateData := &order_model.Order{
// 		ID:          orderUser1ID, // ID заказа пользователя 1
// 		UserID:      user2.ID,     // ID пользователя 2 - НЕВЕРНЫЙ UserID
// 		ProductName: "Updated by Wrong User",
// 		Quantity:    99,
// 	}

// 	err = repo.Update(context.Background(), updateData)
// 	// Ожидаем ErrRecordNotFound, потому что запрос WHERE id = ? AND user_id = ? не найдет запись с user_id пользователя 2.
// 	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "Update заказа другого пользователя должен вернуть ErrRecordNotFound или ошибку доступа")

// 	// Проверяем, что заказ остался без изменений
// 	fetchedOrder, err := repo.GetByID(context.Background(), orderUser1ID, user1.ID) // Ищем оригинальным пользователем
// 	assert.NoError(t, err, "Оригинальный пользователь должен найти свой заказ")
// 	assert.NotNil(t, fetchedOrder, "Оригинальный заказ не должен быть nil")
// 	assert.Equal(t, "User 1 Product", fetchedOrder.ProductName, "Имя продукта не должно было измениться после попытки обновления другим пользователем")
// 	assert.Equal(t, 1, fetchedOrder.Quantity, "Количество не должно было измениться")
// }

// // TestDeleteOrder проверяет успешное удаление заказа для правильного пользователя (мягкое удаление).
// func TestDeleteOrder(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем пользователя и заказ для него
// 	user := &user_model.User{Name: "Delete User", Email: "delete.user@example.com", Age: 40}
// 	err := repo.(*order_db.GormOrderRepository).DB.Create(user).Error
// 	assert.NoError(t, err, "Не удалось создать пользователя для теста удаления заказа")

// 	order := &order_model.Order{UserID: user.ID, ProductName: "Product to Delete", Quantity: 5, Price: 200.00}
// 	err = repo.Create(context.Background(), order)
// 	assert.NoError(t, err, "Не удалось создать заказ для теста удаления")
// 	orderID := order.ID

// 	// Удаляем заказ, указывая ID пользователя
// 	err = repo.Delete(context.Background(), orderID, user.ID)
// 	assert.NoError(t, err, "Delete должен успешно удалить заказ для правильного пользователя")

// 	// Проверяем, что заказ помечен как удаленный
// 	fetchedOrder, err := repo.GetByID(context.Background(), orderID, user.ID)
// 	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "GetByID не должен найти мягко удаленный заказ")
// 	assert.Nil(t, fetchedOrder, "Мягко удаленный заказ должен быть nil при поиске")

// 	// Проверяем наличие записи с непустым DeletedAt через Unscoped
// 	var deletedOrder order_model.Order
// 	result := repo.(*order_db.GormOrderRepository).DB.Unscoped().First(&deletedOrder, orderID)
// 	assert.NoError(t, result.Error, "Unscoped должен найти мягко удаленный заказ")
// 	assert.NotNil(t, deletedOrder.DeletedAt, "Поле DeletedAt должно быть установлено для мягко удаленного заказа")
// 	assert.Equal(t, user.ID, deletedOrder.UserID, "Удаленный заказ должен сохранять UserID")
// }

// // TestDeleteOrder_NotFound проверяет удаление несуществующего заказа.
// func TestDeleteOrder_NotFound(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем пользователя
// 	user := &user_model.User{Name: "User For Not Found Delete", Email: "notfound.delete@example.com", Age: 30}
// 	err := repo.(*order_db.GormOrderRepository).DB.Create(user).Error
// 	assert.NoError(t, err, "Не удалось создать пользователя для теста удаления несуществующего заказа")

// 	// Попытка удалить заказ с ID=999, которого нет, для существующего пользователя
// 	err = repo.Delete(context.Background(), 999, user.ID)
// 	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "Delete несуществующего заказа должен вернуть ErrRecordNotFound")
// }

// // TestDeleteOrder_WrongUser проверяет удаление заказа, принадлежащего другому пользователю.
// func TestDeleteOrder_WrongUser(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем двух пользователей
// 	user1 := &user_model.User{Name: "User 1 Delete", Email: "user1.delete@example.com", Age: 30}
// 	user2 := &user_model.User{Name: "User 2 Delete", Email: "user2.delete@example.com", Age: 35}
// 	err1 := repo.(*order_db.GormOrderRepository).DB.Create(user1).Error
// 	err2 := repo.(*order_db.GormOrderRepository).DB.Create(user2).Error
// 	assert.NoError(t, err1, "Не удалось создать пользователя 1 для теста")
// 	assert.NoError(t, err2, "Не удалось создать пользователя 2 для теста")

// 	// Создаем заказ для первого пользователя
// 	orderUser1 := &order_model.Order{UserID: user1.ID, ProductName: "User 1 Product To Delete", Quantity: 1, Price: 10.00}
// 	err := repo.Create(context.Background(), orderUser1)
// 	assert.NoError(t, err, "Не удалось создать заказ для пользователя 1")
// 	orderUser1ID := orderUser1.ID

// 	// Попытка удалить заказ пользователя 1, используя ID пользователя 2
// 	err = repo.Delete(context.Background(), orderUser1ID, user2.ID) // ID заказа пользователя 1, ID пользователя 2 - НЕВЕРНЫЙ UserID
// 	// Ожидаем ошибку "permission denied or record not found", как реализовано в репозитории
// 	assert.Error(t, err, "Delete заказа другого пользователя должен вернуть ошибку доступа")
// 	assert.EqualError(t, err, "permission denied or record not found", "Текст ошибки должен соответствовать 'permission denied or record not found'")

// 	// Проверяем, что заказ не был удален (все еще существует для оригинального пользователя)
// 	fetchedOrder, err := repo.GetByID(context.Background(), orderUser1ID, user1.ID)
// 	assert.NoError(t, err, "Оригинальный пользователь должен найти свой заказ после неудачной попытки удаления другим")
// 	assert.NotNil(t, fetchedOrder, "Оригинальный заказ не должен быть nil")
// 	assert.Nil(t, fetchedOrder.DeletedAt, "Поле DeletedAt не должно быть установлено") // Проверяем, что мягкое удаление не произошло
// }

// // TestGetByID проверяет получение заказа по ID для правильного пользователя.
// func TestGetByID(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем пользователя и заказ
// 	user := &user_model.User{Name: "Get By ID User", Email: "getbyid.user@example.com", Age: 28}
// 	err := repo.(*order_db.GormOrderRepository).DB.Create(user).Error
// 	assert.NoError(t, err, "Не удалось создать пользователя для теста GetByID")

// 	order := &order_model.Order{UserID: user.ID, ProductName: "Product for GetByID", Quantity: 1, Price: 99.99}
// 	err = repo.Create(context.Background(), order)
// 	assert.NoError(t, err, "Не удалось создать заказ для теста GetByID")
// 	orderID := order.ID

// 	// Получаем заказ, указывая его ID и ID пользователя
// 	fetchedOrder, err := repo.GetByID(context.Background(), orderID, user.ID)
// 	assert.NoError(t, err, "GetByID должен успешно найти заказ для правильного пользователя")
// 	assert.NotNil(t, fetchedOrder, "Найденный заказ не должен быть nil")
// 	assert.Equal(t, orderID, fetchedOrder.ID, "ID заказа должен совпадать")
// 	assert.Equal(t, user.ID, fetchedOrder.UserID, "UserID в найденном заказе должен совпадать")
// 	assert.Equal(t, "Product for GetByID", fetchedOrder.ProductName, "Имя продукта должно совпадать")
// }

// // TestGetByID_NotFound проверяет получение несуществующего заказа.
// func TestGetByID_NotFound(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем пользователя
// 	user := &user_model.User{Name: "User For GetByID NotFound", Email: "getbyid.notfound@example.com", Age: 30}
// 	err := repo.(*order_db.GormOrderRepository).DB.Create(user).Error
// 	assert.NoError(t, err, "Не удалось создать пользователя для теста GetByID NotFound")

// 	// Попытка получить заказ с ID=999, которого нет, для существующего пользователя
// 	fetchedOrder, err := repo.GetByID(context.Background(), 999, user.ID)
// 	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "GetByID несуществующего заказа должен вернуть ErrRecordNotFound")
// 	assert.Nil(t, fetchedOrder, "Найденный заказ должен быть nil")
// }

// // TestGetByID_WrongUser проверяет получение заказа, принадлежащего другому пользователю.
// func TestGetByID_WrongUser(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем двух пользователей
// 	user1 := &user_model.User{Name: "User 1 GetByID", Email: "user1.getbyid@example.com", Age: 30}
// 	user2 := &user_model.User{Name: "User 2 GetByID", Email: "user2.getbyid@example.com", Age: 35}
// 	err1 := repo.(*order_db.GormOrderRepository).DB.Create(user1).Error
// 	err2 := repo.(*order_db.GormOrderRepository).DB.Create(user2).Error
// 	assert.NoError(t, err1, "Не удалось создать пользователя 1 для теста")
// 	assert.NoError(t, err2, "Не удалось создать пользователя 2 для теста")

// 	// Создаем заказ для первого пользователя
// 	orderUser1 := &order_model.Order{UserID: user1.ID, ProductName: "User 1 Product GetByID", Quantity: 1, Price: 10.00}
// 	err := repo.Create(context.Background(), orderUser1)
// 	assert.NoError(t, err, "Не удалось создать заказ для пользователя 1")
// 	orderUser1ID := orderUser1.ID

// 	// Попытка получить заказ пользователя 1, используя ID пользователя 2
// 	fetchedOrder, err := repo.GetByID(context.Background(), orderUser1ID, user2.ID) // ID заказа пользователя 1, ID пользователя 2 - НЕВЕРНЫЙ UserID
// 	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "GetByID заказа другого пользователя должен вернуть ErrRecordNotFound")
// 	assert.Nil(t, fetchedOrder, "Найденный заказ должен быть nil")
// }

// // TestGetAllByUser проверяет получение всех заказов для конкретного пользователя с пагинацией.
// func TestGetAllByUser(t *testing.T) {
// 	repo, cleanup := setupTestRepository(t)
// 	defer cleanup()

// 	// Создаем двух пользователей
// 	user1 := &user_model.User{Name: "User 1 GetAll", Email: "user1.getall@example.com", Age: 30}
// 	user2 := &user_model.User{Name: "User 2 GetAll", Email: "user2.getall@example.com", Age: 35}
// 	err1 := repo.(*order_db.GormOrderRepository).DB.Create(user1).Error
// 	err2 := repo.(*order_db.GormOrderRepository).DB.Create(user2).Error
// 	assert.NoError(t, err1, "Не удалось создать пользователя 1 для теста GetAllByUser")
// 	assert.NoError(t, err2, "Не удалось создать пользователя 2 для теста GetAllByUser")

// 	// Создаем заказы для пользователя 1 (5 заказов)
// 	for i := 1; i <= 5; i++ {
// 		order := &order_model.Order{
// 			UserID:      user1.ID,
// 			ProductName: "Product " + string('A'+rune(i)-1) + " by User 1",
// 			Quantity:    i,
// 			Price:       float64(i) * 10.0,
// 		}
// 		err := repo.Create(context.Background(), order)
// 		assert.NoError(t, err, "Не удалось создать заказ для пользователя 1 в тесте GetAllByUser")
// 		// Небольшая задержка, чтобы гарантировать разный CreatedAt для надежной сортировки в тестах, если это важно.
// 		// time.Sleep(time.Millisecond)
// 	}

// 	// Создаем заказы для пользователя 2 (3 заказа)
// 	for i := 1; i <= 3; i++ {
// 		order := &order_model.Order{
// 			UserID:      user2.ID,
// 			ProductName: "Product " + string('X'+rune(i)-1) + " by User 2",
// 			Quantity:    i,
// 			Price:       float64(i) * 5.0,
// 		}
// 		err := repo.Create(context.Background(), order)
// 		assert.NoError(t, err, "Не удалось создать заказ для пользователя 2 в тесте GetAllByUser")
// 		// time.Sleep(time.Millisecond)
// 	}

// 	// Тест 1: Получение всех заказов для пользователя 1 без пагинации (page=1, limit=100)
// 	ordersUser1, totalUser1, errUser1 := repo.GetAllByUser(context.Background(), user1.ID, 1, 100)
// 	assert.NoError(t, errUser1, "GetAllByUser для пользователя 1 без пагинации должен успешно работать")
// 	assert.Len(t, ordersUser1, 5, "Должно быть получено 5 заказов для пользователя 1")
// 	assert.Equal(t, int64(5), totalUser1, "Общее количество заказов для пользователя 1 должно быть 5")
// 	for _, order := range ordersUser1 {
// 		assert.Equal(t, user1.ID, order.UserID, "Все полученные заказы должны принадлежать пользователю 1")
// 	}

// 	// Тест 2: Пагинация для пользователя 1 - Страница 1, лимит 2
// 	ordersUser1, totalUser1, errUser1 = repo.GetAllByUser(context.Background(), user1.ID, 1, 2)
// 	assert.NoError(t, errUser1, "GetAllByUser для пользователя 1 с пагинацией (стр 1, лимит 2) должен успешно работать")
// 	assert.Len(t, ordersUser1, 2, "Должно быть получено 2 заказа на первой странице для пользователя 1")
// 	assert.Equal(t, int64(5), totalUser1, "Общее количество заказов для пользователя 1 должно быть 5")

// 	// Тест 3: Пагинация для пользователя 1 - Страница 2, лимит 2
// 	ordersUser1, totalUser1, errUser1 = repo.GetAllByUser(context.Background(), user1.ID, 2, 2)
// 	assert.NoError(t, errUser1, "GetAllByUser для пользователя 1 с пагинацией (стр 2, лимит 2) должен успешно работать")
// 	assert.Len(t, ordersUser1, 2, "Должно быть получено 2 заказа на второй странице для пользователя 1")
// 	assert.Equal(t, int64(5), totalUser1, "Общее количество заказов для пользователя 1 должно быть 5")

// 	// Тест 4: Получение всех заказов для пользователя 2 без пагинации
// 	ordersUser2, totalUser2, errUser2 := repo.GetAllByUser(context.Background(), user2.ID, 1, 100)
// 	assert.NoError(t, errUser2, "GetAllByUser для пользователя 2 без пагинации должен успешно работать")
// 	assert.Len(t, ordersUser2, 3, "Должно быть получено 3 заказа для пользователя 2")
// 	assert.Equal(t, int64(3), totalUser2, "Общее количество заказов для пользователя 2 должно быть 3")
// 	for _, order := range ordersUser2 {
// 		assert.Equal(t, user2.ID, order.UserID, "Все полученные заказы должны принадлежать пользователю 2")
// 	}

// 	// Тест 5: Получение заказов для несуществующего пользователя
// 	ordersNotFound, totalNotFound, errNotFound := repo.GetAllByUser(context.Background(), 999, 1, 100)
// 	assert.NoError(t, errNotFound, "GetAllByUser для несуществующего пользователя должен успешно работать (вернуть пустой список)")
// 	assert.Len(t, ordersNotFound, 0, "Для несуществующего пользователя не должно быть получено заказов")
// 	assert.Equal(t, int64(0), totalNotFound, "Для несуществующего пользователя общее количество заказов должно быть 0")

// 	// Тест 6: Пользователь без заказов
// 	user3 := &user_model.User{Name: "User 3 GetAll", Email: "user3.getall@example.com", Age: 40}
// 	err3 := repo.(*order_db.GormOrderRepository).DB.Create(user3).Error
// 	assert.NoError(t, err3, "Не удалось создать пользователя 3 для теста GetAllByUser")
// 	ordersUser3, totalUser3, errUser3 := repo.GetAllByUser(context.Background(), user3.ID, 1, 100)
// 	assert.NoError(t, errUser3, "GetAllByUser для пользователя без заказов должен успешно работать (вернуть пустой список)")
// 	assert.Len(t, ordersUser3, 0, "Для пользователя без заказов не должно быть получено заказов")
// 	assert.Equal(t, int64(0), totalUser3, "Для пользователя без заказов общее количество заказов должно быть 0")
// }
