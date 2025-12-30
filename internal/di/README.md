# Manual Dependency Injection (DI)

Este pacote fornece **Manual Dependency Injection** para o projeto ollama-code.

## 📋 Por que Manual DI?

Optamos por Manual DI ao invés de frameworks como Wire (arquivado) ou Fx porque:

- ✅ **Idiomático em Go**: Go favorece explicitação sobre "mágica"
- ✅ **Sem dependências externas**: Não dependemos de frameworks de terceiros
- ✅ **Fácil de debugar**: Código gerado manualmente é mais fácil de entender
- ✅ **Simples**: Sem curva de aprendizado adicional
- ✅ **Flexível**: Fácil de modificar conforme necessário

## 🏗️ Estrutura

```
internal/di/
├── config.go        # Config struct e converters
├── providers.go     # Provider functions para cada componente
├── agent.go         # InitializeAgent usando providers
└── README.md        # Este arquivo
```

## 📦 Como Usar

### Opção 1: Usar agent.NewAgent (Recomendado)

A forma mais simples é usar o construtor padrão do agent:

```go
import "github.com/johnpitter/ollama-code/internal/agent"

cfg := agent.Config{
    OllamaURL: "http://localhost:11434",
    Model:     "qwen2.5-coder:7b",
    Mode:      modes.ModeInteractive,
    WorkDir:   "/path/to/project",
}

agent, err := agent.NewAgent(cfg)
if err != nil {
    log.Fatal(err)
}
```

### Opção 2: Usar di.InitializeAgent

Se você quiser usar os providers organizados:

```go
import (
    "github.com/johnpitter/ollama-code/internal/di"
    "github.com/johnpitter/ollama-code/internal/modes"
)

cfg := &di.Config{
    OllamaURL: "http://localhost:11434",
    Model:     "qwen2.5-coder:7b",
    Mode:      modes.ModeInteractive,
    WorkDir:   "/path/to/project",
}

agent, err := di.InitializeAgent(cfg)
if err != nil {
    log.Fatal(err)
}
```

### Opção 3: Usar Providers Individuais

Para testes ou customização avançada:

```go
import "github.com/johnpitter/ollama-code/internal/di"

cfg := &di.Config{
    OllamaURL: "http://localhost:11434",
    Model:     "qwen2.5-coder:7b",
    WorkDir:   "/tmp/test",
}

// Criar apenas o que você precisa
llmClient := di.ProvideLLMClient(cfg)
intentDetector := di.ProvideIntentDetector(llmClient)
toolRegistry := di.ProvideToolRegistry(cfg)

// Use os componentes individuais
```

## 🔧 Providers Disponíveis

### Core
- `ProvideLLMClient(cfg)` - Cliente LLM (Ollama)
- `ProvideIntentDetector(client)` - Detector de intenções
- `ProvideMode(cfg)` - Modo de operação
- `ProvideWorkDir(cfg)` - Diretório de trabalho

### Registries
- `ProvideToolRegistry(cfg)` - Registry de ferramentas
- `ProvideCommandRegistry()` - Registry de comandos
- `ProvideSkillRegistry()` - Registry de skills
- `ProvideHandlerRegistry(...)` - Registry de handlers

### Managers (Opcionais)
- `ProvideSessionManager(cfg)` - Gerenciador de sessões
- `ProvideCacheManager(cfg)` - Gerenciador de cache
- `ProvideConfirmationManager()` - Gerenciador de confirmações
- `ProvideWebSearchOrchestrator()` - Orquestrador de busca web

### Outros
- `ProvideStatusLine(cfg)` - Status line
- `ProvideOllamaContext(cfg)` - Contexto OLLAMA.md

### Handlers
- `ProvideFileReadHandler()` - Handler de leitura
- `ProvideFileWriteHandler()` - Handler de escrita
- `ProvideSearchHandler()` - Handler de busca
- `ProvideExecuteHandler()` - Handler de execução
- `ProvideQuestionHandler()` - Handler de perguntas
- `ProvideGitHandler()` - Handler de Git
- `ProvideAnalyzeHandler()` - Handler de análise
- `ProvideWebSearchHandler()` - Handler de busca web

## 🧪 Testing

Em testes, você pode mockar componentes específicos:

```go
func TestMyFeature(t *testing.T) {
    // Criar config de teste
    cfg := &di.Config{
        OllamaURL: "http://localhost:11434",
        Model:     "test-model",
        WorkDir:   t.TempDir(),
    }

    // Usar apenas os providers necessários
    toolRegistry := di.ProvideToolRegistry(cfg)

    // Testar isoladamente
    result, err := toolRegistry.Execute(ctx, "file_reader", params)
    // ...
}
```

## 📊 Benefícios do Manual DI

### 1. Organização
Os providers organizam a criação de dependências em funções pequenas e focadas.

### 2. Reutilização
Providers podem ser reutilizados em testes, scripts e diferentes contextos.

### 3. Testabilidade
Fácil de mockar componentes individuais sem frameworks complexos.

### 4. Manutenção
Código explícito é mais fácil de manter e modificar.

### 5. Performance
Sem overhead de reflection ou código gerado em runtime.

## 🔄 Comparação com Frameworks

### Wire (Arquivado ❌)
```go
// Pros: Código gerado, type-safe
// Cons: Framework arquivado, complexidade adicional
wire.Build(Provider1, Provider2, ...)
```

### Uber Fx (Complexo)
```go
// Pros: Runtime DI, bem mantido
// Cons: Overhead, curva de aprendizado
fx.New(fx.Provide(Provider1, Provider2, ...))
```

### Manual DI (Escolhido ✅)
```go
// Pros: Simples, idiomático, sem deps
// Cons: Mais código boilerplate (mas explícito!)
component := ProvideComponent(dependencies...)
```

## 📚 Referências

- [Go Proverbs](https://go-proverbs.github.io/) - "Clear is better than clever"
- [Effective Go](https://golang.org/doc/effective_go) - Idiomas do Go
- [Dependency Injection in Go](https://blog.drewolson.org/dependency-injection-in-go)
