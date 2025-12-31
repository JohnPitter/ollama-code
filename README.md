# 🤖 Ollama Code - Assistente de Código AI 100% Local

> Seu assistente de programação inteligente que roda completamente no seu computador, sem precisar de internet ou pagar assinaturas!

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![CI/CD](https://github.com/johnpitter/ollama-code/workflows/CI/CD/badge.svg)](https://github.com/johnpitter/ollama-code/actions)
[![Tests](https://img.shields.io/badge/Tests-210+_passing-success)](.)
[![Coverage](https://img.shields.io/badge/Coverage-Codecov-blue)](https://codecov.io/gh/johnpitter/ollama-code)
[![Go Report Card](https://goreportcard.com/badge/github.com/johnpitter/ollama-code)](https://goreportcard.com/report/github.com/johnpitter/ollama-code)

## 📖 Índice

- [O que é?](#-o-que-é)
- [Por que usar?](#-por-que-usar)
- [Instalação Fácil](#-instalação-fácil)
- [Como usar](#-como-usar)
- [Funcionalidades](#-funcionalidades)
- [Configuração](#%EF%B8%8F-configuração)
- [Exemplos Práticos](#-exemplos-práticos)
- [Documentação Completa](#-documentação-completa)
- [Contribuindo](#-contribuindo)

## 🎯 O que é?

Ollama Code é um **assistente de programação com inteligência artificial** que:
- ✅ Roda 100% no seu computador (privacidade total!)
- ✅ Funciona offline para a maioria das operações
- ✅ É grátis e open source
- ✅ Funciona com Ollama (modelos de IA locais)
- ✅ Entende e escreve código em várias linguagens
- ✅ Pesquisa na internet por você (opcional, requer conexão)
- ✅ Analisa seu código e sugere melhorias

## 💡 Por que usar?

### Vs. ChatGPT/Claude/Copilot

| Recurso | Ollama Code | ChatGPT/Claude | GitHub Copilot |
|---------|-------------|----------------|----------------|
| **Privacidade** | ✅ 100% Local | ❌ Envia dados | ❌ Envia dados |
| **Custo** | ✅ Grátis | 💰 $20/mês | 💰 $10/mês |
| **Offline** | ✅ Funciona | ❌ Precisa internet | ❌ Precisa internet |
| **Sem limite** | ✅ Ilimitado | ❌ Limitado | ❌ Limitado |
| **Código proprietário** | ✅ Fica no seu PC | ❌ Vai para servidores | ❌ Vai para servidores |

## 🚀 Instalação Fácil

### Passo 1: Instalar Ollama

**Windows:**
1. Baixe: https://ollama.com/download/windows
2. Execute o instalador
3. Abra o terminal e teste: `ollama --version`

**Linux/Mac:**
```bash
curl -fsSL https://ollama.com/install.sh | sh
```

### Passo 2: Baixar um modelo de IA

Escolha um modelo baseado na sua RAM disponível:

```bash
# Modelo pequeno (4GB RAM) - Rápido, ideal para começar
ollama pull qwen2.5-coder:7b

# Modelo médio (8GB RAM) - Balanceado (Recomendado se tiver RAM)
ollama pull qwen2.5-coder:14b

# Modelo grande (16GB+ RAM) - Mais preciso mas mais lento
ollama pull qwen2.5-coder:32b
```

> **Dica:** Comece com o modelo 7b. Se funcionar bem, experimente o 14b para resultados melhores.

### Passo 3: Instalar Ollama Code

**Opção A: Baixar executável (Mais fácil)**

1. Vá em [Releases](https://github.com/johnpitter/ollama-code/releases)
2. Baixe para seu sistema operacional
3. Coloque em uma pasta no PATH

**Opção B: Compilar do código-fonte**

```bash
# 1. Instalar Go (se não tiver)
# Windows: https://go.dev/dl/
# Linux: sudo apt install golang-go

# 2. Clonar repositório
git clone https://github.com/johnpitter/ollama-code.git
cd ollama-code

# 3. Compilar
chmod +x build.sh
./build.sh

# 4. Testar
./build/ollama-code --version
```

### Passo 4: Primeiro teste!

```bash
./build/ollama-code ask "Como criar uma função que soma dois números em Python?"
```

Se funcionou, você está pronto! 🎉

## 📚 Como usar

Ollama Code tem 3 modos de uso:

### 1. Perguntas rápidas (ask)

Para perguntas pontuais:

```bash
ollama-code ask "Como ler um arquivo JSON em Go?"
ollama-code ask "Qual a diferença entre let e var em JavaScript?"
ollama-code ask "Pesquise na internet sobre Go 1.23"
```

### 2. Chat interativo (chat)

Para conversar e fazer várias perguntas:

```bash
ollama-code chat
```

Dentro do chat:
```
💬 Você: Como criar uma API REST em Go?
🤖 Assistente: [explica...]

💬 Você: Pode me dar um exemplo de código?
🤖 Assistente: [mostra código...]

💬 Você: exit  ← para sair
```

### 3. Modo autônomo (autonomous)

O assistente pode fazer mudanças nos arquivos automaticamente:

```bash
ollama-code chat --mode autonomous
```

⚠️ **Atenção:** Neste modo, o assistente pode modificar seus arquivos sem perguntar!

## ✨ Funcionalidades

### 🌐 Pesquisa na Internet

Ollama Code pode pesquisar na web e trazer informações atualizadas:

```bash
ollama-code ask "Qual a temperatura em São Paulo hoje?"
ollama-code ask "O que há de novo no Python 3.12?"
```

**Como funciona:**
1. Busca no DuckDuckGo
2. Acessa os sites e extrai o conteúdo
3. Resume as informações para você

### 🔧 Skills Especializados

Ollama Code tem habilidades especiais:

**1. Research (Pesquisa)**
- Busca na web
- Compara tecnologias
- Encontra documentação

**2. API**
- Testa endpoints
- Analisa APIs REST
- Faz requisições HTTP

**3. Code Analysis (Análise de Código)**
- Detecta bugs
- Mede complexidade
- Sugere otimizações
- Verifica segurança

### 📝 Sistema OLLAMA.md

Configure o assistente com arquivos OLLAMA.md em 4 níveis:

**1. Enterprise** (~/.ollama/OLLAMA.md)
```markdown
# Padrões da Empresa

- Sempre usar MIT license
- Code review obrigatório
```

**2. Project** (seu-projeto/OLLAMA.md)
```markdown
# Projeto E-commerce

- Usar Clean Architecture
- 80% de cobertura de testes
```

**3. Language** (seu-projeto/.ollama/go/OLLAMA.md)
```markdown
# Convenções Go

- Usar golangci-lint
- Testes table-driven
```

**4. Local** (seu-projeto/api/OLLAMA.md)
```markdown
# API Handlers

- Validar todo input
- Logar todas requisições
```

O assistente segue todas essas regras automaticamente!

## ⚙️ Configuração

### Mudar o modelo de IA

```bash
# Usar modelo maior (mais preciso)
ollama-code chat --model qwen2.5-coder:32b

# Usar modelo menor (mais rápido)
ollama-code chat --model qwen2.5-coder:7b
```

### Modos de operação

```bash
# Modo somente leitura (não altera nada)
ollama-code chat --mode readonly

# Modo interativo (pergunta antes de modificar) ⭐ PADRÃO
ollama-code chat --mode interactive

# Modo autônomo (modifica sem perguntar)
ollama-code chat --mode autonomous
```

### Arquivo de configuração

Crie `~/.ollama-code/config.json`:

```json
{
  "ollama_url": "http://localhost:11434",
  "model": "qwen2.5-coder:7b",
  "mode": "interactive",
  "max_tokens": 4096
}
```

## 💡 Exemplos Práticos

### Exemplo 1: Criar uma função

```bash
$ ollama-code ask "Crie uma função em Python que calcula fatorial"

🤖 Assistente:
Aqui está uma função para calcular fatorial:

```python
def fatorial(n):
    if n == 0 or n == 1:
        return 1
    return n * fatorial(n - 1)

# Uso
print(fatorial(5))  # 120
```
```

### Exemplo 2: Pesquisar na internet

```bash
$ ollama-code ask "Pesquise sobre as novidades do Go 1.23"

🌐 Pesquisando na web: novidades do Go 1.23
📄 Buscando conteúdo de 3 sites...
✓ Conteúdo obtido de https://go.dev/doc/go1.23

🤖 Assistente:
O Go 1.23 trouxe várias novidades:

1. Range over Functions: Agora você pode usar range com funções
2. Generic Type Aliases: Suporte experimental para aliases genéricos
3. Novo pacote iter: Para trabalhar com iteradores
4. Melhorias no compilador: Binários 15% menores
5. Garbage Collector: Pausas 30% mais rápidas

Fonte: go.dev
```

### Exemplo 3: Analisar código

```bash
$ ollama-code ask "Analise este código Go e sugira melhorias..."

🤖 Assistente:
Encontrei alguns pontos de melhoria:

1. Use range em vez de índice
2. Evite comparação com string vazia
3. Adicione documentação
[mostra código melhorado...]
```

## 📚 Documentação Completa

### Guias Principais

- [CLAUDE.md](CLAUDE.md) - Guia completo para desenvolvedores (arquitetura, troubleshooting, padrões)
- [ROADMAP.md](ROADMAP.md) - Roadmap de desenvolvimento e status das features
- [docs/guides/CONTRIBUTING.md](docs/guides/CONTRIBUTING.md) - Como contribuir

### Arquitetura

- [docs/architecture/ARCHITECTURE_REFACTORING.md](docs/architecture/ARCHITECTURE_REFACTORING.md) - Handler Pattern
- [docs/architecture/MANUAL_DI.md](docs/architecture/MANUAL_DI.md) - Dependency Injection
- [docs/architecture/OBSERVABILITY.md](docs/architecture/OBSERVABILITY.md) - Sistema de observabilidade

### Mudanças Recentes

- [Web Search Híbrido](changes/01-web-search-hybrid.md) - Busca real na internet
- [Agent Skills](changes/02-agent-skills.md) - Sistema de habilidades
- [OLLAMA.md](changes/03-ollama-md-system.md) - Configuração hierárquica

### Problemas Comuns

Veja a seção **Performance and Troubleshooting** no [CLAUDE.md](CLAUDE.md#performance-and-troubleshooting) para soluções de:
- GPU sobrecarregada / fallback para CPU
- Respostas lentas do LLM
- Timeouts e travamentos
- Alto uso de memória

## 🛠️ Tecnologias

- **Go 1.21+** - Linguagem principal
- **Ollama** - Modelos de IA locais
- **DuckDuckGo** - Busca na web
- **Cobra** - CLI framework

## 🤝 Contribuindo

Adoramos contribuições!

**Formas de contribuir:**
- 🐛 Reportar bugs
- 💡 Sugerir funcionalidades
- 📝 Melhorar documentação
- 🔧 Enviar pull requests
- ⭐ Dar uma estrela no projeto!

## 📄 Licença

MIT License - Veja [LICENSE](LICENSE)

## 🙏 Agradecimentos

- [Ollama](https://ollama.com) - Por tornar IA local possível
- [awesome-claude-code](https://github.com/hesreallyhim/awesome-claude-code) - Inspiração
- Comunidade Go - Por ferramentas incríveis

---

**Feito com ❤️ e IA local no Brasil 🇧🇷**

⭐ Se você gostou, dê uma estrela no projeto!
