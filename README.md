# 🤖 Ollama Code - AI Code Assistant

> Assistente de código AI inteligente que funciona como Claude Code, 100% local, escrito em Go.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Windows%20%7C%20macOS-lightgrey)]()

---

## ✨ Características

### Base Features
- 🧠 **Linguagem Natural** - Sem comandos especiais (`/read`, `/exec`), apenas fale naturalmente
- 🎯 **Detecção Inteligente** - IA detecta automaticamente suas intenções usando LLM
- 🔧 **8+ Ferramentas** - Leitura/escrita de arquivos, execução de comandos, git, análise de código
- 📷 **Suporte a Imagens** - Lê e analisa imagens (PNG, JPG, GIF, etc)
- 🌐 **Pesquisa Web** - Busca na internet quando necessário (DuckDuckGo, Stack Overflow, GitHub)
- 🎛️ **3 Modos de Operação**:
  - **READ-ONLY**: Somente leitura
  - **INTERACTIVE**: Com confirmação (padrão)
  - **AUTONOMOUS**: Totalmente automático
- ⚡ **Performance Máxima** - Startup <15ms, streaming em tempo real
- 🔒 **Privacidade** - 100% local, sem envio de dados para nuvem

### Enterprise Features ✨ NEW!
- 💾 **Checkpoints & Recovery** - Volte no tempo, desfaça mudanças, recupere estados anteriores
- 📂 **Session Management** - Salve e retome sessões de trabalho
- 🧠 **Hierarchical Memory** - 5 níveis de memória (Enterprise → Project → Rules → User → Local)
- ⚡ **Slash Commands** - 10+ comandos built-in (/help, /checkpoint, /session, /doctor, etc)
- 🪝 **Hooks System** - Pre/post hooks para validação e automação
- 🎨 **Output Styles** - 4 estilos de output (default, explanatory, learning, corporate)
- 🚀 **Performance** - Context cache, async tasks, otimizações
- 🏥 **Diagnostics** - /doctor para health checks completos
- 🖥️ **Hardware Auto-Detection** - Detecta seu hardware e otimiza automaticamente
- ⚙️ **3 Presets de Configuração** - Compatibility, Performance e Ultra

---

## 🎯 Objetivo

Criar um assistente de código que funciona como Claude Code, mas rodando completamente local usando Ollama.

**Exemplo de uso:**
```bash
$ ollama-code chat

Você: Cria um servidor HTTP em Go com endpoint /health

🤖: Vou criar um servidor HTTP básico...

🔔 Confirmação necessária:
   Ação: Criar arquivo server.go
   Tipo: WRITE_FILE

Executar? [y/N]: y

✅ Arquivo criado: server.go

🤖: Servidor criado! Quer que eu execute para testar?
```

---

## 📋 Requisitos

### Hardware Alvo
- **CPU**: Intel i9 14ª gen (24 cores) ou similar
- **RAM**: 64GB
- **GPU**: NVIDIA RTX Ada 2000 (16GB VRAM) ou similar
- **Storage**: 1TB NVMe SSD

### Software
- **Go**: 1.21+
- **Ollama**: Última versão
- **CUDA**: 11.8+ (para GPU NVIDIA)
- **OS**: Linux, Windows ou macOS

---

## 🚀 Como Rodar o Projeto

### Pré-requisitos

1. **Go 1.21+** instalado
2. **Ollama** instalado e rodando
3. **Modelo Ollama** baixado

### Instalação Rápida (3 passos)

#### 1️⃣ Instalar Ollama

```bash
# Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Windows
# Baixe e instale de: https://ollama.ai/download/windows

# macOS
brew install ollama
```

Inicie o serviço Ollama:
```bash
ollama serve
```

#### 2️⃣ Baixar um modelo

