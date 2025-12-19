# 🚀 Guia de Instalação - Ollama Code

## Pré-requisitos

### 1. Instalar Go 1.21+

**Windows:**
```powershell
# Baixar e instalar de: https://go.dev/dl/
# Ou via winget:
winget install GoLang.Go

# Verificar:
go version
```

**Linux:**
```bash
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version
```

**macOS:**
```bash
brew install go@1.21
go version
```

### 2. Instalar Ollama

**Windows:**
```powershell
# Baixar de: https://ollama.ai/download/windows
# Executar OllamaSetup.exe
```

**Linux:**
```bash
curl -fsSL https://ollama.ai/install.sh | sh
sudo systemctl enable ollama
sudo systemctl start ollama
```

**macOS:**
```bash
# Baixar de: https://ollama.ai/download/mac
# Ou via brew:
brew install ollama
brew services start ollama
```

### 3. Baixar Modelo

Execute o script de download incluído ou baixe manualmente:

**Opção 1 - Script automático:**
```bash
# Linux/macOS:
./download-models-direct.sh

# Windows:
powershell -ExecutionPolicy Bypass -File download-models-direct.ps1
```

**Opção 2 - Manual:**
```bash
ollama pull qwen2.5-coder:32b-instruct-q6_K
```

---

## Instalação do Ollama Code

### Método 1: Build Manual

```bash
# Clone o repositório
git clone https://github.com/johnpitter/ollama-code.git
cd ollama-code

# Baixar dependências
go mod download
go mod tidy

# Build
make build

# Instalar localmente
make install-local

# Ou instalar globalmente (Linux/macOS requer sudo)
make install
```

### Método 2: Build Rápido

```bash
cd ollama-code

# Desenvolvimento
make dev

# Executar diretamente
./build/ollama-code chat
```

### Método 3: Executar sem Build

```bash
cd ollama-code
go run ./cmd/ollama-code chat
```

---

## Configuração

### Variáveis de Ambiente (Opcional)

```bash
# URL do Ollama (padrão: http://localhost:11434)
export OLLAMA_URL=http://localhost:11434

# Modelo padrão
export OLLAMA_MODEL=qwen2.5-coder:32b-instruct-q6_K

# Modo de operação padrão
export OLLAMA_CODE_MODE=interactive
```

### Configuração do Ollama (RTX Ada 2000)

**Linux/macOS** (`~/.config/ollama/env.conf`):
```bash
export OLLAMA_GPU_LAYERS=999
export OLLAMA_NUM_GPU=1
export CUDA_VISIBLE_DEVICES=0
export OLLAMA_MAX_LOADED_MODELS=2
export OLLAMA_NUM_PARALLEL=4
export OLLAMA_FLASH_ATTENTION=1
export OLLAMA_MAX_VRAM=16384
```

**Windows** (PowerShell como Admin):
```powershell
[System.Environment]::SetEnvironmentVariable('OLLAMA_GPU_LAYERS', '999', 'Machine')
[System.Environment]::SetEnvironmentVariable('OLLAMA_NUM_GPU', '1', 'Machine')
[System.Environment]::SetEnvironmentVariable('OLLAMA_MAX_VRAM', '16384', 'Machine')
```

Reinicie o serviço Ollama após configurar.

---

## Uso

### Modo Chat Interativo

```bash
# Padrão (modo interactive)
ollama-code chat

# Modo readonly (apenas leitura)
ollama-code chat --mode readonly

# Modo autonomous (automático, sem confirmações)
ollama-code chat --mode autonomous

# Especificar diretório de trabalho
ollama-code chat --workdir /path/to/project

# Modelo customizado
ollama-code chat --model deepseek-coder:33b
```

### Pergunta Rápida

```bash
ollama-code ask "como fazer um servidor HTTP em Go?"
```

### Exemplos de Uso

```bash
# Iniciar chat
ollama-code chat

# No chat, você pode:
💬 Você: leia o arquivo main.go
💬 Você: mostre a estrutura do projeto
💬 Você: execute os testes
💬 Você: busque por "handleRequest" no código
💬 Você: pesquise na internet como fazer goroutines
💬 Você: crie um arquivo test.go com testes básicos
💬 Você: commita as mudanças
```

---

## Comandos no Chat

- `exit` ou `quit` - Sair
- `help` - Mostrar ajuda
- `clear` - Limpar histórico
- `mode` - Mostrar modo atual
- `pwd` - Mostrar diretório atual

---

## Verificação

### Testar se está funcionando:

```bash
# 1. Verificar Ollama
ollama list

# 2. Testar modelo
ollama run qwen2.5-coder:32b-instruct-q6_K "Hello"

# 3. Testar Ollama Code
ollama-code ask "Hello, how are you?"
```

---

## Troubleshooting

### Ollama não conecta:
```bash
# Verificar se está rodando
curl http://localhost:11434/api/tags

# Verificar serviço (Linux)
sudo systemctl status ollama

# Restart (Linux)
sudo systemctl restart ollama
```

### Go não encontrado:
```bash
# Verificar instalação
which go
go version

# Adicionar ao PATH se necessário
export PATH=$PATH:/usr/local/go/bin
```

### Modelo não encontrado:
```bash
# Listar modelos instalados
ollama list

# Baixar modelo
ollama pull qwen2.5-coder:32b-instruct-q6_K
```

### Build falha:
```bash
# Limpar e rebuild
make clean
go mod tidy
make deps
make build
```

---

## Próximos Passos

Após instalação bem-sucedida:

1. Leia [README.md](README.md) para visão geral
2. Veja [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) para detalhes técnicos
3. Explore [ENTERPRISE_FEATURES.md](ENTERPRISE_FEATURES.md) para features avançadas

---

**Pronto para usar!** 🎉

Execute `ollama-code chat` e comece a codificar com IA!
