package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/johnpitter/ollama-code/internal/llm"
)

// Manager gerenciador de sessões
type Manager struct {
	sessionDir     string
	currentSession *Session
}

// NewManager cria novo gerenciador
func NewManager(baseDir string) *Manager {
	sessionDir := filepath.Join(baseDir, ".ollama-code", "sessions")
	os.MkdirAll(sessionDir, 0755)

	return &Manager{
		sessionDir: sessionDir,
	}
}

// New cria nova sessão
func (m *Manager) New(name, workDir, mode string) (*Session, error) {
	session := &Session{
		ID:           generateSessionID(),
		Name:         name,
		StartTime:    time.Now(),
		LastActivity: time.Now(),
		Messages:     []llm.Message{},
		WorkDir:      workDir,
		Mode:         mode,
		Metadata:     make(map[string]interface{}),
		Active:       true,
	}

	if err := m.saveSession(session); err != nil {
		return nil, err
	}

	m.currentSession = session
	return session, nil
}

// Resume retoma sessão existente
func (m *Manager) Resume(sessionID string) (*Session, error) {
	session, err := m.loadSession(sessionID)
	if err != nil {
		return nil, err
	}

	session.Active = true
	session.LastActivity = time.Now()

	if err := m.saveSession(session); err != nil {
		return nil, err
	}

	m.currentSession = session
	return session, nil
}

// Continue continua última sessão
func (m *Manager) Continue() (*Session, error) {
	sessions, err := m.List(1)
	if err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return nil, fmt.Errorf("no sessions found")
	}

	return m.Resume(sessions[0].ID)
}

// Save salva sessão atual
func (m *Manager) Save() error {
	if m.currentSession == nil {
		return fmt.Errorf("no active session")
	}

	m.currentSession.LastActivity = time.Now()
	return m.saveSession(m.currentSession)
}

// End encerra sessão atual
func (m *Manager) End() error {
	if m.currentSession == nil {
		return nil
	}

	m.currentSession.Active = false
	m.currentSession.LastActivity = time.Now()

	if err := m.saveSession(m.currentSession); err != nil {
		return err
	}

	m.currentSession = nil
	return nil
}

// AddMessage adiciona mensagem à sessão
func (m *Manager) AddMessage(msg llm.Message) error {
	if m.currentSession == nil {
		return fmt.Errorf("no active session")
	}

	m.currentSession.Messages = append(m.currentSession.Messages, msg)
	m.currentSession.LastActivity = time.Now()

	return m.Save()
}

// GetCurrent retorna sessão atual
func (m *Manager) GetCurrent() *Session {
	return m.currentSession
}

// List lista todas as sessões
func (m *Manager) List(limit int) ([]*Session, error) {
	files, err := os.ReadDir(m.sessionDir)
	if err != nil {
		return nil, err
	}

	sessions := make([]*Session, 0)

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		session, err := m.loadSessionByFilename(file.Name())
		if err != nil {
			continue
		}

		sessions = append(sessions, session)
	}

	// Ordenar por atividade recente
	sort.Sort(SessionList(sessions))

	// Limitar resultados
	if limit > 0 && len(sessions) > limit {
		sessions = sessions[:limit]
	}

	return sessions, nil
}

// Get obtém sessão por ID
func (m *Manager) Get(sessionID string) (*Session, error) {
	return m.loadSession(sessionID)
}

// Delete remove sessão
func (m *Manager) Delete(sessionID string) error {
	path := m.sessionPath(sessionID)
	return os.Remove(path)
}

// UpdateMetadata atualiza metadata da sessão
func (m *Manager) UpdateMetadata(key string, value interface{}) error {
	if m.currentSession == nil {
		return fmt.Errorf("no active session")
	}

	m.currentSession.Metadata[key] = value
	return m.Save()
}

// AddTag adiciona tag à sessão
func (m *Manager) AddTag(tag string) error {
	if m.currentSession == nil {
		return fmt.Errorf("no active session")
	}

	// Verificar se já existe
	for _, t := range m.currentSession.Tags {
		if t == tag {
			return nil
		}
	}

	m.currentSession.Tags = append(m.currentSession.Tags, tag)
	return m.Save()
}

// saveSession persiste sessão
func (m *Manager) saveSession(session *Session) error {
	path := m.sessionPath(session.ID)

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// loadSession carrega sessão por ID
func (m *Manager) loadSession(sessionID string) (*Session, error) {
	path := m.sessionPath(sessionID)
	return m.loadSessionFromPath(path)
}

// loadSessionByFilename carrega por nome de arquivo
func (m *Manager) loadSessionByFilename(filename string) (*Session, error) {
	path := filepath.Join(m.sessionDir, filename)
	return m.loadSessionFromPath(path)
}

// loadSessionFromPath carrega de caminho
func (m *Manager) loadSessionFromPath(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// sessionPath retorna caminho da sessão
func (m *Manager) sessionPath(sessionID string) string {
	return filepath.Join(m.sessionDir, sessionID+".json")
}

// generateSessionID gera ID único
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// GetStats retorna estatísticas
func (m *Manager) GetStats() (map[string]interface{}, error) {
	sessions, err := m.List(0)
	if err != nil {
		return nil, err
	}

	activeCount := 0
	totalMessages := 0

	for _, s := range sessions {
		if s.Active {
			activeCount++
		}
		totalMessages += len(s.Messages)
	}

	return map[string]interface{}{
		"total_sessions":  len(sessions),
		"active_sessions": activeCount,
		"total_messages":  totalMessages,
		"current_session": m.currentSession != nil,
	}, nil
}
