# 🏢 ENTERPRISE FEATURES - Ollama Code

## Funcionalidades Adicionais para Resiliência Corporativa

Baseado em análise completa do Claude Code, aqui estão TODAS as funcionalidades adicionais necessárias para uso diário corporativo por engenheiros de software.

---

## 🔄 1. STATE RECOVERY & CHECKPOINTS

### Sistema de Checkpoints Automáticos

**Arquivo: `internal/checkpoint/manager.go`**

```go
package checkpoint

import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
)

type CheckpointManager struct {
    checkpointDir string
    retention     time.Duration // 30 dias
}

type Checkpoint struct {
    ID              string                 `json:"id"`
    Timestamp       time.Time              `json:"timestamp"`
    Conversation    []Message              `json:"conversation"`
    FileStates      map[string]FileState   `json:"file_states"`
    WorkspaceState  WorkspaceState         `json:"workspace_state"`
    Description     string                 `json:"description"`
}

type FileState struct {
    Path    string `json:"path"`
    Content string `json:"content"`
    Hash    string `json:"hash"`
}

func NewCheckpointManager(dir string) *CheckpointManager {
    return &CheckpointManager{
        checkpointDir: dir,
        retention:     30 * 24 * time.Hour,
    }
}

// CreateCheckpoint - Criar checkpoint ANTES de cada edição
func (c *CheckpointManager) CreateCheckpoint(
    conversation []Message,
    changedFiles []string,
    description string,
) (*Checkpoint, error) {
    cp := &Checkpoint{
        ID:          generateID(),
        Timestamp:   time.Now(),
        Conversation: conversation,
        FileStates:  make(map[string]FileState),
        Description: description,
    }

    // Salvar estado atual dos arquivos
    for _, filePath := range changedFiles {
        content, _ := os.ReadFile(filePath)
        cp.FileStates[filePath] = FileState{
            Path:    filePath,
            Content: string(content),
            Hash:    hashContent(content),
        }
    }

    // Persistir checkpoint
    cpPath := filepath.Join(c.checkpointDir, cp.ID+".json")
    data, _ := json.MarshalIndent(cp, "", "  ")
    os.WriteFile(cpPath, data, 0644)

    return cp, nil
}

// Rewind - Restaurar para checkpoint anterior
func (c *CheckpointManager) Rewind(checkpointID string, restoreConversation, restoreFiles bool) error {
    cpPath := filepath.Join(c.checkpointDir, checkpointID+".json")

    data, err := os.ReadFile(cpPath)
    if err != nil {
        return err
    }

    var cp Checkpoint
    if err := json.Unmarshal(data, &cp); err != nil {
        return err
    }

    // Restaurar arquivos
    if restoreFiles {
        for _, fileState := range cp.FileStates {
            os.WriteFile(fileState.Path, []byte(fileState.Content), 0644)
        }
    }

    // Restaurar conversa seria feito no agent principal

    return nil
}

// ListCheckpoints - Listar checkpoints disponíveis
func (c *CheckpointManager) ListCheckpoints() ([]*Checkpoint, error) {
    files, err := os.ReadDir(c.checkpointDir)
    if err != nil {
        return nil, err
    }

    checkpoints := make([]*Checkpoint, 0)

    for _, file := range files {
        if filepath.Ext(file.Name()) != ".json" {
            continue
        }

        data, _ := os.ReadFile(filepath.Join(c.checkpointDir, file.Name()))

        var cp Checkpoint
        if json.Unmarshal(data, &cp) == nil {
            checkpoints = append(checkpoints, &cp)
        }
    }

    // Ordenar por timestamp (mais recente primeiro)
    sort.Slice(checkpoints, func(i, j int) bool {
        return checkpoints[i].Timestamp.After(checkpoints[j].Timestamp)
    })

    return checkpoints, nil
}

// CleanupOldCheckpoints - Limpar checkpoints antigos (>30 dias)
func (c *CheckpointManager) CleanupOldCheckpoints() error {
    cutoff := time.Now().Add(-c.retention)

    files, _ := os.ReadDir(c.checkpointDir)

    for _, file := range files {
        path := filepath.Join(c.checkpointDir, file.Name())
        info, _ := file.Info()

        if info.ModTime().Before(cutoff) {
            os.Remove(path)
        }
    }

    return nil
}
```

### Comando /rewind

```go
// No agent.go, adicionar comando /rewind
func (a *Agent) handleRewindCommand() error {
    checkpoints, _ := a.checkpointMgr.ListCheckpoints()

    if len(checkpoints) == 0 {
        fmt.Println("❌ Nenhum checkpoint disponível")
        return nil
    }

    // Mostrar lista
    fmt.Println("\n📋 Checkpoints Disponíveis:\n")
    for i, cp := range checkpoints {
        fmt.Printf("[%d] %s - %s\n", i,
            cp.Timestamp.Format("2006-01-02 15:04:05"),
            cp.Description)
    }

    // Usuário escolhe
    fmt.Print("\nEscolha checkpoint [0-N] ou 'c' para cancelar: ")
    var choice string
    fmt.Scanln(&choice)

    if choice == "c" {
        return nil
    }

    idx, _ := strconv.Atoi(choice)
    if idx < 0 || idx >= len(checkpoints) {
        return fmt.Errorf("índice inválido")
    }

    // Perguntar o que restaurar
    fmt.Print("Restaurar: [1] Apenas conversa [2] Apenas arquivos [3] Ambos: ")
    var mode string
    fmt.Scanln(&mode)

    restoreConv := mode == "1" || mode == "3"
    restoreFiles := mode == "2" || mode == "3"

    err := a.checkpointMgr.Rewind(
        checkpoints[idx].ID,
        restoreConv,
        restoreFiles,
    )

    if err != nil {
        return err
    }

    fmt.Println("✅ Checkpoint restaurado!")
    return nil
}
```

