# 🗺️ Roadmap - Ollama Code

**Objetivo:** Alcançar paridade funcional com Claude Code CLI oficial mantendo as vantagens de ser 100% local e gratuito.

**Última Atualização:** 30 de Dezembro de 2024

---

## 📊 Status Atual vs Claude Code CLI

| Categoria | Paridade | Comentário |
|-----------|----------|------------|
| **Core Features** | ✅ 95% | File ops, code search, git, web search |
| **Advanced Tools** | ✅ 90% | 15 tools, profiler, scanner |
| **Extensibility (MCP)** | ❌ 0% | **MAIOR GAP** |
| **User Interaction** | ⚠️ 60% | Falta TODO tracking, AskUserQuestion avançado |
| **Multimodal** | ❌ 0% | Limitação do LLM (aguardando Ollama) |
| **Subagents** | ❌ 0% | Task tool não implementado |
| **Observability** | ✅ 100% | Logging, metrics, tracing |

**Paridade Geral:** ⚠️ **70%**

**Para chegar a 95%:** Implementar MCP + TODO Tracking + Enhanced User Interaction

---

## 🎯 Visão de Longo Prazo (6-12 meses)

**Missão:** Ser o melhor coding assistant 100% local, com paridade funcional ao Claude Code CLI e vantagens únicas em privacidade e customização.

**Diferenciadores:**
1. ✅ **100% Local** - Zero dependência de APIs externas
2. ✅ **Gratuito** - Sem custos recorrentes
3. ✅ **Privacidade Total** - Código nunca sai da máquina
4. 🔄 **Extensível via MCP** - Ecosystem de plugins (roadmap)
5. 🔄 **Multi-Model** - Suporte a diferentes modelos Ollama (roadmap)
6. ✅ **Observability First** - Debugging e performance tracking avançados

---

## 📅 Fases de Desenvolvimento

### 🟢 Fase 0: Estabilização (COMPLETO ✅)
**Status:** ✅ Concluído (30/12/2024)
**Duração:** 2 semanas

#### Entregas
- [x] Handler Pattern implementation
- [x] Manual Dependency Injection
- [x] Observability System (logging, metrics, tracing)
- [x] Bug fixes críticos (read-only, code search, multi-file)
- [x] Testes de regressão automatizados (6 testes)
- [x] CHANGELOG.md e documentação
- [x] Performance troubleshooting docs

#### Métricas Alcançadas
- ✅ 100% testes de regressão passando (6/6)
- ✅ 0 bugs críticos
- ✅ Build estável

---

### 🟡 Fase 1: Quick Wins - User Experience (1-2 semanas)
**Status:** 🔄 EM PROGRESSO
**Prioridade:** 🔴 ALTA
**Início:** 30/12/2024

#### Objetivo
Melhorar drasticamente a UX em tarefas complexas com TODO tracking e interação avançada.

#### Features

##### 1.1 TODO Tracking System 🎯
**Prioridade:** 🔴 CRÍTICA
**Esforço:** 1 semana
**Inspiração:** TodoWrite tool do Claude Code CLI
**Status:** ✅ COMPLETO (30/12/2024)

**Funcionalidades:**
- [x] CRUD de TODOs em memória
- [x] Estados: pending, in_progress, completed
- [x] Formato: {content, status, activeForm}
- [x] Persistência opcional em JSON file
- [x] API compatível com Claude Code TodoWrite
- [x] Integração com handlers (auto-update de status)

**Arquivos a Criar:**
```
internal/todos/
├── manager.go      # TODO manager
├── types.go        # TODO types
├── storage.go      # Persistência
└── manager_test.go # Testes
```

**API Exemplo:**
```go
type TodoManager interface {
    Add(content, activeForm string) error
    Update(id string, status TodoStatus) error
    Complete(id string) error
    List() []Todo
    Clear() error
}
```

**Integração:**
```go
// Em file_write_handler.go
deps.TodoManager.Update("write-file-1", StatusInProgress)
// ... executa operação ...
deps.TodoManager.Complete("write-file-1")
```

**Métricas de Sucesso:**
- [ ] 100% dos handlers integrados com TODOs
- [ ] Testes unitários >90% coverage
- [ ] QA manual com tarefas multi-step

---

##### 1.2 Enhanced User Interaction 💬
**Prioridade:** 🟡 ALTA
**Esforço:** 1 semana
**Inspiração:** AskUserQuestion tool do Claude Code CLI
**Status:** ✅ COMPLETO (30/12/2024)

**Funcionalidades:**
- [x] Perguntas com múltiplas opções
- [x] Suporte a multiselect
- [x] Headers e descrições por opção
- [x] Validação de respostas
- [x] Fallback para input customizado

