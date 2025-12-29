package data

import "gorm.io/gorm"

// GetChatConfig retrieves the chat configuration for a user.
// Returns a default config if none exists.
func GetChatConfig(userID uint) (*ChatConfig, error) {
	var cfg ChatConfig
	err := DB.Where("user_id = ?", userID).First(&cfg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default config
			return &ChatConfig{
				UserID:   userID,
				BaseURL:  "https://api.openai.com/v1",
				LLMModel: "gpt-3.5-turbo",
			}, nil
		}
		return nil, err
	}
	return &cfg, nil
}

// SaveChatConfig saves or updates the chat configuration for a user.
func SaveChatConfig(userID uint, cfg *ChatConfig) error {
	var existing ChatConfig
	err := DB.Where("user_id = ?", userID).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new
			cfg.UserID = userID
			return DB.Create(cfg).Error
		}
		return err
	}

	// Update existing
	existing.APIKey = cfg.APIKey
	existing.BaseURL = cfg.BaseURL
	existing.LLMModel = cfg.LLMModel
	return DB.Save(&existing).Error
}