```bash
# Modelo recomendado para começar (4.7GB)
ollama pull qwen2.5-coder:7b

# Ou modelos mais poderosos (se tiver GPU com 16GB+ VRAM)
ollama pull qwen2.5-coder:14b-instruct-q5_K_M  # 9.9GB
ollama pull qwen2.5-coder:32b-instruct-q6_K    # 21GB
```

#### 3️⃣ Compilar e Executar

```bash
# Clone o repositório
git clone https://github.com/JohnPitter/ollama-code.git
cd ollama-code

# Compile a aplicação
./build.sh          # Linux/macOS
# ou
.\build.bat         # Windows

# Execute!
./build/ollama-code chat
```

**Pronto!** A aplicação irá:
- ✅ Detectar automaticamente seu hardware
- ✅ Criar configuração otimizada em `~/.ollama-code/config.json`
- ✅ Iniciar o modo chat interativo

### Comandos Disponíveis

```bash
# Modo chat interativo (recomendado)
./build/ollama-code chat

# Fazer uma pergunta direta
./build/ollama-code ask "como criar um loop em Go?"

# Modo somente leitura (sem modificações)
./build/ollama-code chat --mode readonly

# Modo autônomo (sem confirmações)
./build/ollama-code chat --mode autonomous

# Ver ajuda completa
./build/ollama-code help
```

### Instalação Global (Opcional)

Para usar `ollama-code` de qualquer diretório:

```bash
# Linux/macOS
sudo cp build/ollama-code /usr/local/bin/
ollama-code chat

# Windows (PowerShell como Admin)
Copy-Item build/ollama-code.exe C:\Windows\System32\
ollama-code chat
```

---

## ⚙️ Configuração Avançada (Opcional)

### Otimizar para GPU NVIDIA

**Linux/macOS** (`~/.config/ollama/env.conf`):
```bash
export OLLAMA_GPU_LAYERS=999
export OLLAMA_NUM_GPU=1
export OLLAMA_MAX_LOADED_MODELS=2
export OLLAMA_NUM_PARALLEL=4
export OLLAMA_FLASH_ATTENTION=1
export OLLAMA_MAX_VRAM=16384
```

**Windows** (PowerShell como Admin):
```powershell
[System.Environment]::SetEnvironmentVariable('OLLAMA_GPU_LAYERS', '999', 'Machine')
[System.Environment]::SetEnvironmentVariable('OLLAMA_NUM_GPU', '1', 'Machine')
Restart-Service Ollama
```

### Ambiente Corporativo com Proxy

Use os scripts de download direto:
```bash
# Linux/macOS
chmod +x download-models-direct.sh
./download-models-direct.sh

# Windows
.\download-models-direct.ps1
```

### Editar Configuração

A aplicação cria automaticamente `~/.ollama-code/config.json` na primeira execução.
Para customizar, edite o arquivo ou veja [CONFIG.md](CONFIG.md)

---

## 📖 Uso

### Primeira Execução - Detecção Automática de Hardware

Na primeira vez que você executar o Ollama Code:

```bash
$ ./build/ollama-code chat

🔍 First run detected - Analyzing your hardware...

╔════════════════════════════════════════════════════════════╗
║          OLLAMA CODE - HARDWARE DETECTION REPORT           ║
╚════════════════════════════════════════════════════════════╝

🖥️  HARDWARE DETECTED:
   CPU: AMD Ryzen 5 5600 6-Core Processor
   Cores/Threads: 12 / 12
   RAM: 32694 MB total (14494 MB available)
   GPU: NVIDIA GeForce RTX 4070
   VRAM: 12282 MB (1 GPU(s))
   OS: windows / amd64

⚡ PERFORMANCE TIER: mid-range

🎯 PRESET SELECTED: performance
   Performance - Balanceamento entre velocidade e compatibilidade

⚙️  OPTIMIZED CONFIGURATION:
   Model: qwen2.5-coder:14b-instruct-q5_K_M
   Temperature: 0.7
   GPU Layers: 35
   Max VRAM: 9825 MB
   Sessions: enabled
   Cache: enabled (15 min)

✅ Configuration optimized for your hardware!
   Config saved to: ~/.ollama-code/config.json
```

