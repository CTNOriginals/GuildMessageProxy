package storage

import "errors"

// MockStore implements Store interface for testing.
// It combines tracking flags for verifying calls with error injection for error path testing.
type MockStore struct {
	// Internal maps for storing data
	guilds        map[string]*Guild
	guildConfigs  map[string]*GuildConfig
	proxyMessages map[string]*ProxyMessage

	// Tracking flags for verifying method calls
	SaveProxyMessageCalled  bool
	GetProxyMessageCalled   bool
	UpdateProxyMessageCalled bool

	// Error injection fields for testing error paths
	SaveError   error
	GetError    error
	UpdateError error
	DeleteError error
}

// NewMockStore creates a new MockStore with initialized maps.
func NewMockStore() *MockStore {
	return &MockStore{
		guilds:        make(map[string]*Guild),
		guildConfigs:  make(map[string]*GuildConfig),
		proxyMessages: make(map[string]*ProxyMessage),
	}
}

// SaveGuild stores a guild. Returns SaveError if set.
func (m *MockStore) SaveGuild(guildID, name string) error {
	if m.SaveError != nil {
		return m.SaveError
	}
	m.guilds[guildID] = &Guild{ID: guildID, Name: name}
	return nil
}

// GetGuild retrieves a guild by ID. Returns GetError if set.
func (m *MockStore) GetGuild(guildID string) (*Guild, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	return m.guilds[guildID], nil
}

// DeleteGuild removes a guild and its config. Returns DeleteError if set.
func (m *MockStore) DeleteGuild(guildID string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}
	delete(m.guilds, guildID)
	delete(m.guildConfigs, guildID)
	return nil
}

// SaveGuildConfig stores guild configuration. Returns SaveError if set.
func (m *MockStore) SaveGuildConfig(config GuildConfig) error {
	if m.SaveError != nil {
		return m.SaveError
	}
	m.guildConfigs[config.GuildID] = &config
	return nil
}

// GetGuildConfig retrieves guild configuration. Returns GetError if set.
func (m *MockStore) GetGuildConfig(guildID string) (*GuildConfig, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	return m.guildConfigs[guildID], nil
}

// SaveProxyMessage stores a proxy message.
// Sets SaveProxyMessageCalled flag and returns SaveError if set.
func (m *MockStore) SaveProxyMessage(msg ProxyMessage) error {
	m.SaveProxyMessageCalled = true
	if m.SaveError != nil {
		return m.SaveError
	}
	var key = msg.GuildID + ":" + msg.MessageID
	m.proxyMessages[key] = &msg
	return nil
}

// GetProxyMessage retrieves a proxy message by guild and message ID.
// Sets GetProxyMessageCalled flag and returns GetError if set.
func (m *MockStore) GetProxyMessage(guildID, messageID string) (*ProxyMessage, error) {
	m.GetProxyMessageCalled = true
	if m.GetError != nil {
		return nil, m.GetError
	}
	var key = guildID + ":" + messageID
	return m.proxyMessages[key], nil
}

// UpdateProxyMessage updates an existing proxy message.
// Sets UpdateProxyMessageCalled flag, checks existence, and returns UpdateError if set.
func (m *MockStore) UpdateProxyMessage(msg ProxyMessage) error {
	m.UpdateProxyMessageCalled = true
	if m.UpdateError != nil {
		return m.UpdateError
	}
	var key = msg.GuildID + ":" + msg.MessageID
	if _, exists := m.proxyMessages[key]; !exists {
		return errors.New("proxy message not found")
	}
	m.proxyMessages[key] = &msg
	return nil
}

// DeleteProxyMessage removes a proxy message. Returns DeleteError if set.
func (m *MockStore) DeleteProxyMessage(guildID, messageID string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}
	var key = guildID + ":" + messageID
	delete(m.proxyMessages, key)
	return nil
}
