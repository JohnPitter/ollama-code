package checkpoint

import (
	"time"

	"github.com/johnpitter/ollama-code/internal/llm"
)

// Checkpoint representa um checkpoint do estado
type Checkpoint struct {
	ID              string                 `json:"id"`
	Timestamp       time.Time              `json:"timestamp"`
	Conversation    []llm.Message          `json:"conversation"`
	FileStates      map[string]FileState   `json:"file_states"`
	WorkspaceState  WorkspaceState         `json:"workspace_state"`
	Description     string                 `json:"description"`
	AutoCreated     bool                   `json:"auto_created"`
	Tags            []string               `json:"tags,omitempty"`
}

// FileState estado de um arquivo
type FileState struct {
	Path         string    `json:"path"`
	Content      string    `json:"content"`
	Hash         string    `json:"hash"`
	ModifiedTime time.Time `json:"modified_time"`
	Size         int64     `json:"size"`
}

// WorkspaceState estado do workspace
type WorkspaceState struct {
	WorkingDir   string            `json:"working_dir"`
	GitBranch    string            `json:"git_branch,omitempty"`
	GitCommit    string            `json:"git_commit,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	OpenFiles    []string          `json:"open_files,omitempty"`
}

// CheckpointList lista de checkpoints
type CheckpointList []*Checkpoint

// Len implementa sort.Interface
func (c CheckpointList) Len() int {
	return len(c)
}

// Less implementa sort.Interface
func (c CheckpointList) Less(i, j int) bool {
	return c[i].Timestamp.After(c[j].Timestamp)
}

// Swap implementa sort.Interface
func (c CheckpointList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
