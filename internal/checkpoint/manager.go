package checkpoint

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/johnpitter/ollama-code/internal/llm"
)

// Manager gerenciador de checkpoints
type Manager struct {
	checkpointDir  string
	retention      time.Duration
	maxCheckpoints int
}

// NewManager cria novo gerenciador
func NewManager(baseDir string) *Manager {
	checkpointDir := filepath.Join(baseDir, ".ollama-code", "checkpoints")
	os.MkdirAll(checkpointDir, 0755)

	return &Manager{
		checkpointDir:  checkpointDir,
		retention:      30 * 24 * time.Hour, // 30 dias
		maxCheckpoints: 100,
	}
}

// CreateCheckpoint cria checkpoint ANTES de cada edição
func (m *Manager) CreateCheckpoint(
	conversation []llm.Message,
	changedFiles []string,
	workDir string,
	description string,
	autoCreated bool,
) (*Checkpoint, error) {
	cp := &Checkpoint{
		ID:           generateID(),
		Timestamp:    time.Now(),
		Conversation: conversation,
		FileStates:   make(map[string]FileState),
		WorkspaceState: WorkspaceState{
			WorkingDir: workDir,
		},
		Description: description,
		AutoCreated: autoCreated,
	}

	// Salvar estado atual dos arquivos
	for _, filePath := range changedFiles {
		absPath := filepath.Join(workDir, filePath)

		content, err := os.ReadFile(absPath)
		if err != nil {
			// Arquivo pode não existir (será criado)
			continue
		}

		info, _ := os.Stat(absPath)

		cp.FileStates[filePath] = FileState{
			Path:         filePath,
			Content:      string(content),
			Hash:         hashContent(content),
			ModifiedTime: info.ModTime(),
			Size:         info.Size(),
		}
	}

	// Capturar estado do Git se disponível
	m.captureGitState(cp, workDir)

	// Persistir checkpoint
	if err := m.saveCheckpoint(cp); err != nil {
		return nil, err
	}

	// Limpar checkpoints antigos
	go m.CleanupOldCheckpoints()

	return cp, nil
}

// Rewind restaura para checkpoint anterior
func (m *Manager) Rewind(checkpointID string, restoreConversation, restoreFiles bool) (*Checkpoint, error) {
	cp, err := m.loadCheckpoint(checkpointID)
	if err != nil {
		return nil, fmt.Errorf("load checkpoint: %w", err)
	}

	// Restaurar arquivos
	if restoreFiles {
		for _, fileState := range cp.FileStates {
			absPath := filepath.Join(cp.WorkspaceState.WorkingDir, fileState.Path)

			// Criar diretórios se necessário
			dir := filepath.Dir(absPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("create directory: %w", err)
			}

			// Restaurar arquivo
			if err := os.WriteFile(absPath, []byte(fileState.Content), 0644); err != nil {
				return nil, fmt.Errorf("restore file %s: %w", fileState.Path, err)
			}
		}
	}

	return cp, nil
}

// List lista checkpoints disponíveis
func (m *Manager) List(limit int) ([]*Checkpoint, error) {
	files, err := os.ReadDir(m.checkpointDir)
	if err != nil {
		return nil, err
	}

	checkpoints := make([]*Checkpoint, 0)

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		cp, err := m.loadCheckpointByFilename(file.Name())
		if err != nil {
			continue
		}

		checkpoints = append(checkpoints, cp)
	}

	// Ordenar por timestamp (mais recente primeiro)
	sort.Sort(CheckpointList(checkpoints))

	// Limitar resultados
	if limit > 0 && len(checkpoints) > limit {
		checkpoints = checkpoints[:limit]
	}

	return checkpoints, nil
}

// Get obtém checkpoint por ID
func (m *Manager) Get(checkpointID string) (*Checkpoint, error) {
	return m.loadCheckpoint(checkpointID)
}

// Delete remove checkpoint
func (m *Manager) Delete(checkpointID string) error {
	path := m.checkpointPath(checkpointID)
	return os.Remove(path)
}

// CleanupOldCheckpoints limpa checkpoints antigos
func (m *Manager) CleanupOldCheckpoints() error {
	cutoff := time.Now().Add(-m.retention)

	checkpoints, err := m.List(0)
	if err != nil {
		return err
	}

	deletedCount := 0

	// Remover checkpoints expirados
	for _, cp := range checkpoints {
		if cp.Timestamp.Before(cutoff) {
			if err := m.Delete(cp.ID); err == nil {
				deletedCount++
			}
		}
	}

	// Limitar número total de checkpoints
	if len(checkpoints) > m.maxCheckpoints {
		// Remover os mais antigos (já estão ordenados)
		for i := m.maxCheckpoints; i < len(checkpoints); i++ {
			if checkpoints[i].AutoCreated { // Só remove auto-criados
				if err := m.Delete(checkpoints[i].ID); err == nil {
					deletedCount++
				}
			}
		}
	}

	return nil
}

// saveCheckpoint persiste checkpoint
func (m *Manager) saveCheckpoint(cp *Checkpoint) error {
	path := m.checkpointPath(cp.ID)

	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// loadCheckpoint carrega checkpoint por ID
func (m *Manager) loadCheckpoint(checkpointID string) (*Checkpoint, error) {
	path := m.checkpointPath(checkpointID)
	return m.loadCheckpointFromPath(path)
}

// loadCheckpointByFilename carrega por nome de arquivo
func (m *Manager) loadCheckpointByFilename(filename string) (*Checkpoint, error) {
	path := filepath.Join(m.checkpointDir, filename)
	return m.loadCheckpointFromPath(path)
}

// loadCheckpointFromPath carrega de caminho
func (m *Manager) loadCheckpointFromPath(path string) (*Checkpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return nil, err
	}

	return &cp, nil
}

// checkpointPath retorna caminho do checkpoint
func (m *Manager) checkpointPath(checkpointID string) string {
	return filepath.Join(m.checkpointDir, checkpointID+".json")
}

// captureGitState captura estado do Git
func (m *Manager) captureGitState(cp *Checkpoint, workDir string) {
	// TODO: Implementar captura real do estado Git
	// Usar exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	// e exec.Command("git", "rev-parse", "HEAD")
	_ = workDir
}

// generateID gera ID único
func generateID() string {
	return fmt.Sprintf("cp_%d", time.Now().UnixNano())
}

// hashContent calcula hash do conteúdo
func hashContent(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// GetStats retorna estatísticas
func (m *Manager) GetStats() (map[string]interface{}, error) {
	checkpoints, err := m.List(0)
	if err != nil {
		return nil, err
	}

	autoCount := 0
	manualCount := 0
	totalSize := int64(0)

	for _, cp := range checkpoints {
		if cp.AutoCreated {
			autoCount++
		} else {
			manualCount++
		}

		for _, fs := range cp.FileStates {
			totalSize += fs.Size
		}
	}

	return map[string]interface{}{
		"total_checkpoints": len(checkpoints),
		"auto_created":      autoCount,
		"manual_created":    manualCount,
		"total_size_bytes":  totalSize,
		"retention_days":    int(m.retention.Hours() / 24),
		"max_checkpoints":   m.maxCheckpoints,
	}, nil
}
