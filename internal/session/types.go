package session

import (
	"time"

	"github.com/johnpitter/ollama-code/internal/llm"
)

// Session representa uma sessão de trabalho
type Session struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name,omitempty"`
	StartTime    time.Time              `json:"start_time"`
	LastActivity time.Time              `json:"last_activity"`
	Messages     []llm.Message          `json:"messages"`
	WorkDir      string                 `json:"work_dir"`
	Mode         string                 `json:"mode"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Active       bool                   `json:"active"`
}

// SessionList lista de sessões
type SessionList []*Session

func (s SessionList) Len() int {
	return len(s)
}

func (s SessionList) Less(i, j int) bool {
	return s[i].LastActivity.After(s[j].LastActivity)
}

func (s SessionList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
