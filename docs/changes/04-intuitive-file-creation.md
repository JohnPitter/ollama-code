# Melhoria: Criação Intuitiva de Arquivos com Linguagem Natural

**Data:** 2024-12-19
**Tipo:** Enhancement
**Componentes:** `internal/agent/handlers.go`

## 📋 Problema

Quando usuários pediam para criar arquivos usando linguagem natural, o sistema falhava com erro "conteúdo não especificado":

```bash
💬 Você: cria uma pagina html e css para divulgar um novo produto financeiro e que tenha um design bonito

🔍 Detectando intenção...
Intenção: write_file (confiança: 95%)

🤖 Assistente:
Erro: conteúdo não especificado
```

**Causa Raiz:**
O handler `handleWriteFile()` esperava que o conteúdo do arquivo já estivesse extraído dos parâmetros da detecção de intenção. Mas quando o usuário pede para "criar" algo, ele está pedindo para o LLM **gerar** o conteúdo, não fornecendo o conteúdo diretamente.

## ✨ Solução Implementada

### 1. Geração Automática de Conteúdo

Quando o conteúdo não é fornecido explicitamente, o sistema agora:

1. **Detecta** que precisa gerar conteúdo
2. **Usa LLM** para gerar o código completo baseado na descrição
3. **Mostra preview** do código gerado
4. **Pede confirmação** (em modo interativo)
5. **Escreve o arquivo**

### 2. Método Principal com JSON

O sistema usa um prompt estruturado que pede ao LLM para retornar JSON:

```go
generationPrompt := `Você é um assistente de programação. O usuário pediu:

"%s"

TAREFA:
1. Identifique o tipo de arquivo que o usuário quer criar
2. Identifique o nome/caminho do arquivo (se não especificado, sugira um apropriado)
3. Gere o conteúdo completo do arquivo conforme solicitado

Responda APENAS com um JSON no seguinte formato:
{
  "file_path": "caminho/do/arquivo.ext",
  "content": "conteúdo completo do arquivo aqui",
  "mode": "create"
}

IMPORTANTE:
- O campo "content" deve conter TODO o código/conteúdo solicitado
- Use boas práticas de código
- Adicione comentários quando apropriado
- Se for HTML/CSS, crie algo visualmente atraente
- Não inclua explicações fora do JSON`
```

### 3. Método Fallback

Se o parsing JSON falhar, há um método alternativo mais simples:
- Pede ao LLM para gerar o conteúdo de forma mais direta
- Extrai o nome do arquivo da primeira linha
- Usa o resto como conteúdo

### 4. Parse JSON Adequado

Substituiu parse manual por `encoding/json`:

```go
func parseJSON(jsonStr string, result *map[string]interface{}) error {
    err := json.Unmarshal([]byte(jsonStr), result)
    if err != nil {
        return fmt.Errorf("failed to parse JSON: %w", err)
    }

    if _, ok := (*result)["file_path"]; !ok {
        return fmt.Errorf("JSON missing required field: file_path")
    }

    return nil
}
```

## 📊 Fluxo de Trabalho Novo

```
Usuário: "cria uma pagina html bonita"
           ↓
Intent Detector: write_file (95%)
           ↓
handleWriteFile: content vazio?
           ↓ SIM
💭 Gerando conteúdo...
           ↓
LLM gera JSON com:
  - file_path: "index.html"
  - content: "<html>...</html>"
  - mode: "create"
           ↓
Parse JSON → extrair campos
           ↓
Preview do conteúdo gerado
           ↓
Confirmação do usuário
           ↓
✓ Arquivo criado!
```

## 🎯 Exemplos de Uso

### Exemplo 1: HTML/CSS
```bash
💬 Você: cria uma landing page moderna para um app de fitness

💭 Gerando conteúdo...

📄 Conteúdo gerado:
Arquivo: landing-page.html
Tamanho: 2.4KB

Preview:
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <title>FitApp - Seu Treino Personalizado</title>
    <style>
        body {
            margin: 0;
            font-family: 'Arial', sans-serif;
            ...
        }
    </style>
</head>
...

Executar? [y/N]: y

✓ Arquivo criado: landing-page.html
```

### Exemplo 2: Python Script
```bash
💬 Você: cria um script python que baixa imagens de uma URL

💭 Gerando conteúdo...

📄 Conteúdo gerado:
Arquivo: download_images.py
Tamanho: 1.2KB

Preview:
#!/usr/bin/env python3
"""
Script para download de imagens de uma URL
"""
import requests
from bs4 import BeautifulSoup
...

Executar? [y/N]: y

✓ Arquivo criado: download_images.py
```

### Exemplo 3: Configuração JSON
```bash
💬 Você: gera um package.json para projeto React com TypeScript

💭 Gerando conteúdo...

📄 Conteúdo gerado:
Arquivo: package.json
Tamanho: 856 bytes

Preview:
{
  "name": "react-typescript-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build"
  },
  ...

Executar? [y/N]: y

✓ Arquivo criado: package.json
```

## 🔧 Mudanças Técnicas

### Arquivos Modificados

1. **`internal/agent/handlers.go`**
   - Linha 59-122: Nova lógica de geração de conteúdo em `handleWriteFile()`
   - Linha 507-520: Função `parseJSON()` com `encoding/json`
   - Linha 522-657: Função `generateAndWriteFileSimple()` (fallback)
   - Linha 659-665: Função `truncate()` helper

### Novas Dependências
- `encoding/json` (stdlib, já disponível)

## ✅ Benefícios

1. **Mais Intuitivo**: Usuários podem pedir para criar arquivos naturalmente
2. **Menos Passos**: Não precisa especificar conteúdo manualmente
3. **Código de Qualidade**: LLM gera código seguindo boas práticas
4. **Preview Antes de Salvar**: Usuário vê o que será criado
5. **Fallback Robusto**: Se JSON falhar, há método alternativo

## 🧪 Testes Recomendados

Para testar a funcionalidade:

```bash
# 1. Compilar
./build.sh

# 2. Testar criação de HTML
./build/ollama-code ask "cria uma pagina html simples com header e footer"

# 3. Testar criação de Python
./build/ollama-code ask "gera um script python que lista arquivos"

# 4. Testar criação de JSON
./build/ollama-code ask "cria um config.json para minha aplicação"

# 5. Testar em modo autônomo
./build/ollama-code chat --mode autonomous
> cria 3 arquivos: index.html, style.css e script.js para uma calculadora
```

## 📝 Notas

- Requer modelo Ollama capaz de gerar código (ex: qwen2.5-coder)
- Temperatura 0.7 para balancear criatividade e consistência
- MaxTokens 3000 para suportar arquivos grandes
- Preview limitado a 500 chars para não poluir terminal
- Funciona com todos os modos: readonly (bloqueia), interactive (confirma), autonomous (automático)

## 🚀 Próximos Passos

Possíveis melhorias futuras:
- [ ] Suporte para múltiplos arquivos em uma solicitação
- [ ] Templates predefinidos (ex: "cria projeto React completo")
- [ ] Validação de sintaxe antes de salvar
- [ ] Opção de editar conteúdo gerado antes de salvar
- [ ] Cache de prompts de geração comuns

---

**Feedback do Usuário:**
> "tem algumas coisas que nao estao funcionando de forma tao intuitiva, refina mais"

**Resultado:** ✅ Sistema agora suporta criação intuitiva de arquivos com linguagem natural
