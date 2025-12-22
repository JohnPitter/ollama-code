# 🛠️ Guia de Uso das Ferramentas Avançadas

Este guia apresenta as **7 ferramentas avançadas** implementadas no Ollama Code, conforme especificação do QA Plan.

## 📦 1. Dependency Manager

**Propósito:** Gerenciamento inteligente de dependências do projeto

### Operações Disponíveis

#### Verificar Dependências (`check`)
```bash
# Lista todas as dependências do projeto
ollama-code ask "verifique as dependências do projeto"
```

**O que faz:**
- Detecta automaticamente o tipo de projeto (Node.js, Go, Python, Rust)
- Lista dependências instaladas
- Mostra dependências desatualizadas (Node.js)

#### Instalar Pacote (`install`)
```bash
# Instala nova dependência
ollama-code ask "instale o pacote express"
ollama-code ask "adicione a dependência github.com/gin-gonic/gin"
```

#### Atualizar Dependências (`update`)
```bash
# Atualiza todas as dependências
ollama-code ask "atualize todas as dependências"
```

#### Auditoria de Segurança (`audit`)
```bash
# Verifica vulnerabilidades
ollama-code ask "verifique vulnerabilidades nas dependências"
```

**Ferramentas suportadas:**
- **Node.js**: npm audit, npm outdated
- **Go**: govulncheck
- **Python**: safety check

---

## 📚 2. Documentation Generator

**Propósito:** Geração automática de documentação profissional

### Tipos de Documentação

#### Documentação Automática (`auto`)
```bash
# Detecta tipo de projeto e gera documentação apropriada
ollama-code ask "gere documentação automática do projeto"
```

#### README.md
```bash
# Cria README.md completo
ollama-code ask "crie um README.md para este projeto"
```

**Gera:**
- Nome do projeto
- Seções: Descrição, Instalação, Uso, Contribuição, Licença
- Template customizável

#### Documentação Go (GoDoc)
```bash
ollama-code ask "gere documentação GoDoc"
```

#### Documentação JavaScript (JSDoc)
```bash
ollama-code ask "gere documentação JSDoc"
```

#### Documentação de API
```bash
# Detecta arquivos OpenAPI/Swagger
ollama-code ask "gere documentação da API"
```

---

## 🔒 3. Security Scanner

**Propósito:** Análise de segurança multicamada do código

### Tipos de Scan

#### Scan Completo (`all`)
```bash
# Executa todos os scans de segurança
ollama-code ask "faça um scan de segurança completo"
```

#### Detecção de Secrets
```bash
# Busca secrets vazados no código
ollama-code ask "procure por secrets no código"
```

**Detecta:**
- ✅ API Keys
- ✅ AWS Access Keys
- ✅ Passwords em código
- ✅ Private Keys (RSA, DSA, EC, OpenSSH)
- ✅ JWT Tokens
- ✅ GitHub Tokens (ghp_...)

#### SAST (Static Analysis)
```bash
# Análise estática de segurança
ollama-code ask "execute análise estática de segurança"
```

**Ferramentas:**
- **Go**: gosec + go vet
- **JavaScript**: eslint-plugin-security
- **Python**: bandit

#### Scan de Dependências
```bash
# Verifica vulnerabilidades em dependências
ollama-code ask "verifique vulnerabilidades nas libs"
```

---

## 🔄 4. Advanced Refactoring

**Propósito:** Refatorações automatizadas complexas

### Operações Disponíveis

#### Renomear Símbolo (`rename`)
```bash
# Renomeia função/variável em todo o projeto
ollama-code ask "renomeie a função 'oldName' para 'newName'"

# Renomear apenas em arquivo específico
ollama-code ask "renomeie 'oldFunc' para 'newFunc' apenas em main.go"
```

**Recursos:**
- Parse AST para Go (máxima precisão)
- Renomeia em múltiplos arquivos
- Suporte: Go, JavaScript, Python, Java, C++

#### Encontrar Duplicações (`find_duplicates`)
```bash
# Detecta código duplicado
ollama-code ask "encontre código duplicado no projeto"
```

**Detecta:**
- Blocos de 5+ linhas duplicados
- Localização exata (arquivo:linha)
- Sugestões de refatoração

