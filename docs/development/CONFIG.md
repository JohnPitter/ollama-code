# ⚙️ Configuration Guide - Ollama Code

## 📍 Localização do Arquivo de Configuração

O Ollama Code usa um arquivo JSON para configuração. Localização padrão:

**Linux/macOS:**
```
~/.ollama-code/config.json
```

**Windows:**
```
C:\Users\<seu-usuario>\.ollama-code\config.json
```

## 🚀 Inicialização Automática

Na primeira execução, o Ollama Code **automaticamente**:
1. Detecta que não existe arquivo de configuração
2. Cria o diretório `~/.ollama-code/`
3. Gera `config.json` com valores padrão
4. Salva o arquivo

Você **não precisa** criar manualmente!

## 📝 Estrutura do Arquivo

O arquivo config.json possui 3 seções principais:

### 1. Ollama (Configurações do Servidor)

```json
{
  "ollama": {
    "url": "http://localhost:11434",
    "model": "qwen2.5-coder:32b-instruct-q6_K",
    "temperature": 0.7,
    "max_tokens": 4096,
    "gpu_layers": 999,
    "num_gpu": 1,
    "max_vram": 16384,
    "num_parallel": 4,
    "flash_attention": true
  }
}
```

**Campos:**
- `url` - URL do servidor Ollama (padrão: http://localhost:11434)
- `model` - Modelo a ser usado
- `temperature` - Criatividade (0.0-1.0, padrão: 0.7)
- `max_tokens` - Máximo de tokens por resposta
- `gpu_layers` - Layers a carregar na GPU (999 = todas)
- `num_gpu` - Número de GPUs a usar
- `max_vram` - Máximo de VRAM em MB (16384 = 16GB)
- `num_parallel` - Requisições paralelas
- `flash_attention` - Usar flash attention (mais rápido)

### 2. App (Configurações da Aplicação)

```json
{
  "app": {
    "mode": "interactive",
    "work_dir": ".",
    "output_style": "default",
    "enable_colors": true,
    "enable_checkpoints": true,
    "enable_sessions": true,
    "enable_memory": true,
    "checkpoint_retention": 30,
    "max_checkpoints": 100,
    "log_level": "info",
    "log_file": ""
  }
}
```

**Campos:**
- `mode` - Modo padrão: `readonly`, `interactive`, `autonomous`
- `work_dir` - Diretório de trabalho padrão
- `output_style` - Estilo: `default`, `explanatory`, `learning`, `corporate`
- `enable_colors` - Usar cores no terminal
- `enable_checkpoints` - Habilitar sistema de checkpoints
- `enable_sessions` - Habilitar gerenciamento de sessões
- `enable_memory` - Habilitar memória hierárquica
- `checkpoint_retention` - Dias de retenção de checkpoints
- `max_checkpoints` - Máximo de checkpoints armazenados
- `log_level` - Nível de log: `debug`, `info`, `warn`, `error`
- `log_file` - Arquivo de log (vazio = não salvar)

### 3. Performance (Otimizações)

```json
{
  "performance": {
    "cache_ttl": 15,
    "enable_cache": true,
    "max_concurrent_tools": 3,
    "command_timeout": 60
  }
}
```

**Campos:**
- `cache_ttl` - Tempo de vida do cache em minutos
- `enable_cache` - Habilitar cache de contexto
- `max_concurrent_tools` - Máximo de ferramentas executando em paralelo
- `command_timeout` - Timeout de comandos shell em segundos

## 🎯 Uso

### 1. Usar configuração padrão

```bash
ollama-code chat
```

O sistema automaticamente:
- Procura `~/.ollama-code/config.json`
- Se não existe, cria com valores padrão
- Carrega e valida configuração

### 2. Especificar arquivo customizado

```bash
ollama-code chat --config /path/to/custom-config.json
```

### 3. Sobrescrever com flags

As flags de linha de comando **sobrescrevem** o arquivo:

```bash
# Sobrescreve mode do config.json
ollama-code chat --mode autonomous

# Sobrescreve model e url
ollama-code chat --model llama3:8b --url http://192.168.1.100:11434

# Múltiplas sobrescritas
ollama-code chat --mode readonly --workdir /project
```

**Ordem de prioridade:**
1. **Flags CLI** (maior prioridade)
2. **Arquivo customizado** (--config)
3. **Arquivo padrão** (~/.ollama-code/config.json)
4. **Defaults hardcoded** (menor prioridade)

## 📋 Exemplos de Configuração

### Configuração para Desenvolvimento

```json
{
  "ollama": {
    "url": "http://localhost:11434",
    "model": "qwen2.5-coder:7b",
    "temperature": 0.5,
    "max_tokens": 2048
  },
  "app": {
    "mode": "interactive",
    "enable_colors": true,
    "enable_checkpoints": true,
    "log_level": "debug",
    "log_file": "/tmp/ollama-code.log"
  },
  "performance": {
    "enable_cache": true,
    "max_concurrent_tools": 2,
    "command_timeout": 30
  }
}
```

### Configuração para Produção/Corporativo

```json
{
  "ollama": {
    "url": "http://ollama-server:11434",
    "model": "qwen2.5-coder:32b-instruct-q6_K",
    "temperature": 0.3,
    "max_tokens": 8192
  },
  "app": {
    "mode": "interactive",
    "output_style": "corporate",
    "enable_colors": false,
    "enable_checkpoints": true,
    "enable_sessions": true,
    "enable_memory": true,
    "checkpoint_retention": 90,
    "log_level": "info"
  },
  "performance": {
    "enable_cache": true,
    "max_concurrent_tools": 5,
    "command_timeout": 120
  }
}
```

### Configuração Readonly (Segurança)

```json
{
  "ollama": {
    "url": "http://localhost:11434",
    "model": "qwen2.5-coder:32b-instruct-q6_K"
  },
  "app": {
    "mode": "readonly",
    "enable_checkpoints": false,
    "enable_sessions": false,
    "log_level": "warn"
  },
  "performance": {
    "enable_cache": true,
    "max_concurrent_tools": 1
  }
}
```

## 🔧 Comandos Úteis

### Ver configuração atual
```bash
cat ~/.ollama-code/config.json | jq .
```

### Editar configuração
```bash
# Linux/macOS
nano ~/.ollama-code/config.json

# Windows
notepad %USERPROFILE%\.ollama-code\config.json
```

### Resetar para padrões
```bash
# Apagar arquivo - será recriado automaticamente
rm ~/.ollama-code/config.json
```

### Validar arquivo JSON
```bash
cat ~/.ollama-code/config.json | jq . > /dev/null && echo "✅ JSON válido" || echo "❌ JSON inválido"
```

## ⚠️ Notas Importantes

1. **Formato JSON estrito** - Use vírgulas corretamente, sem trailing commas
2. **Strings entre aspas duplas** - JSON requer `"` não `'`
3. **Booleans em lowercase** - `true`/`false` não `True`/`False`
4. **Números sem aspas** - `999` não `"999"`
5. **Paths em Windows** - Use `\\` ou `/` em paths: `"C:/Users/..."`

## 🐛 Troubleshooting

### Config não está sendo usado
```bash
# Verificar se arquivo existe
ls -la ~/.ollama-code/config.json

# Verificar permissões
chmod 644 ~/.ollama-code/config.json

# Forçar uso de config customizado
ollama-code chat --config ./my-config.json
```

### Erro ao carregar config
```bash
# Validar JSON
cat config.json | jq .

# Ver mensagem de erro detalhada
ollama-code chat --config config.json 2>&1
```

### Valores não estão sendo aplicados
```bash
# Flags CLI sobrescrevem config
# Remova flags para usar valores do arquivo
ollama-code chat  # Sem --mode, --model, etc
```

## 📚 Referência Completa

Veja `config.example.json` na raiz do projeto para um exemplo completo comentado.

---

**Pronto para configurar!** 🚀

Edite `~/.ollama-code/config.json` conforme suas necessidades.