A aplicação automaticamente:
- ✅ Detecta CPU, RAM, GPU e VRAM disponíveis
- ✅ Classifica seu hardware (high-end, mid-range, entry, low-end)
- ✅ Seleciona o melhor preset (ultra, performance, compatibility)
- ✅ Gera configuração otimizada
- ✅ Salva em `~/.ollama-code/config.json`

### Modo Interativo (Padrão)

```bash
ollama-code chat

Você: Analisa esse projeto
🤖: [Lê arquivos e explica a estrutura]

Você: Cria um servidor REST em Go
🤖: [Gera código e pede confirmação antes de criar arquivo]
```

### Modo Read-Only (Somente Leitura)

```bash
ollama-code chat --mode readonly

Você: Mostra o main.go
🤖: [Mostra conteúdo]

Você: Corrija os erros
❌ Ação bloqueada: Escrita não permitida em modo READ-ONLY
```

### Modo Autônomo (Sem Confirmação)

```bash
ollama-code chat --mode autonomous

Você: Cria um projeto completo com CRUD e testes

[10:23:45] ✓ Criado: main.go
[10:23:46] ✓ Criado: handlers/user.go
[10:23:47] ✓ Criado: tests/user_test.go
[10:23:48] ⚙️  go mod tidy
[10:23:49] ⚙️  go test ./...
[10:23:52] ✅ Testes passando
```

### Leitura de Imagens

```bash
Você: Leia a imagem screenshot.png e me diga o que tem nela

🤖: [Lê e analisa a imagem]
    A imagem mostra uma interface de usuário com...
```

### Pesquisa na Internet

```bash
Você: Como corrigir erro "permission denied" no Docker?

🌐 Pesquisando na internet...
✓ Encontrei 3 fontes relevantes

🤖: O erro ocorre quando... [solução com exemplos]

📚 Fontes:
[1] Stack Overflow - https://...
[2] Docker Docs - https://...
```

---

## 🎛️ Modos de Operação

| Modo | Flag | Descrição | Uso Recomendado |
|------|------|-----------|-----------------|
| **INTERACTIVE** | `--mode interactive` (padrão) | Confirma ações destrutivas | Desenvolvimento do dia a dia |
| **READ-ONLY** | `--mode readonly` | Apenas leitura | Code review, exploração |
| **AUTONOMOUS** | `--mode autonomous` | Tudo automático | Automação, prototipagem |

---

## 🔧 Ferramentas Disponíveis

O sistema detecta automaticamente qual ferramenta usar:

- **FileReader**: Lê arquivos de texto e imagens
- **FileWriter**: Escreve/modifica arquivos
- **CommandExecutor**: Executa comandos shell
- **CodeSearcher**: Busca em código (ripgrep)
- **ProjectAnalyzer**: Analisa estrutura do projeto
- **GitOperations**: Operações git (commit, push, etc)
- **WebSearcher**: Pesquisa na internet

**Você não precisa especificar qual ferramenta usar** - a IA escolhe automaticamente baseado no seu pedido!

---

## 📊 Performance

**No hardware alvo (i9 14ª gen + RTX Ada 2000):**

```
Startup time:      < 15ms
Memory (base):     ~10MB
Binary size:       ~8MB (otimizado)
LLM throughput:    ~30-40 tokens/s
File operations:   < 10ms
Web search:        ~2-5s (cache: <100ms)
```

---

## 🛠️ Desenvolvimento

### Build

```bash
# Build padrão
make build

# Build otimizado (produção)
make build-optimized

# Executar sem instalar
make run

# Testes
make test

# Limpar
make clean
```

### Estrutura do Projeto