#### Extract Method (Planejado)
```bash
ollama-code ask "extraia este código para um método separado"
```

#### Extract Class (Planejado)
```bash
ollama-code ask "extraia estes campos para uma classe"
```

---

## 🧪 5. Test Runner

**Propósito:** Execução e gerenciamento de testes automatizados

### Ações Disponíveis

#### Executar Testes (`run`)
```bash
# Executa todos os testes
ollama-code ask "execute os testes"
```

**Suporte:**
- **Go**: `go test ./...`
- **Node.js**: `npm test`
- **Python**: `pytest` ou `unittest`

#### Cobertura de Código (`coverage`)
```bash
# Testes com cobertura
ollama-code ask "execute testes com cobertura"
```

**Gera:**
- Go: `coverage.html` (visualização web)
- Node.js: Relatório Jest
- Python: Relatório pytest-cov (HTML + terminal)

#### Modo Watch
```bash
# Modo watch para desenvolvimento
ollama-code ask "ative modo watch dos testes"
```

**Sugestões por linguagem:**
- Node.js: `npm test -- --watch`
- Python: `pytest-watch`
- Go: `gow test ./...`

#### Teste Único (`single`)
```bash
# Executa teste específico
ollama-code ask "execute o teste TestUserLogin"
```

---

## ⏱️ 6. Background Task Manager

**Propósito:** Gerenciamento de tarefas assíncronas

### Tarefas Pré-configuradas

#### Long Test
```bash
# Simula teste longo (10 etapas)
ollama-code ask "inicie tarefa long_test em background"
```

#### Build
```bash
# Simula build completo (4 fases)
ollama-code ask "execute build em background"
```

**Fases:** Compilando → Linkando → Otimizando → Empacotando

#### Deploy
```bash
# Simula deployment (4 fases)
ollama-code ask "faça deploy em background"
```

**Fases:** Preparando → Uploading → Configurando → Validando

#### Analysis
```bash
# Análise de código assíncrona (3 fases)
ollama-code ask "execute análise em background"
```

### Gerenciamento de Tarefas

#### Listar Tarefas
```bash
ollama-code ask "liste tarefas em background"
```

**Mostra:**
- 📋 ID da tarefa
- ⏳ Status (pending, running, completed, failed)
- 📊 Progresso (0-100%)

#### Verificar Status
```bash
ollama-code ask "verifique status da tarefa task_12345"
```

#### Cancelar Tarefa
```bash
ollama-code ask "cancele a tarefa task_12345"
```

#### Obter Resultado
```bash
ollama-code ask "mostre resultado da tarefa task_12345"
```

---

## ⚡ 7. Performance Profiler

**Propósito:** Análise de performance e profiling

### Tipos de Profiling

#### Benchmarks
```bash
# Executa benchmarks
ollama-code ask "execute benchmarks"

# Benchmark com padrão específico (Go)
ollama-code ask "execute benchmark de string operations"
```

**Suporte:**
- **Go**: `go test -bench -benchmem`
- **Node.js**: benchmark.js, tinybench, vitest
- **Python**: pytest-benchmark

#### CPU Profiling
```bash
# Profiling de CPU
ollama-code ask "execute CPU profiling"
```

**Instruções por linguagem:**

**Go:**
```bash
# Durante testes
go test -cpuprofile=cpu.prof -bench=.

# Visualizar
go tool pprof -http=:8080 cpu.prof
```

**Node.js:**
```bash
# Com --prof flag
node --prof app.js
node --prof-process isolate-*.log > processed.txt

# Com clinic.js
clinic doctor -- node app.js
```

**Python:**
```bash
# cProfile
python -m cProfile -o output.prof script.py

# py-spy
py-spy record -o profile.svg -- python script.py
```

#### Memory Profiling
```bash
# Profiling de memória
ollama-code ask "execute memory profiling"
```

**Ferramentas:**
- Go: heap analysis, alloc/inuse space
- Node.js: Chrome DevTools, clinic heapprofiler
- Python: memory_profiler, tracemalloc

#### Execution Tracing
```bash
# Tracing de execução
ollama-code ask "execute execution tracing"
```