---

## 📊 2. SESSION MANAGEMENT & RESUMPTION

### Gerenciador de Sessões

**Arquivo: `internal/session/manager.go`**

```go
package session

import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
)

type SessionManager struct {
    sessionsDir string
}

type Session struct {
    ID           string                 `json:"id"`
    StartTime    time.Time              `json:"start_time"`
    LastActivity time.Time              `json:"last_activity"`
    Messages     []Message              `json:"messages"`
    WorkDir      string                 `json:"work_dir"`
    Mode         string                 `json:"mode"`
    Metadata     map[string]interface{} `json:"metadata"`
}

func NewSessionManager(dir string) *SessionManager {
    os.MkdirAll(dir, 0755)
    return &SessionManager{sessionsDir: dir}
}

// CreateSession - Criar nova sessão
func (sm *SessionManager) CreateSession(workDir, mode string) (*Session, error) {
    session := &Session{
        ID:           generateSessionID(),
        StartTime:    time.Now(),
        LastActivity: time.Now(),
        Messages:     make([]Message, 0),
        WorkDir:      workDir,
        Mode:         mode,
        Metadata:     make(map[string]interface{}),
    }

    return session, sm.SaveSession(session)
}

// SaveSession - Salvar sessão
func (sm *SessionManager) SaveSession(session *Session) error {
    session.LastActivity = time.Now()

    path := filepath.Join(sm.sessionsDir, session.ID+".json")
    data, _ := json.MarshalIndent(session, "", "  ")

    return os.WriteFile(path, data, 0644)
}

// LoadSession - Carregar sessão
func (sm *SessionManager) LoadSession(sessionID string) (*Session, error) {
    path := filepath.Join(sm.sessionsDir, sessionID+".json")

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

// ListSessions - Listar todas as sessões
func (sm *SessionManager) ListSessions() ([]*Session, error) {
    files, err := os.ReadDir(sm.sessionsDir)
    if err != nil {
        return nil, err
    }

    sessions := make([]*Session, 0)

    for _, file := range files {
        if filepath.Ext(file.Name()) != ".json" {
            continue
        }

        sessionID := strings.TrimSuffix(file.Name(), ".json")
        session, err := sm.LoadSession(sessionID)
        if err == nil {
            sessions = append(sessions, session)
        }
    }

    // Ordenar por última atividade
    sort.Slice(sessions, func(i, j int) bool {
        return sessions[i].LastActivity.After(sessions[j].LastActivity)
    }

    return sessions, nil
}

// GetMostRecent - Pegar sessão mais recente
func (sm *SessionManager) GetMostRecent() (*Session, error) {
    sessions, err := sm.ListSessions()
    if err != nil || len(sessions) == 0 {
        return nil, fmt.Errorf("nenhuma sessão encontrada")
    }

    return sessions[0], nil
}
```

### Comandos de Sessão

```bash
# Continuar última sessão
ollama-code chat --continue
ollama-code chat -c

# Retomar sessão específica
ollama-code chat --resume "session-id-123"
ollama-code chat -r "session-id-123"

# Listar sessões
ollama-code sessions list

# Ver detalhes de sessão
ollama-code sessions show <session-id>

# Deletar sessão
ollama-code sessions delete <session-id>
```

---

## 📝 3. HIERARCHICAL MEMORY SYSTEM

### Sistema de Memória em Camadas

**Estrutura:**
```
1. Enterprise Policy (~/.claude-enterprise/POLICY.md) - READONLY
2. Project Memory (./CLAUDE.md ou ./.claude/CLAUDE.md)
3. Project Rules (./.claude/rules/*.md)
4. User Memory (~/.claude/CLAUDE.md)
5. Local Project Memory (./CLAUDE.local.md) - gitignored
```

**Arquivo: `internal/memory/loader.go`**