**Arquivos a Modificar:**
```
internal/confirmation/
├── manager.go      # Adicionar AskQuestion()
├── types.go        # Question, Option types
└── manager_test.go # Testes
```

**API Exemplo:**
```go
type Question struct {
    Question    string
    Header      string
    Options     []Option
    MultiSelect bool
}

type Option struct {
    Label       string
    Description string
}

func (m *Manager) AskQuestion(q Question) ([]string, error)
```

**Casos de Uso:**
```go
// Escolher biblioteca
response, _ := deps.ConfirmManager.AskQuestion(Question{
    Question: "Which frontend framework?",
    Header: "Framework",
    Options: []Option{
        {Label: "React", Description: "Component-based"},
        {Label: "Vue", Description: "Progressive framework"},
        {Label: "Svelte", Description: "Compiled framework"},
    },
    MultiSelect: false,
})
```

**Métricas de Sucesso:**
- [ ] API implementada e documentada
- [ ] Integração em 3+ handlers
- [ ] UX fluida em modo interativo

---

##### 1.3 Better Diff/Edit Operations 🔧
**Prioridade:** 🟢 MÉDIA
**Esforço:** 3 dias

**Funcionalidades:**
- [ ] Edit com ranges de linha (start:end)
- [ ] Preview de mudanças antes de aplicar
- [ ] Rollback de edições
- [ ] Diff colorizado no output

**Arquivos a Criar:**
```
internal/diff/
├── differ.go       # Diff engine
├── preview.go      # Preview de mudanças
└── differ_test.go  # Testes
```

**Métricas de Sucesso:**
- [ ] Edit tool com preview funcionando
- [ ] Rollback implementado
- [ ] Testes E2E

---

#### Entregáveis da Fase 1
- [ ] TODO Tracking System completo
- [ ] Enhanced User Interaction (AskQuestion)
- [ ] Better Diff/Edit operations
- [ ] Testes unitários (>85% coverage)
- [ ] Documentação atualizada
- [ ] Testes E2E

**ROI:** ⭐⭐⭐⭐⭐ (Muito Alto)
**Impacto no Usuário:** 🚀 Transformacional

---

### 🔴 Fase 2: MCP Protocol Support (3-4 semanas)
**Status:** 📋 PLANEJADA
**Prioridade:** 🔴 CRÍTICA
**Início Previsto:** Fevereiro 2025

#### Objetivo
Implementar suporte completo ao Model Context Protocol (MCP) para extensibilidade via plugins.

#### Background
MCP é o protocolo que permite ao Claude Code CLI usar ferramentas externas via `mcp__*` prefix. É o **maior gap** funcional atualmente.

#### Funcionalidades

##### 2.1 MCP Client Implementation 🔌
**Prioridade:** 🔴 CRÍTICA
**Esforço:** 2 semanas

**Requisitos:**
- [ ] Implementar MCP protocol client
- [ ] Discovery de MCP servers
- [ ] Conexão via stdio, SSE, ou WebSocket
- [ ] Message passing (request/response)
- [ ] Error handling robusto
- [ ] Timeout e retry logic

**Arquivos a Criar:**
```
internal/mcp/
├── client.go           # MCP client
├── protocol.go         # Protocol types
├── transport.go        # Transport layer (stdio, sse, websocket)
├── discovery.go        # Server discovery
├── registry.go         # MCP tool registry
├── adapters.go         # Adapter MCP tools → internal tools
└── client_test.go      # Testes
```

**Spec de Referência:**
- https://modelcontextprotocol.io/docs/
- https://github.com/modelcontextprotocol/servers

**Arquitetura:**
```
┌─────────────────┐
│  Ollama Code    │
│   (client)      │
└────────┬────────┘
         │ MCP Protocol
         │ (stdio/sse/ws)
         │
┌────────▼────────┐
│  MCP Server     │
│  (filesystem,   │
│   github, etc)  │
└─────────────────┘
```

---

##### 2.2 MCP Tool Registry 📦
**Prioridade:** 🔴 CRÍTICA
**Esforço:** 1 semana

**Funcionalidades:**
- [ ] Auto-discovery de MCP servers
- [ ] Registro dinâmico de tools
- [ ] Namespace: mcp__server_tool
- [ ] Listing de MCP tools
- [ ] Schema validation

**API Exemplo:**
```go
type MCPRegistry interface {
    DiscoverServers() ([]MCPServer, error)
    RegisterServer(server MCPServer) error
    ListTools() []MCPTool
    Execute(tool string, params map[string]interface{}) (MCPResult, error)
}
```

