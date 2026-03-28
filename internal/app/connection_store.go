package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	connectionStoreDirName  = "service-bus-tui"
	connectionStoreFileName = "connections.json"
)

var (
	ErrEmptyConnectionName   = errors.New("connection name cannot be empty")
	ErrEmptyConnectionString = errors.New("connection string cannot be empty")
)

type SavedConnection struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ConnectionString string    `json:"connectionString"`
	CreatedAt        time.Time `json:"createdAt"`
	LastUsedAt       time.Time `json:"lastUsedAt"`
}

type connectionStorePayload struct {
	Connections []SavedConnection `json:"connections"`
}

type ConnectionStore struct {
	path string
}

func NewConnectionStore() *ConnectionStore {
	storePath, err := defaultConnectionStorePath()
	if err != nil {
		storePath = filepath.Join(connectionStoreDirName, connectionStoreFileName)
	}
	return &ConnectionStore{path: storePath}
}

func NewConnectionStoreAtPath(path string) *ConnectionStore {
	return &ConnectionStore{path: path}
}

func (s *ConnectionStore) List() ([]SavedConnection, error) {
	payload, err := s.readPayload()
	if err != nil {
		return nil, err
	}

	connections := append([]SavedConnection(nil), payload.Connections...)
	sort.Slice(connections, func(i, j int) bool {
		return connections[i].LastUsedAt.After(connections[j].LastUsedAt)
	})

	return connections, nil
}

func (s *ConnectionStore) Save(name, connectionString string) (SavedConnection, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return SavedConnection{}, ErrEmptyConnectionName
	}

	normalizedConnectionString := strings.TrimSpace(connectionString)
	if normalizedConnectionString == "" {
		return SavedConnection{}, ErrEmptyConnectionString
	}

	payload, err := s.readPayload()
	if err != nil {
		return SavedConnection{}, err
	}

	now := time.Now().UTC()
	for i := range payload.Connections {
		if normalizeConnectionString(payload.Connections[i].ConnectionString) == normalizedConnectionString {
			payload.Connections[i].Name = trimmedName
			payload.Connections[i].ConnectionString = normalizedConnectionString
			if payload.Connections[i].CreatedAt.IsZero() {
				payload.Connections[i].CreatedAt = now
			}
			payload.Connections[i].LastUsedAt = now

			if err := s.writePayload(payload); err != nil {
				return SavedConnection{}, err
			}
			return payload.Connections[i], nil
		}
	}

	newEntry := SavedConnection{
		ID:               uuid.NewString(),
		Name:             trimmedName,
		ConnectionString: normalizedConnectionString,
		CreatedAt:        now,
		LastUsedAt:       now,
	}

	payload.Connections = append(payload.Connections, newEntry)
	if err := s.writePayload(payload); err != nil {
		return SavedConnection{}, err
	}

	return newEntry, nil
}

func (s *ConnectionStore) Delete(id string) error {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return nil
	}

	payload, err := s.readPayload()
	if err != nil {
		return err
	}

	filtered := payload.Connections[:0]
	for _, conn := range payload.Connections {
		if conn.ID != trimmedID {
			filtered = append(filtered, conn)
		}
	}
	payload.Connections = filtered

	return s.writePayload(payload)
}

func (s *ConnectionStore) readPayload() (connectionStorePayload, error) {
	bytes, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return connectionStorePayload{Connections: []SavedConnection{}}, nil
		}
		return connectionStorePayload{}, fmt.Errorf("read saved connections: %w", err)
	}

	if len(bytes) == 0 {
		return connectionStorePayload{Connections: []SavedConnection{}}, nil
	}

	var payload connectionStorePayload
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return connectionStorePayload{}, fmt.Errorf("parse saved connections: %w", err)
	}
	if payload.Connections == nil {
		payload.Connections = []SavedConnection{}
	}

	return payload, nil
}

func (s *ConnectionStore) writePayload(payload connectionStorePayload) error {
	if payload.Connections == nil {
		payload.Connections = []SavedConnection{}
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	bytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("encode saved connections: %w", err)
	}
	bytes = append(bytes, '\n')

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, bytes, 0o600); err != nil {
		return fmt.Errorf("write saved connections temp file: %w", err)
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace saved connections file: %w", err)
	}

	if err := os.Chmod(s.path, 0o600); err != nil {
		return fmt.Errorf("set saved connections file permissions: %w", err)
	}

	return nil
}

func defaultConnectionStorePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(configDir, connectionStoreDirName, connectionStoreFileName), nil
}

func normalizeConnectionString(value string) string {
	return strings.TrimSpace(value)
}