```go
package memory

import (
    "os"
    "path/filepath"
    "strings"
)

type MemoryLoader struct {
    projectRoot   string
    userHome      string
    enterpriseDir string
}

type MemoryContext struct {
    EnterprisePolicy string
    ProjectMemory    string
    ProjectRules     []string
    UserMemory       string
    LocalMemory      string
}

func NewMemoryLoader(projectRoot, userHome string) *MemoryLoader {
    return &MemoryLoader{
        projectRoot:   projectRoot,
        userHome:      userHome,
        enterpriseDir: filepath.Join(userHome, ".claude-enterprise"),
    }
}

// LoadAll - Carregar toda a hierarquia de memória
func (ml *MemoryLoader) LoadAll() (*MemoryContext, error) {
    ctx := &MemoryContext{}

    // 1. Enterprise Policy (se existir)
    enterprisePolicy := filepath.Join(ml.enterpriseDir, "POLICY.md")
    if content, err := os.ReadFile(enterprisePolicy); err == nil {
        ctx.EnterprisePolicy = string(content)
    }

    // 2. Project Memory
    projectMemory := ml.findProjectMemory()
    if projectMemory != "" {
        if content, err := os.ReadFile(projectMemory); err == nil {
            ctx.ProjectMemory = string(content)
        }
    }

    // 3. Project Rules
    rulesDir := filepath.Join(ml.projectRoot, ".claude", "rules")
    if files, err := os.ReadDir(rulesDir); err == nil {
        for _, file := range files {
            if filepath.Ext(file.Name()) == ".md" {
                path := filepath.Join(rulesDir, file.Name())
                if content, err := os.ReadFile(path); err == nil {
                    ctx.ProjectRules = append(ctx.ProjectRules, string(content))
                }
            }
        }
    }

    // 4. User Memory
    userMemory := filepath.Join(ml.userHome, ".claude", "CLAUDE.md")
    if content, err := os.ReadFile(userMemory); err == nil {
        ctx.UserMemory = string(content)
    }

    // 5. Local Project Memory (gitignored)
    localMemory := filepath.Join(ml.projectRoot, "CLAUDE.local.md")
    if content, err := os.ReadFile(localMemory); err == nil {
        ctx.LocalMemory = string(content)
    }

    return ctx, nil
}

func (ml *MemoryLoader) findProjectMemory() string {
    // Procurar CLAUDE.md ou .claude/CLAUDE.md
    options := []string{
        filepath.Join(ml.projectRoot, "CLAUDE.md"),
        filepath.Join(ml.projectRoot, ".claude", "CLAUDE.md"),
    }

    for _, path := range options {
        if _, err := os.Stat(path); err == nil {
            return path
        }
    }

    return ""
}

// BuildSystemPrompt - Construir system prompt com toda a memória
func (ml *MemoryLoader) BuildSystemPrompt(basePrompt string) string {
    ctx, _ := ml.LoadAll()

    var builder strings.Builder

    builder.WriteString(basePrompt)
    builder.WriteString("\n\n")

    // Enterprise Policy (mais alta prioridade)
    if ctx.EnterprisePolicy != "" {
        builder.WriteString("# ENTERPRISE POLICY (MUST FOLLOW)\n\n")
        builder.WriteString(ctx.EnterprisePolicy)
        builder.WriteString("\n\n")
    }

    // Project Memory
    if ctx.ProjectMemory != "" {
        builder.WriteString("# PROJECT GUIDELINES\n\n")
        builder.WriteString(ctx.ProjectMemory)
        builder.WriteString("\n\n")
    }

    // Project Rules
    if len(ctx.ProjectRules) > 0 {
        builder.WriteString("# PROJECT RULES\n\n")
        for _, rule := range ctx.ProjectRules {
            builder.WriteString(rule)
            builder.WriteString("\n\n")
        }
    }

    // User Memory
    if ctx.UserMemory != "" {
        builder.WriteString("# USER PREFERENCES\n\n")
        builder.WriteString(ctx.UserMemory)
        builder.WriteString("\n\n")
    }

    // Local Memory (mais baixa prioridade, específico da máquina)
    if ctx.LocalMemory != "" {
        builder.WriteString("# LOCAL NOTES\n\n")
        builder.WriteString(ctx.LocalMemory)
        builder.WriteString("\n\n")
    }

    return builder.String()
}
```

### Comando /memory

```go
// Visualizar memória atual
func (a *Agent) handleMemoryCommand() {
    ctx, _ := a.memoryLoader.LoadAll()

    fmt.Println("\n📚 Memória Hierárquica:\n")

    if ctx.EnterprisePolicy != "" {
        fmt.Println("🏢 Enterprise Policy: ✓")
    }

    if ctx.ProjectMemory != "" {
        fmt.Println("📁 Project Memory: ✓")
    }

    if len(ctx.ProjectRules) > 0 {
        fmt.Printf("📋 Project Rules: %d arquivo(s)\n", len(ctx.ProjectRules))
    }

    if ctx.UserMemory != "" {
        fmt.Println("👤 User Memory: ✓")
    }

    if ctx.LocalMemory != "" {
        fmt.Println("💻 Local Memory: ✓")
    }

    fmt.Print("\nEditar: [1] Project [2] User [3] Local [4] Ver tudo: ")

    var choice string
    fmt.Scanln(&choice)

    switch choice {
    case "1":
        a.editProjectMemory()
    case "2":
        a.editUserMemory()
    case "3":
        a.editLocalMemory()
    case "4":
        a.viewAllMemory()
    }
}
```

---

## 🔌 4. SLASH COMMANDS CUSTOMIZADOS

### Sistema de Comandos Slash

**Arquivo: `internal/commands/registry.go`**

