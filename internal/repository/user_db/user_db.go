package user_db

import (
	"context"
	"errors"

	"github.com/IlyushinDM/user-order-api/internal/models/user_model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	Create(ctx context.Context, user *user_model.User) error
	Update(ctx context.Context, user *user_model.User) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*user_model.User, error)
	GetByEmail(ctx context.Context, email string) (*user_model.User, error)
	GetAll(ctx context.Context, page, limit int, filters map[string]interface{}) ([]user_model.User, int64, error)
}

type GormUserRepository struct {
	db  *gorm.DB
	log *logrus.Logger
}

// NewGormUserRepository creates a new user repository using GORM.
func NewGormUserRepository(db *gorm.DB, log *logrus.Logger) UserRepository {
	return &GormUserRepository{db: db, log: log}
}

func (r *GormUserRepository) Create(ctx context.Context, user *user_model.User) error {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.Create")
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to create user")
		return result.Error
	}
	logger.WithField("user_id", user.ID).Info("User created successfully")
	return nil
}

func (r *GormUserRepository) Update(ctx context.Context, user *user_model.User) error {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.Update").WithField("user_id", user.ID)
	// Use Updates to only update non-zero fields or specified fields in a map
	// Use Select("*") with Save() if you want to overwrite all fields, including clearing some.
	// Here, assuming `user` object contains only the fields to be updated (usually handled by service layer).
	result := r.db.WithContext(ctx).Model(user).Updates(user) // Updates non-zero fields
	// If you need to update specific fields including zero values, use Select:
	// result := r.db.WithContext(ctx).Model(user).Select("Name", "Email", "Age").Updates(user)
	// Or use Save to replace the whole record (ensure ID is set):
	// result := r.db.WithContext(ctx).Save(user)

	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to update user")
		return result.Error
	}
	if result.RowsAffected == 0 {
		logger.Warn("User update attempted but no rows affected (possibly user not found or no changes)")
		// GORM doesn't return ErrRecordNotFound on Updates if the record doesn't exist with Model(&user).
		// You might need a GetByID check before Update in the service layer if this is critical.
		// return gorm.ErrRecordNotFound // Or a custom error
	}
	logger.Info("User updated successfully")
	return nil
}

func (r *GormUserRepository) Delete(ctx context.Context, id uint) error {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.Delete").WithField("user_id", id)
	// GORM's default Delete performs a soft delete if the model has DeletedAt field
	result := r.db.WithContext(ctx).Delete(&user_model.User{}, id)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to delete user")
		return result.Error
	}
	if result.RowsAffected == 0 {
		logger.Warn("User deletion attempted but no rows affected (user not found)")
		return gorm.ErrRecordNotFound // Make it explicit
	}
	logger.Info("User deleted successfully (soft delete)")
	return nil
}

func (r *GormUserRepository) GetByID(ctx context.Context, id uint) (*user_model.User, error) {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.GetByID").WithField("user_id", id)
	var user user_model.User
	// Use Preload to load associated orders if needed, e.g., r.db.WithContext(ctx).Preload("Orders").First(&user, id)
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Warn("User not found")
		} else {
			logger.WithError(result.Error).Error("Failed to get user by ID")
		}
		return nil, result.Error
	}
	logger.Info("User retrieved successfully")
	return &user, nil
}

func (r *GormUserRepository) GetByEmail(ctx context.Context, email string) (*user_model.User, error) {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.GetByEmail").WithField("email", email)
	var user user_model.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Warn("User not found by email")
		} else {
			logger.WithError(result.Error).Error("Failed to get user by email")
		}
		return nil, result.Error
	}
	logger.Info("User retrieved successfully by email")
	return &user, nil
}

func (r *GormUserRepository) GetAll(
	ctx context.Context,
	page, limit int,
	filters map[string]any) ([]user_model.User, int64, error) {
	logger := r.log.WithContext(ctx).WithField("method", "UserRepository.GetAll")
	var users []user_model.User
	var total int64

	query := r.db.WithContext(ctx).Model(&user_model.User{})

	// Apply filters
	if minAge, ok := filters["min_age"].(int); ok && minAge > 0 {
		query = query.Where("age >= ?", minAge)
		logger.Debugf("Applying filter: age >= %d", minAge)
	}
	if maxAge, ok := filters["max_age"].(int); ok && maxAge > 0 {
		query = query.Where("age <= ?", maxAge)
		logger.Debugf("Applying filter: age <= %d", maxAge)
	}
	// Add more filters as needed (e.g., by name)
	if name, ok := filters["name"].(string); ok && name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%") // Case-insensitive search
		logger.Debugf("Applying filter: name ILIKE %%%s%%", name)
	}

	// Count total records matching filters
	countQuery := query // Create a copy for counting before applying limit/offset
	if err := countQuery.Count(&total).Error; err != nil {
		logger.WithError(err).Error("Failed to count users")
		return nil, 0, err
	}

	logger.WithFields(logrus.Fields{
		"filters_applied":                 filters,
		"total_records_before_pagination": total,
	}).Debug("Counted users matching filters")

	// Apply pagination
	offset := (page - 1) * limit
	result := query.Offset(offset).Limit(limit).Find(&users)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to retrieve paginated users")
		return nil, 0, result.Error
	}

	logger.WithFields(logrus.Fields{
		"page":            page,
		"limit":           limit,
		"offset":          offset,
		"retrieved_count": len(users),
		"total_count":     total,
	}).Info("Users retrieved successfully")

	return users, total, nil
}
