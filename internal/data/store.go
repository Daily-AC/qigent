package data

import (
	"encoding/json"
	"os"
	"sync"
)

var (
	dataDir           = "data"
	conversationsFile = "data/conversations.json"
	rolesFile         = "data/roles.json"
	storeLock         sync.RWMutex
)

// EnsureDataDir creates the data directory if it doesn't exist
func EnsureDataDir() {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.Mkdir(dataDir, 0755)
	}
}

// -- Conversations --

func LoadConversations() ([]Conversation, error) {
	storeLock.RLock()
	defer storeLock.RUnlock()

	EnsureDataDir()

	data, err := os.ReadFile(conversationsFile)
	if os.IsNotExist(err) {
		return []Conversation{}, nil
	}
	if err != nil {
		return nil, err
	}

	var convs []Conversation
	if err := json.Unmarshal(data, &convs); err != nil {
		return []Conversation{}, nil
	}
	return convs, nil
}

func SaveConversation(conv Conversation) error {
	storeLock.Lock()
	defer storeLock.Unlock()

	EnsureDataDir()

	// Load existing
	convs, _ := loadConversationsUnsafe()

	// Update or Append
	found := false
	for i, c := range convs {
		if c.ID == conv.ID {
			convs[i] = conv
			found = true
			break
		}
	}
	if !found {
		convs = append([]Conversation{conv}, convs...) // Prepend new
	}

	return saveConversationsUnsafe(convs)
}

func GetConversation(id string) (*Conversation, error) {
	storeLock.RLock()
	defer storeLock.RUnlock()

	convs, err := loadConversationsUnsafe()
	if err != nil {
		return nil, err
	}

	for _, c := range convs {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil // Not found
}

func DeleteConversation(id string) error {
	storeLock.Lock()
	defer storeLock.Unlock()

	convs, err := loadConversationsUnsafe()
	if err != nil {
		return err
	}

	var newConvs []Conversation
	for _, c := range convs {
		if c.ID != id {
			newConvs = append(newConvs, c)
		}
	}

	return saveConversationsUnsafe(newConvs)
}

// Helpers without locking for internal use
func loadConversationsUnsafe() ([]Conversation, error) {
	data, err := os.ReadFile(conversationsFile)
	if os.IsNotExist(err) {
		return []Conversation{}, nil
	}
	if err != nil {
		return nil, err
	}
	var convs []Conversation
	json.Unmarshal(data, &convs)
	return convs, nil
}

func saveConversationsUnsafe(convs []Conversation) error {
	data, err := json.MarshalIndent(convs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(conversationsFile, data, 0644)
}

// -- Roles --

func LoadRoles() ([]Role, error) {
	storeLock.RLock()
	defer storeLock.RUnlock()

	EnsureDataDir()

	data, err := os.ReadFile(rolesFile)
	if os.IsNotExist(err) {
		// Return defaults if not exist
		return DefaultRoles(), nil
	}
	if err != nil {
		return nil, err
	}

	var roles []Role
	if err := json.Unmarshal(data, &roles); err != nil {
		return DefaultRoles(), nil
	}
	return roles, nil
}

func DefaultRoles() []Role {
	return []Role{
		{Name: "苏格拉底", Prompt: "你是一个苏格拉底式的哲学家，喜欢反问。"},
		{Name: "乔布斯", Prompt: "你是一个追求极致产品体验的创新者。"},
		{Name: "马斯克", Prompt: "你是一个疯狂的梦想家，思考第一性原理。"},
		{Name: "孔子", Prompt: "你是一位儒家圣人，讲究仁义礼智信。"},
		{Name: "现代大学生", Prompt: "你是一个务实的现代大学生。"},
	}
}

func AddRole(role Role) error {
	storeLock.Lock()
	defer storeLock.Unlock()

	roles, err := loadRolesUnsafe()
	if err != nil {
		// If fails to load, maybe try default? Or init empty?
		// Assume loadRolesUnsafe handles default if missing
		roles = DefaultRoles()
	}

	// Check if exists
	for _, r := range roles {
		if r.Name == role.Name {
			// Update or Error? Let's Error for duplicate name
			return nil // Or return error? For MVP we just ignore or overwrite?
			// Let's overwrite
		}
	}

	// If overwrite needed, find index. Since we didn't...
	// Simplest: Check unique name
	newRoles := []Role{}
	for _, r := range roles {
		if r.Name != role.Name {
			newRoles = append(newRoles, r)
		}
	}
	newRoles = append(newRoles, role)

	return saveRolesUnsafe(newRoles)
}

func DeleteRole(name string) error {
	storeLock.Lock()
	defer storeLock.Unlock()

	roles, err := loadRolesUnsafe()
	if err != nil {
		return err
	}

	newRoles := []Role{}
	for _, r := range roles {
		if r.Name != name {
			newRoles = append(newRoles, r)
		}
	}

	return saveRolesUnsafe(newRoles)
}

// Helpers for Roles
func loadRolesUnsafe() ([]Role, error) {
	data, err := os.ReadFile(rolesFile)
	if os.IsNotExist(err) {
		return DefaultRoles(), nil
	}
	if err != nil {
		return nil, err
	}
	var roles []Role
	if err := json.Unmarshal(data, &roles); err != nil {
		return DefaultRoles(), nil
	}
	return roles, nil
}

func saveRolesUnsafe(roles []Role) error {
	data, err := json.MarshalIndent(roles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(rolesFile, data, 0644)
}