```
ollama-code/
├── cmd/ollama-code/main.go          # Entry point
├── internal/
│   ├── agent/                       # Agente principal
│   ├── intent/                      # Detecção de intenções
│   ├── tools/                       # Ferramentas
│   ├── websearch/                   # Pesquisa web
│   ├── llm/                         # Client Ollama
│   └── confirmation/                # Confirmações
├── Makefile
└── IMPLEMENTATION_PLAN.md           # Plano completo
```

---

## 📚 Documentação

- **[IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md)** - Plano técnico completo de implementação (base)
- **[ENTERPRISE_FEATURES.md](ENTERPRISE_FEATURES.md)** - Funcionalidades enterprise-grade completas
- **[download-models-direct.sh](download-models-direct.sh)** - Script para download de modelos (Linux/macOS)
- **[download-models-direct.ps1](download-models-direct.ps1)** - Script para download de modelos (Windows)

---

## 🤝 Contribuindo

Este projeto foi criado como um plano de implementação completo para ser executado por uma IA (como Grok Code Fast 1).

Para contribuir:
1. Leia o [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md)
2. Siga a estrutura definida
3. Implemente fase por fase
4. Teste cada componente
5. Submeta PR

---

## 📝 Exemplos Avançados

### Criar projeto completo
```bash
Você: Cria um projeto REST API em Go com:
      - CRUD de usuários
      - Autenticação JWT
      - Testes unitários
      - Dockerfile
      - README

🤖: [Cria estrutura completa do projeto]
```

### Análise e refatoração
```bash
Você: Analisa o código e refatora seguindo Clean Code

🤖: [Analisa, sugere melhorias e aplica refatorações]
```

### Debug com pesquisa
```bash
Você: Estou tendo erro X no código, pesquise soluções e corrija

🤖: 🌐 Pesquisando...
    [Encontra solução, aplica correção]
```

---

## ⚠️ Avisos Importantes

1. **Modo Autônomo**: Use com cuidado! Todas as ações são executadas sem confirmação.
2. **GPU**: Para melhor performance, configure todas as layers para rodar na GPU.
3. **Proxy Corporativo**: Use os scripts de download direto para baixar modelos.
4. **Backup**: Sempre faça backup antes de usar modo autônomo.

---

## 🎯 Roadmap

### Base (10-12 dias)
- [x] Detecção inteligente de intenções
- [x] 3 modos de operação
- [x] Pesquisa na internet
- [x] Suporte a imagens
- [x] Streaming de respostas
- [x] 8+ ferramentas integradas

### Enterprise (24 dias adicionais)
- [x] **Checkpoints & Rewind** - Recuperação de estado
- [x] **Session Management** - Múltiplas sessões, resumir
- [x] **Hierarchical Memory** - CLAUDE.md em 5 níveis
- [x] **40+ Slash Commands** - Customizáveis
- [x] **Hooks System** - Pre/Post execution
- [x] **Telemetry** - OpenTelemetry, métricas
- [x] **Sandboxing** - Isolamento de processos
- [x] **/doctor** - Health checks & diagnostics
- [x] **Background Tasks** - Async execution
- [x] **CI/CD** - GitHub Actions, GitLab

### Futuro
- [ ] Cache de embeddings (Redis)
- [ ] Suporte a plugins MCP
- [ ] Interface web
- [ ] Integração com VS Code

---

## 📄 Licença

Apache 2.0 - Veja [LICENSE](LICENSE)

---

## 👤 Autor

Criado como especificação técnica completa para implementação por IA.

**Hardware alvo:** PC high-end com i9 14ª gen, 64GB RAM, RTX Ada 2000

---

## 🙏 Agradecimentos

- **Ollama** - Por fornecer uma forma simples de rodar LLMs localmente
- **QWen 2.5 Coder** - Modelo state-of-the-art para código
- **Claude Code** - Inspiração para o design do sistema

---

**🚀 Comece agora:**
```bash
git clone <repo>
cd ollama-code
make install
ollama-code chat
```