**Integração:**
```go
// No ToolRegistry
func (r *Registry) Execute(ctx, tool, params) {
    if strings.HasPrefix(tool, "mcp__") {
        return r.mcpRegistry.Execute(tool, params)
    }
    // ... lógica normal
}
```

---

##### 2.3 Configuration System ⚙️
**Prioridade:** 🟡 ALTA
**Esforço:** 3 dias

**Funcionalidades:**
- [ ] Config file: ~/.ollama-code/mcp.json
- [ ] Definir MCP servers e seus endpoints
- [ ] Environment variables
- [ ] Validation de config

**Formato de Config:**
```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"],
      "transport": "stdio"
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      },
      "transport": "stdio"
    },
    "brave-search": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-brave-search"],
      "env": {
        "BRAVE_API_KEY": "${BRAVE_API_KEY}"
      },
      "transport": "stdio"
    }
  }
}
```

---

##### 2.4 Popular MCP Servers Integration 🌐
**Prioridade:** 🟢 MÉDIA
**Esforço:** 1 semana

**MCP Servers a Suportar:**
- [ ] @modelcontextprotocol/server-filesystem
- [ ] @modelcontextprotocol/server-github
- [ ] @modelcontextprotocol/server-brave-search
- [ ] @modelcontextprotocol/server-puppeteer
- [ ] @modelcontextprotocol/server-postgres
- [ ] Custom servers (documentar como criar)

**Documentação:**
- [ ] Guia de instalação de MCP servers
- [ ] Exemplos de uso
- [ ] Como criar MCP server customizado

---

#### Entregáveis da Fase 2
- [ ] MCP Client completo
- [ ] MCP Tool Registry
- [ ] Configuration system
- [ ] Integração com 3+ MCP servers populares
- [ ] Documentação completa de MCP
- [ ] Testes E2E com MCP servers
- [ ] Exemplos de uso

**ROI:** ⭐⭐⭐⭐⭐ (Crítico)
**Impacto:** 🚀 Desbloquearia ecossistema de plugins

---

### 🟡 Fase 3: Advanced Agent Features (2-3 semanas)
**Status:** 📋 PLANEJADA
**Prioridade:** 🟡 MÉDIA-ALTA
**Início Previsto:** Março 2025

#### Objetivo
Implementar features avançadas de agentes: subagents, context isolation, e multi-model support.

#### Features

##### 3.1 Subagent System (Task Tool) 🤖
**Prioridade:** 🟡 MÉDIA
**Esforço:** 2 semanas

**Funcionalidades:**
- [ ] Spawning de subagents
- [ ] Context isolation
- [ ] Different models per subagent
- [ ] Agent types: Explore, Plan, Execute
- [ ] Communication entre agents
- [ ] Timeout e resource limits

**Arquivos a Criar:**
```
internal/subagent/
├── manager.go      # Subagent manager
├── types.go        # Agent types
├── context.go      # Context isolation
├── executor.go     # Agent executor
└── manager_test.go # Testes
```

**API Exemplo:**
```go
type SubagentManager interface {
    Spawn(agentType string, prompt string, model string) (*Subagent, error)
    Wait(agent *Subagent) (string, error)
    Kill(agent *Subagent) error
}
```

**Casos de Uso:**
```go
// Delegar busca complexa para subagent Explore
agent, _ := deps.SubagentManager.Spawn("Explore",
    "Find all API endpoints in the codebase",
    "qwen2.5-coder:1.5b") // modelo mais rápido

result, _ := deps.SubagentManager.Wait(agent)
```

---

##### 3.2 Multi-Model Support 🎭
**Prioridade:** 🟡 MÉDIA
**Esforço:** 1 semana

**Funcionalidades:**
- [ ] Configurar modelos diferentes por operação
- [ ] Fast model para intent detection
- [ ] Smart model para code generation
- [ ] Model switching dinâmico

**Config Exemplo:**
```json
{
  "models": {
    "intent": "qwen2.5-coder:1.5b",      // rápido
    "code": "qwen2.5-coder:7b",          // preciso
    "search": "qwen2.5-coder:0.5b",      // ultra-rápido
    "default": "qwen2.5-coder:7b"
  }
}
```

---

##### 3.3 Background Task Management 🔄
**Prioridade:** 🟢 BAIXA
**Esforço:** 1 semana

**Funcionalidades:**
- [ ] Run tasks in background
- [ ] Monitor task output (BashOutput equivalente)
- [ ] Kill background tasks
- [ ] Task status tracking