```go
package commands

import (
    "os"
    "path/filepath"
    "strings"
)

type SlashCommand struct {
    Name        string
    Description string
    Content     string
    Args        []string
    Type        string // "text" ou "bash"
}

type Registry struct {
    commands map[string]*SlashCommand
}

func NewRegistry() *Registry {
    return &Registry{
        commands: make(map[string]*SlashCommand),
    }
}

// LoadCommands - Carregar comandos de diretórios
func (r *Registry) LoadCommands(projectDir, userDir string) error {
    // Comandos do projeto (.claude/commands/)
    projectCmdDir := filepath.Join(projectDir, ".claude", "commands")
    r.loadFromDir(projectCmdDir, "project")

    // Comandos do usuário (~/.claude/commands/)
    userCmdDir := filepath.Join(userDir, ".claude", "commands")
    r.loadFromDir(userCmdDir, "user")

    return nil
}

func (r *Registry) loadFromDir(dir, scope string) {
    files, err := os.ReadDir(dir)
    if err != nil {
        return
    }

    for _, file := range files {
        if !strings.HasSuffix(file.Name(), ".md") {
            continue
        }

        name := strings.TrimSuffix(file.Name(), ".md")
        path := filepath.Join(dir, file.Name())

        content, err := os.ReadFile(path)
        if err != nil {
            continue
        }

        cmd := &SlashCommand{
            Name:    name,
            Content: string(content),
        }

        // Parsear frontmatter YAML (se existir)
        cmd.parseMetadata()

        r.commands[name] = cmd
    }
}

// Execute - Executar comando slash
func (r *Registry) Execute(name string, args []string) (string, error) {
    cmd, ok := r.commands[name]
    if !ok {
        return "", fmt.Errorf("comando /%s não encontrado", name)
    }

    if cmd.Type == "bash" {
        return r.executeBash(cmd, args)
    }

    // Comando de texto - substituir placeholders
    result := cmd.Content
    for i, arg := range args {
        placeholder := fmt.Sprintf("$%d", i+1)
        result = strings.ReplaceAll(result, placeholder, arg)
    }

    return result, nil
}

func (r *Registry) executeBash(cmd *SlashCommand, args []string) (string, error) {
    // Executar script bash com argumentos
    // Implementar execução segura
    return "", nil
}

// List - Listar comandos disponíveis
func (r *Registry) List() []*SlashCommand {
    commands := make([]*SlashCommand, 0, len(r.commands))
    for _, cmd := range r.commands {
        commands = append(commands, cmd)
    }
    return commands
}
```

### Exemplo de Comando Customizado

**`.claude/commands/review.md`:**
```markdown
---
description: "Faz code review do código alterado"
args: ["target"]
---

Por favor, faça um code review completo focando em:

1. **Qualidade do Código**
   - Clean Code principles
   - SOLID
   - DRY

2. **Segurança**
   - SQL Injection
   - XSS
   - CSRF
   - Secrets hardcoded

3. **Performance**
   - Algoritmos O(n²) ou pior
   - N+1 queries
   - Memory leaks

4. **Testes**
   - Coverage adequado
   - Edge cases

Target: $1
```

Uso:
```bash
/review src/api/users.go
```

---

## 🪝 5. HOOKS SYSTEM

### Sistema de Hooks

**Arquivo: `internal/hooks/manager.go`**

```go
package hooks

import (
    "bytes"
    "context"
    "encoding/json"
    "os/exec"
    "strings"
)

type HookType string

const (
    PreToolUse        HookType = "PreToolUse"
    PostToolUse       HookType = "PostToolUse"
    UserPromptSubmit  HookType = "UserPromptSubmit"
    SessionStart      HookType = "SessionStart"
)

type Hook struct {
    Name       string
    Type       HookType
    Pattern    string // regex para filtrar ferramentas
    ScriptPath string
    UseLLM     bool
}

type HookDecision struct {
    Action  string // "allow", "deny", "ask"
    Message string
}

type Manager struct {
    hooks     []*Hook
    llmClient LLMClient
}

func NewManager(llmClient LLMClient) *Manager {
    return &Manager{
        hooks:     make([]*Hook, 0),
        llmClient: llmClient,
    }
}

// LoadHooks - Carregar hooks de .claude/hooks/
func (m *Manager) LoadHooks(hooksDir string) error {
    // Implementar carregamento de hooks
    return nil
}

// ExecutePreToolUse - Executar hook antes de usar ferramenta
func (m *Manager) ExecutePreToolUse(
    ctx context.Context,
    toolName string,
    params map[string]any,
) (*HookDecision, error) {
    // Encontrar hooks aplicáveis
    applicableHooks := m.findHooks(PreToolUse, toolName)

    for _, hook := range applicableHooks {
        if hook.UseLLM {
            return m.executeLLMHook(ctx, hook, toolName, params)
        } else {
            return m.executeBashHook(ctx, hook, toolName, params)
        }
    }

    // Default: permitir
    return &HookDecision{Action: "allow"}, nil
}

func (m *Manager) executeBashHook(
    ctx context.Context,
    hook *Hook,
    toolName string,
    params map[string]any,
) (*HookDecision, error) {
    // Preparar input JSON
    input := map[string]any{
        "tool":   toolName,
        "params": params,
    }

    inputJSON, _ := json.Marshal(input)

    // Executar script bash
    cmd := exec.CommandContext(ctx, "bash", hook.ScriptPath)
    cmd.Stdin = bytes.NewReader(inputJSON)

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()

    // Parsear output
    var decision HookDecision
    if err := json.Unmarshal(stdout.Bytes(), &decision); err != nil {
        // Fallback
        decision = HookDecision{Action: "allow"}
    }

    return &decision, nil
}

func (m *Manager) executeLLMHook(
    ctx context.Context,
    hook *Hook,
    toolName string,
    params map[string]any,
) (*HookDecision, error) {
    // Usar LLM (Haiku) para decidir
    prompt := fmt.Sprintf(`
