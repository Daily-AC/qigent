package data

import (
	"errors"

	"gorm.io/gorm"
)

// -- Users --

func CreateUser(username, passwordHash string) (*User, error) {
	user := User{Username: username, Password: passwordHash}
	err := DB.Create(&user).Error
	return &user, err
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	err := DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(id uint) (*User, error) {
	var user User
	err := DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// -- Conversations --

func CreateConversation(conv *Conversation) error {
	// GORM will create or update
	return DB.Save(conv).Error
}

func GetConversations(userID uint) ([]Conversation, error) {
	var convs []Conversation
	// Preload? History is simpler if stored as JSON blob
	err := DB.Where("user_id = ?", userID).Order("updated_at desc").Find(&convs).Error
	return convs, err
}

func GetConversation(id string) (*Conversation, error) {
	var conv Conversation
	err := DB.Where("id = ?", id).First(&conv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // Return nil if not found
	}
	return &conv, err
}

func SaveConversation(conv *Conversation) error {
	// Save includes Create or Update
	// Ensure all fields are saved
	return DB.Save(conv).Error
}

func DeleteConversation(id string, userID uint) error {
	return DB.Where("id = ? AND user_id = ?", id, userID).Delete(&Conversation{}).Error
}

// -- Roles --

func GetRoles(userID uint) ([]Role, error) {
	var roles []Role
	// Fetch System Roles (UserID=0) AND User Roles
	err := DB.Where("user_id = ? OR user_id = 0", userID).Find(&roles).Error
	return roles, err
}

func AddRole(role *Role) error {
	return DB.Create(role).Error
}

func DeleteRole(name string, userID uint) error {
	// Only delete custom roles owned by user
	return DB.Where("name = ? AND user_id = ?", name, userID).Delete(&Role{}).Error
}

// Default Roles Seeding
func SeedRoles() {
	// Ensure DB is init
	if DB == nil {
		return
	}

	var count int64
	DB.Model(&Role{}).Where("user_id = 0").Count(&count)
	if count == 0 {
		defaults := []Role{
			{Name: "苏格拉底", Prompt: "你是一个苏格拉底式的哲学家，喜欢反问。", UserID: 0},
			{Name: "乔布斯", Prompt: "你是一个追求极致产品体验的创新者。", UserID: 0},
			{Name: "马斯克", Prompt: "你是一个疯狂的梦想家，思考第一性原理。", UserID: 0},
			{Name: "孔子", Prompt: "你是一位儒家圣人，讲究仁义礼智信。", UserID: 0},
			{Name: "现代大学生", Prompt: "你是一个务实的现代大学生。", UserID: 0},
		}
		DB.Create(&defaults)
	}
}