**API Exemplo:**
```go
taskID, _ := deps.BackgroundManager.Run("npm install")
output := deps.BackgroundManager.GetOutput(taskID)
deps.BackgroundManager.Kill(taskID)
```

---

#### Entregáveis da Fase 3
- [ ] Subagent system completo
- [ ] Multi-model support
- [ ] Background task management
- [ ] Testes E2E
- [ ] Documentação

**ROI:** ⭐⭐⭐⭐ (Alto)
**Impacto:** 🚀 Permite tarefas complexas paralelas

---

### 🔵 Fase 4: Multimodal Support (3-6 meses)
**Status:** 🕐 AGUARDANDO OLLAMA
**Prioridade:** 🔴 ALTA (quando disponível)
**Início Previsto:** Q2-Q3 2025

#### Objetivo
Adicionar suporte a análise de imagens, screenshots, e PDFs quando Ollama lançar modelos multimodais.

#### Dependências
- ⏳ Aguardando Ollama lançar modelo multimodal (LLaVA, Qwen-VL, etc)
- ⏳ Modelo precisa ser competitivo com GPT-4V/Claude 3

#### Features

##### 4.1 Image Analysis 🖼️
**Funcionalidades:**
- [ ] Read tool suporta PNG, JPG, JPEG
- [ ] Screenshot analysis
- [ ] UI debugging
- [ ] Diagram understanding
- [ ] OCR capabilities

---

##### 4.2 PDF Reading 📄
**Funcionalidades:**
- [ ] Read tool suporta PDF
- [ ] Text extraction
- [ ] Visual element analysis
- [ ] Multi-page support

---

##### 4.3 Video Frame Analysis 🎬
**Funcionalidades (opcional):**
- [ ] Analisar frames de vídeo
- [ ] UI flow understanding
- [ ] Demo analysis

---

#### Entregáveis da Fase 4
- [ ] Image analysis completo
- [ ] PDF reading
- [ ] Video frame analysis (opcional)
- [ ] Documentação
- [ ] Benchmarks vs Claude 3

**ROI:** ⭐⭐⭐⭐⭐ (Muito Alto quando disponível)
**Impacto:** 🚀 Paridade total com Claude Code CLI

---

### 🟢 Fase 5: IDE Integration & Polish (2-3 meses)
**Status:** 📋 PLANEJADA
**Prioridade:** 🟢 BAIXA-MÉDIA
**Início Previsto:** Q3 2025

#### Objetivo
Integração com IDEs e polish geral da UX.

#### Features

##### 5.1 VS Code Extension 📝
**Prioridade:** 🟢 MÉDIA
**Esforço:** 1 mês

**Funcionalidades:**
- [ ] Sidebar panel
- [ ] Inline suggestions
- [ ] File context menu
- [ ] Chat interface
- [ ] Diff preview

---

##### 5.2 Jupyter Notebook Support 📓
**Prioridade:** 🟢 BAIXA
**Esforço:** 2 semanas

**Funcionalidades:**
- [ ] NotebookEdit tool
- [ ] Cell manipulation
- [ ] Execute notebooks
- [ ] Output analysis

---

##### 5.3 Enhanced Terminal UX 💻
**Prioridade:** 🟡 MÉDIA
**Esforço:** 1 semana

**Funcionalidades:**
- [ ] Syntax highlighting no output
- [ ] Progress bars melhores
- [ ] Emoji consistency
- [ ] Color themes

---

#### Entregáveis da Fase 5
- [ ] VS Code extension (beta)
- [ ] Jupyter support
- [ ] Enhanced terminal UX
- [ ] Documentação
- [ ] Marketplace listing (VS Code)

**ROI:** ⭐⭐⭐ (Médio - nice to have)

---

## 📊 Métricas de Sucesso

### Por Fase

| Fase | Métrica | Target | Status |
|------|---------|--------|--------|
| **Fase 0** | Taxa de sucesso QA | 100% (6/6) | ✅ 100% |
| **Fase 0** | Bugs críticos | 0 | ✅ 0 |
| **Fase 1** | TODO tracking usage | >80% dos handlers | 🔄 TBD |
| **Fase 1** | User satisfaction | >4.5/5 | 🔄 TBD |
| **Fase 2** | MCP servers suportados | ≥3 | 📋 TBD |
| **Fase 2** | MCP tools disponíveis | ≥20 | 📋 TBD |
| **Fase 3** | Subagents working | 100% | 📋 TBD |
| **Fase 4** | Multimodal accuracy | ≥90% vs Claude | 🕐 TBD |
| **Fase 5** | VS Code downloads | ≥1000 | 📋 TBD |

### Paridade Geral

**Objetivo Final:** ≥95% paridade com Claude Code CLI