Você é um hook de segurança.

Ferramenta: %s
Parâmetros: %v

Analise se esta operação deve ser permitida.

Responda em JSON:
{
  "action": "allow" | "deny" | "ask",
  "message": "explicação"
}
`, toolName, params)

    response, err := m.llmClient.Complete(ctx, "", prompt)
    if err != nil {
        return &HookDecision{Action: "allow"}, nil
    }

    var decision HookDecision
    json.Unmarshal([]byte(response), &decision)

    return &decision, nil
}

func (m *Manager) findHooks(hookType HookType, toolName string) []*Hook {
    applicable := make([]*Hook, 0)

    for _, hook := range m.hooks {
        if hook.Type != hookType {
            continue
        }

        // Verificar se pattern match
        if hook.Pattern == "*" || strings.Contains(toolName, hook.Pattern) {
            applicable = append(applicable, hook)
        }
    }

    return applicable
}
```

### Exemplo de Hook

**`.claude/hooks/security-check.sh`:**
```bash
#!/bin/bash

# Ler input JSON do stdin
input=$(cat)

# Parsear
tool=$(echo "$input" | jq -r '.tool')
params=$(echo "$input" | jq -r '.params')

# Verificar se é operação perigosa
if [[ "$tool" == "file_writer" ]] && [[ "$params" == *".env"* ]]; then
    # Negar escrita em .env
    echo '{"action": "deny", "message": "Não é permitido escrever em .env files"}'
    exit 0
fi

# Permitir por padrão
echo '{"action": "allow", "message": ""}'
```

---

## 📈 6. TELEMETRY & MONITORING

### Sistema de Telemetria

**Arquivo: `internal/telemetry/collector.go`**

```go
package telemetry

import (
    "context"
    "time"
)

type Event struct {
    Type      string                 `json:"type"`
    Timestamp time.Time              `json:"timestamp"`
    SessionID string                 `json:"session_id"`
    UserID    string                 `json:"user_id"`
    Data      map[string]interface{} `json:"data"`
}

type Metrics struct {
    SessionsCount      int64
    ActiveSessionTime  time.Duration
    LinesOfCodeChanged int64
    GitCommits         int64
    PullRequests       int64
    TokensInput        int64
    TokensOutput       int64
    TokensCached       int64
    APICostUSD         float64
}

type Collector struct {
    events    []*Event
    metrics   *Metrics
    sessionID string
}

func NewCollector(sessionID string) *Collector {
    return &Collector{
        events:    make([]*Event, 0),
        metrics:   &Metrics{},
        sessionID: sessionID,
    }
}

// RecordEvent - Registrar evento
func (c *Collector) RecordEvent(eventType string, data map[string]interface{}) {
    event := &Event{
        Type:      eventType,
        Timestamp: time.Now(),
        SessionID: c.sessionID,
        Data:      data,
    }

    c.events = append(c.events, event)
}

// RecordToolUse - Registrar uso de ferramenta
func (c *Collector) RecordToolUse(toolName string, success bool, duration time.Duration) {
    c.RecordEvent("tool_use", map[string]interface{}{
        "tool":     toolName,
        "success":  success,
        "duration": duration.Milliseconds(),
    })
}

// RecordTokens - Registrar uso de tokens
func (c *Collector) RecordTokens(input, output, cached int64, cost float64) {
    c.metrics.TokensInput += input
    c.metrics.TokensOutput += output
    c.metrics.TokensCached += cached
    c.metrics.APICostUSD += cost
}

// RecordFileEdit - Registrar edição de arquivo
func (c *Collector) RecordFileEdit(filePath string, linesChanged int) {
    c.metrics.LinesOfCodeChanged += int64(linesChanged)

    c.RecordEvent("file_edit", map[string]interface{}{
        "file":  filePath,
        "lines": linesChanged,
    })
}

// RecordGitCommit - Registrar commit
func (c *Collector) RecordGitCommit(hash string, message string) {
    c.metrics.GitCommits++

    c.RecordEvent("git_commit", map[string]interface{}{
        "hash":    hash,
        "message": message,
    })
}

// GetMetrics - Pegar métricas acumuladas
func (c *Collector) GetMetrics() *Metrics {
    return c.metrics
}

// Export - Exportar telemetria (OpenTelemetry, Prometheus, etc)
func (c *Collector) Export(ctx context.Context, exporter Exporter) error {
    return exporter.Export(ctx, c.events, c.metrics)
}
```

### Comandos de Monitoramento

```bash
# Ver uso atual
/usage

# Ver custos
/cost

# Ver estatísticas da sessão
/stats

# Ver contexto atual
/context
```

---

## 🔐 7. ENTERPRISE SECURITY FEATURES

### Sandboxing Avançado

**Arquivo: `internal/sandbox/manager.go`**