**Visualização:**
- Go: `go tool trace trace.out`
- Node.js: chrome://tracing

#### Analisar Profiles
```bash
# Detecta e analisa profiles existentes
ollama-code ask "analise profiles de performance"
```

**Detecta:**
- cpu.prof, mem.prof, trace.out
- profile.prof, heap.prof
- Mostra tamanho, data, sugestões de visualização

---

## 📊 Resumo de Comandos

### Por Categoria

**Dependências:**
```bash
ollama-code ask "verifique as dependências"
ollama-code ask "instale express"
ollama-code ask "atualize dependências"
ollama-code ask "audite vulnerabilidades"
```

**Documentação:**
```bash
ollama-code ask "gere documentação automática"
ollama-code ask "crie README.md"
ollama-code ask "gere GoDoc"
```

**Segurança:**
```bash
ollama-code ask "scan de segurança completo"
ollama-code ask "procure secrets"
ollama-code ask "análise estática"
```

**Refatoração:**
```bash
ollama-code ask "renomeie oldFunc para newFunc"
ollama-code ask "encontre código duplicado"
```

**Testes:**
```bash
ollama-code ask "execute testes"
ollama-code ask "testes com cobertura"
ollama-code ask "modo watch"
```

**Background:**
```bash
ollama-code ask "execute build em background"
ollama-code ask "liste tarefas"
ollama-code ask "status da tarefa task_12345"
```

**Performance:**
```bash
ollama-code ask "execute benchmarks"
ollama-code ask "CPU profiling"
ollama-code ask "memory profiling"
ollama-code ask "analise profiles"
```

---

## 🎯 Dicas de Uso

### 1. Detecção Automática
Todas as ferramentas detectam automaticamente o tipo de projeto (Go, Node.js, Python, Rust).

### 2. Modo Interativo vs Autônomo
- **Interativo**: Pede confirmação antes de executar
- **Autônomo**: Executa automaticamente

```bash
# Modo interativo (padrão)
ollama-code chat

# Modo autônomo
ollama-code chat --mode autonomous
```

### 3. Combinando Ferramentas
```bash
# Workflow completo
ollama-code ask "verifique vulnerabilidades, depois atualize dependências e execute testes"
```

### 4. Tarefas Longas
Use Background Tasks para operações demoradas:
```bash
ollama-code ask "execute análise completa em background"
# Continua trabalhando enquanto processa
ollama-code ask "liste tarefas"  # Verifica progresso
```

---

## 🔧 Configuração Avançada

### Personalização

Cada ferramenta pode ser configurada via `config.json`:

```json
{
  "tools": {
    "dependency_manager": {
      "enabled": true,
      "auto_update": false
    },
    "security_scanner": {
      "enabled": true,
      "scan_on_save": false
    },
    "test_runner": {
      "enabled": true,
      "auto_run": false,
      "watch_mode": false
    }
  }
}
```

---

## 📝 Exemplos Práticos

### Workflow de Desenvolvimento Completo

```bash
# 1. Iniciar projeto
ollama-code ask "crie README.md"

# 2. Verificar dependências
ollama-code ask "verifique dependências desatualizadas"

# 3. Scan de segurança
ollama-code ask "scan de segurança completo"

# 4. Executar testes
ollama-code ask "execute testes com cobertura"

# 5. Benchmarks
ollama-code ask "execute benchmarks"

# 6. Refatoração
ollama-code ask "encontre código duplicado"
```

### Análise de Performance Profunda

```bash
# 1. Benchmarks iniciais
ollama-code ask "execute benchmarks"

# 2. CPU Profiling
ollama-code ask "execute CPU profiling"

# 3. Memory Profiling
ollama-code ask "execute memory profiling"

# 4. Análise de profiles
ollama-code ask "analise todos os profiles"
```

---

## 🚀 Suporte e Contribuição

Para mais informações, consulte:
- [Documentação Completa de Implementação](ADVANCED_TOOLS_IMPLEMENTATION.md)
- [QA Test Plan](../docs/QA_TEST_PLAN.md)
- [Issues no GitHub](https://github.com/johnpitter/ollama-code/issues)

---

*Documentação gerada para Ollama Code - 22/12/2024*