| Categoria | Atual | Target Fase 1 | Target Fase 2 | Target Final |
|-----------|-------|---------------|---------------|--------------|
| Core Features | 95% | 95% | 95% | 95% |
| Extensibility | 0% | 0% | **90%** | 95% |
| User Interaction | 60% | **85%** | 90% | 95% |
| Agent Features | 0% | 10% | 20% | **80%** |
| Multimodal | 0% | 0% | 0% | **90%** |
| **TOTAL** | **70%** | **78%** | **83%** | **≥95%** |

---

## 🎯 Priorização de Esforço

### ROI Matrix

```
Alto ROI, Baixo Esforço (QUICK WINS):
├─ ✅ TODO Tracking (1 semana)
├─ ✅ Enhanced User Interaction (1 semana)
└─ ✅ Better Diff/Edit (3 dias)

Alto ROI, Alto Esforço (INVESTIMENTOS):
├─ 🔴 MCP Protocol Support (3-4 semanas)
├─ 🟡 Subagent System (2 semanas)
└─ 🔵 Multimodal (aguardando Ollama)

Baixo ROI, Baixo Esforço (PREENCHER GAPS):
├─ 🟢 Background Task Management (1 semana)
└─ 🟢 Enhanced Terminal UX (1 semana)

Baixo ROI, Alto Esforço (EVITAR/ADIAR):
├─ 🟢 VS Code Extension (1 mês)
└─ 🟢 Jupyter Support (2 semanas)
```

---

## 📅 Cronograma Estimado

```
2025 Q1 (Jan-Mar):
├─ Janeiro: Fase 1 (TODO + User Interaction)
├─ Fevereiro: Fase 2 (MCP - parte 1)
└─ Março: Fase 2 (MCP - parte 2) + Fase 3 início

2025 Q2 (Apr-Jun):
├─ Abril: Fase 3 (Subagents)
├─ Maio: Fase 3 (Multi-Model)
└─ Junho: Aguardando Ollama multimodal

2025 Q3 (Jul-Sep):
├─ Julho: Fase 4 (Multimodal - se disponível)
├─ Agosto: Fase 5 (IDE Integration)
└─ Setembro: Polish e testes

2025 Q4 (Oct-Dec):
├─ Outubro: Beta testing público
├─ Novembro: Bug fixes e otimizações
└─ Dezembro: Release v1.0.0
```

---

## 🚀 Como Contribuir

### Para Desenvolvedores

1. **Escolha uma feature** do roadmap
2. **Crie uma issue** no GitHub com proposta
3. **Fork e implemente** seguindo padrões do projeto
4. **Testes** (>85% coverage)
5. **Pull Request** com documentação

### Para Usuários

1. **Teste** as features em desenvolvimento
2. **Reporte bugs** via GitHub Issues
3. **Sugira features** para roadmap
4. **Contribua documentação**

---

## 📝 Notas e Decisões

### Por Que Manual DI ao Invés de Wire?
- Wire foi arquivado em 2024 pelo Google
- Manual DI é mais idiomático em Go
- Sem dependências externas
- Mais fácil de debugar

### Por Que Aguardar Multimodal?
- Ollama ainda não tem modelos multimodais competitivos
- Implementar agora seria usar APIs externas (contra filosofia do projeto)
- Melhor aguardar Ollama lançar modelo nativo

### Por Que MCP é Prioridade?
- Maior gap funcional vs Claude Code CLI
- Desbloquearia ecossistema inteiro de plugins
- Community-driven extensions
- ROI altíssimo

---

## 🔗 Referências

- [Model Context Protocol Spec](https://modelcontextprotocol.io/)
- [Claude Code CLI Documentation](https://claude.com/code/docs)
- [Ollama Model Library](https://ollama.ai/library)
- [CHANGELOG.md](./CHANGELOG.md)
- [CLAUDE.md](./CLAUDE.md)

---

## ✅ Checklist de Governança

Antes de cada release:

- [ ] Todos os testes passando (unit + E2E + regression)
- [ ] CHANGELOG.md atualizado
- [ ] Documentação atualizada
- [ ] Performance benchmarks rodados
- [ ] Security audit realizado
- [ ] Breaking changes documentadas
- [ ] Migration guide (se necessário)
- [ ] Release notes escritas

---

**Roadmap mantido por:** Claude Code + Community
**Última Revisão:** 30 de Dezembro de 2024
**Próxima Revisão:** 31 de Janeiro de 2025

---

**Status das Fases:**
- ✅ Completo
- 🔄 Em Progresso
- 📋 Planejado
- 🕐 Aguardando Dependências
- ❌ Cancelado