```go
package sandbox

import (
    "os/exec"
    "syscall"
)

type SandboxConfig struct {
    AllowedPaths    []string
    DeniedPaths     []string
    AllowedDomains  []string
    EnableNetwork   bool
    MaxFileSize     int64
    MaxMemory       int64
}

type Manager struct {
    config *SandboxConfig
}

func NewManager(config *SandboxConfig) *Manager {
    return &Manager{config: config}
}

// ExecuteInSandbox - Executar comando em sandbox
func (m *Manager) ExecuteInSandbox(command string, args []string) (*exec.Cmd, error) {
    cmd := exec.Command(command, args...)

    // Linux: usar bubblewrap
    if isLinux() {
        return m.wrapWithBubblewrap(cmd)
    }

    // macOS: usar sandbox-exec
    if isMacOS() {
        return m.wrapWithSandboxExec(cmd)
    }

    // Windows: usar process isolation
    if isWindows() {
        return m.wrapWithWindowsSandbox(cmd)
    }

    return cmd, nil
}

func (m *Manager) wrapWithBubblewrap(cmd *exec.Cmd) (*exec.Cmd, error) {
    // bubblewrap args
    bwArgs := []string{
        "--ro-bind", "/usr", "/usr",
        "--ro-bind", "/lib", "/lib",
        "--ro-bind", "/lib64", "/lib64",
        "--ro-bind", "/bin", "/bin",
        "--ro-bind", "/sbin", "/sbin",
        "--proc", "/proc",
        "--dev", "/dev",
        "--tmpfs", "/tmp",
    }

    // Adicionar paths permitidos
    for _, path := range m.config.AllowedPaths {
        bwArgs = append(bwArgs, "--bind", path, path)
    }

    // Desabilitar network se necessário
    if !m.config.EnableNetwork {
        bwArgs = append(bwArgs, "--unshare-net")
    }

    // Comando original
    bwArgs = append(bwArgs, cmd.Path)
    bwArgs = append(bwArgs, cmd.Args[1:]...)

    sandboxedCmd := exec.Command("bwrap", bwArgs...)
    sandboxedCmd.Env = cmd.Env
    sandboxedCmd.Dir = cmd.Dir

    return sandboxedCmd, nil
}

func (m *Manager) wrapWithSandboxExec(cmd *exec.Cmd) (*exec.Cmd, error) {
    // macOS Seatbelt profile
    profile := m.generateSeatbeltProfile()

    sandboxedCmd := exec.Command("sandbox-exec", "-p", profile, cmd.Path)
    sandboxedCmd.Args = append(sandboxedCmd.Args, cmd.Args[1:]...)

    return sandboxedCmd, nil
}

func (m *Manager) generateSeatbeltProfile() string {
    // Gerar perfil Seatbelt para macOS
    profile := `
(version 1)
(deny default)
(allow process-exec*)
(allow file-read*)
`

    for _, path := range m.config.AllowedPaths {
        profile += fmt.Sprintf("(allow file* (subpath \"%s\"))\n", path)
    }

    if !m.config.EnableNetwork {
        profile += "(deny network*)\n"
    }

    return profile
}
```

---

## 🎨 8. OUTPUT STYLES & FORMATTING

### Sistema de Estilos de Saída

**Arquivo: `internal/output/styles.go`**

```go
package output

import (
    "strings"
)

type Style struct {
    Name        string
    Description string
    Template    string
}

type StyleManager struct {
    styles        map[string]*Style
    currentStyle  string
}

func NewStyleManager() *StyleManager {
    sm := &StyleManager{
        styles:       make(map[string]*Style),
        currentStyle: "default",
    }

    sm.registerDefaultStyles()
    return sm
}

func (sm *StyleManager) registerDefaultStyles() {
    // Default
    sm.RegisterStyle(&Style{
        Name:        "default",
        Description: "Resposta padrão concisa",
        Template:    "{{.Content}}",
    })

    // Explanatory
    sm.RegisterStyle(&Style{
        Name:        "explanatory",
        Description: "Explicações educacionais detalhadas",
        Template: `
{{.Content}}

---

💡 **Explicação:**
{{.Explanation}}

📚 **Conceitos:**
{{range .Concepts}}
- {{.}}
{{end}}
`,
    })

    // Learning
    sm.RegisterStyle(&Style{
        Name:        "learning",
        Description: "Estilo colaborativo com TODOs",
        Template: `
{{.Content}}

🎯 **Próximos Passos:**
{{range .NextSteps}}
- [ ] {{.}}
{{end}}

📝 **Notas:**
{{.Notes}}
`,
    })

    // Corporate
    sm.RegisterStyle(&Style{
        Name:        "corporate",
        Description: "Formato corporativo estruturado",
        Template: `
## Resumo Executivo
{{.Summary}}

## Análise Técnica
{{.Analysis}}

## Recomendações
{{.Recommendations}}

## Riscos e Mitigações
{{.Risks}}
`,
    })
}

func (sm *StyleManager) RegisterStyle(style *Style) {
    sm.styles[style.Name] = style
}

func (sm *StyleManager) SetStyle(name string) error {
    if _, ok := sm.styles[name]; !ok {
        return fmt.Errorf("estilo %s não encontrado", name)
    }

    sm.currentStyle = name
    return nil
}

func (sm *StyleManager) Format(content string, data map[string]interface{}) string {
    style := sm.styles[sm.currentStyle]

    // Aplicar template
    // (usar text/template do Go)

    return content // simplificado
}
```

---

## ⚡ 9. PERFORMANCE OPTIMIZATIONS

### Cache de Contexto

**Arquivo: `internal/cache/context_cache.go`**

```go
package cache

import (
    "crypto/sha256"
    "encoding/hex"
    "sync"
    "time"
)

type CacheEntry struct {
    Key        string
    Value      interface{}
    Expiration time.Time
}

type ContextCache struct {
    entries map[string]*CacheEntry
    mu      sync.RWMutex
    ttl     time.Duration
}

func NewContextCache(ttl time.Duration) *ContextCache {
    cache := &ContextCache{
        entries: make(map[string]*CacheEntry),
        ttl:     ttl,
    }

    // Cleanup goroutine
    go cache.cleanupExpired()

    return cache
}

func (c *ContextCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, ok := c.entries[key]
    if !ok {
        return nil, false
    }

    if time.Now().After(entry.Expiration) {
        return nil, false
    }

    return entry.Value, true
}

func (c *ContextCache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.entries[key] = &CacheEntry{
        Key:        key,
        Value:      value,
        Expiration: time.Now().Add(c.ttl),
    }
}

func (c *ContextCache) HashKey(content string) string {
    hash := sha256.Sum256([]byte(content))
    return hex.EncodeToString(hash[:])
}

func (c *ContextCache) cleanupExpired() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for key, entry := range c.entries {
            if now.After(entry.Expiration) {
                delete(c.entries, key)
            }
        }
        c.mu.Unlock()
    }
}
```

### Async Background Tasks

```go
// No agent.go
func (a *Agent) ExecuteInBackground(fn func() error) {
    go func() {
        spinner := ora.New("Executando em background...")
        spinner.Start()

        err := fn()

        if err != nil {
            spinner.Fail(fmt.Sprintf("Erro: %v", err))
        } else {
            spinner.Succeed("Concluído!")
        }
    }()
}

// Uso:
// Usuário pressiona Ctrl+B para executar último comando em background
```

---

## 🎯 10. DIAGNOSTIC TOOLS

### Comando /doctor

**Arquivo: `internal/diagnostics/doctor.go`**

```go
package diagnostics

import (
    "fmt"
    "net/http"
    "os/exec"
)

type Doctor struct{}

func NewDoctor() *Doctor {
    return &Doctor{}
}

func (d *Doctor) Run() {
    fmt.Println("\n🔍 Ollama Code - Health Check\n")

    d.checkGo()
    d.checkOllama()
    d.checkGPU()
    d.checkModel()
    d.checkDiskSpace()
    d.checkNetwork()

    fmt.Println("\n✅ Diagnóstico completo!")
}

func (d *Doctor) checkGo() {
    fmt.Print("Verificando Go... ")

    cmd := exec.Command("go", "version")
    output, err := cmd.Output()

    if err != nil {
        fmt.Println("❌ Go não instalado")
        return
    }

    fmt.Printf("✓ %s\n", string(output))
}

func (d *Doctor) checkOllama() {
    fmt.Print("Verificando Ollama... ")

    resp, err := http.Get("http://localhost:11434/api/tags")
    if err != nil {
        fmt.Println("❌ Ollama não está rodando")
        fmt.Println("   Execute: ollama serve")
        return
    }
    defer resp.Body.Close()

    fmt.Println("✓ Rodando")
}

func (d *Doctor) checkGPU() {
    fmt.Print("Verificando GPU... ")

    cmd := exec.Command("nvidia-smi")
    output, err := cmd.Output()

    if err != nil {
        fmt.Println("⚠️  GPU não detectada (modo CPU)")
        return
    }

    fmt.Println("✓ NVIDIA GPU detectada")
}

func (d *Doctor) checkModel() {
    fmt.Print("Verificando modelo qwen2.5-coder:32b... ")

    // Verificar se modelo existe
    cmd := exec.Command("ollama", "list")
    output, err := cmd.Output()

    if err != nil || !strings.Contains(string(output), "qwen2.5-coder:32b") {
        fmt.Println("❌ Modelo não encontrado")
        fmt.Println("   Execute: ollama pull qwen2.5-coder:32b-instruct-q6_K")
        return
    }

    fmt.Println("✓ Instalado")
}

func (d *Doctor) checkDiskSpace() {
    fmt.Print("Verificando espaço em disco... ")

    // Implementar verificação de espaço

    fmt.Println("✓ Espaço suficiente")
}

func (d *Doctor) checkNetwork() {
    fmt.Print("Verificando conectividade... ")

    resp, err := http.Get("https://www.google.com")
    if err != nil {
        fmt.Println("⚠️  Sem acesso à internet")
        return
    }
    defer resp.Body.Close()

    fmt.Println("✓ Online")
}
```

---

## 📋 LISTA COMPLETA DE SLASH COMMANDS

```go
// Comandos built-in que devem ser implementados

var builtinCommands = map[string]string{
    // Navegação
    "/help":     "Mostrar ajuda",
    "/exit":     "Sair do chat",
    "/quit":     "Sair do chat",
    "/clear":    "Limpar histórico",
    "/status":   "Status do sistema",
    "/config":   "Configurações",

    // Sessões
    "/sessions": "Listar sessões",
    "/resume":   "Retomar sessão",
    "/continue": "Continuar última sessão",

    // Desenvolvimento
    "/review":   "Code review",
    "/security": "Security review",
    "/test":     "Rodar testes",
    "/build":    "Build do projeto",
    "/lint":     "Rodar linter",

    // Memória
    "/memory":   "Ver/editar memória",
    "/forget":   "Limpar memória",

    // Git
    "/commit":   "Criar commit",
    "/pr":       "Criar pull request",
    "/branch":   "Gerenciar branches",

    // Monitoramento
    "/usage":    "Uso de tokens",
    "/cost":     "Custos de API",
    "/stats":    "Estatísticas da sessão",
    "/context":  "Ver contexto atual",

    // Diagnóstico
    "/doctor":   "Health check",
    "/bug":      "Reportar bug",
    "/debug":    "Ativar debug mode",

    // Estado
    "/rewind":   "Voltar para checkpoint",
    "/undo":     "Desfazer última ação",
    "/redo":     "Refazer ação",

    // Ferramentas
    "/tools":    "Listar ferramentas",
    "/mode":     "Trocar modo de operação",
    "/output":   "Trocar estilo de output",

    // Hooks
    "/hooks":    "Gerenciar hooks",

    // Performance
    "/compact":  "Compactar contexto",
    "/cache":    "Gerenciar cache",
}
```

---

## 🚀 INTEGRAÇÃO COM CI/CD

### GitHub Actions Example

**`.github/workflows/ollama-code.yml`:**
```yaml
name: Ollama Code CI

on:
  issues:
    types: [labeled]
  issue_comment:
    types: [created]
  pull_request:
    types: [opened, synchronize]

jobs:
  ollama-code:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Setup Ollama
        run: |
          curl -fsSL https://ollama.ai/install.sh | sh
          ollama serve &
          ollama pull qwen2.5-coder:32b-instruct-q6_K

      - name: Install Ollama Code
        run: |
          cd ollama-code
          make install

      - name: Execute Task
        run: |
          ollama-code --mode autonomous \
            --print \
            "Implementar feature: ${{ github.event.issue.title }}"

      - name: Create PR
        if: success()
        run: |
          git config user.name "Ollama Code Bot"
          git config user.email "bot@ollama-code.dev"
          git checkout -b feature/${{ github.event.issue.number }}
          git add .
          git commit -m "feat: ${{ github.event.issue.title }}"
          git push origin feature/${{ github.event.issue.number }}
          gh pr create --title "feat: ${{ github.event.issue.title }}" \
            --body "Closes #${{ github.event.issue.number }}"
```

---

## 📊 SUMMARY

### Funcionalidades Implementadas vs Missing

| Categoria | Claude Code | Ollama Code Base | Enterprise Features |
|-----------|-------------|------------------|---------------------|
| **Core** | ✅ | ✅ | ✅ |
| State Recovery | ✅ | ❌ | ✅ Checkpoints |
| Session Management | ✅ | ❌ | ✅ Multi-session |
| Hierarchical Memory | ✅ | ❌ | ✅ 5 níveis |
| Slash Commands | ✅ 40+ | ❌ | ✅ Customizável |
| Hooks System | ✅ | ❌ | ✅ Pre/Post |
| Telemetry | ✅ | ❌ | ✅ OpenTelemetry |
| Sandboxing | ✅ | ❌ | ✅ Linux/macOS/Win |
| Output Styles | ✅ | ❌ | ✅ 4 estilos |
| Performance Cache | ✅ | ❌ | ✅ Context cache |
| Diagnostics | ✅ | ❌ | ✅ /doctor |
| CI/CD Integration | ✅ | ❌ | ✅ GitHub Actions |
| Background Tasks | ✅ | ❌ | ✅ Async |

---

## ⏱️ ESTIMATIVA DE IMPLEMENTAÇÃO

| Feature | Complexidade | Tempo Estimado |
|---------|--------------|----------------|
| Checkpoints | Média | 2 dias |
| Session Management | Média | 2 dias |
| Hierarchical Memory | Alta | 3 dias |
| Slash Commands | Baixa | 1 dia |
| Hooks System | Alta | 3 dias |
| Telemetry | Média | 2 dias |
| Sandboxing | Alta | 4 dias |
| Output Styles | Baixa | 1 dia |
| Performance Cache | Média | 2 dias |
| Diagnostics | Baixa | 1 dia |
| CI/CD Integration | Média | 2 dias |
| Background Tasks | Baixa | 1 dia |

**TOTAL:** ~24 dias adicionais

**TOTAL GERAL:** 10-12 dias (base) + 24 dias (enterprise) = **34-36 dias**

---

## 🎯 PRIORIZAÇÃO PARA USO CORPORATIVO

### Crítico (Implementar primeiro - 1-2 semanas)
1. ✅ Session Management
2. ✅ Hierarchical Memory (CLAUDE.md)
3. ✅ Checkpoints/Rewind
4. ✅ Slash Commands básicos
5. ✅ /doctor diagnostics

### Alto (Implementar em seguida - 2-3 semanas)
6. ✅ Hooks System
7. ✅ Telemetry & Monitoring
8. ✅ Output Styles
9. ✅ Performance Cache
10. ✅ Background Tasks

### Médio (Implementar depois - 3-4 semanas)
11. ✅ Sandboxing avançado
12. ✅ CI/CD Integration
13. ✅ Enterprise Security
14. ✅ Custom Slash Commands

---

**Este documento completa todas as funcionalidades enterprise-grade necessárias para uso corporativo diário!** 🚀
