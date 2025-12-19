# 🧪 Plano de Testes QA - Ollama Code

**Versão:** 1.0.0
**Data:** 2024-12-19
**Objetivo:** Validar todas as funcionalidades da aplicação como QA profissional

## 📋 Índice

1. [Testes de Criação de Código](#1-testes-de-criação-de-código)
2. [Testes de Correção de Bugs](#2-testes-de-correção-de-bugs)
3. [Testes de Pesquisa Web](#3-testes-de-pesquisa-web)
4. [Testes de Busca em Código](#4-testes-de-busca-em-código)
5. [Testes de Análise de Projeto](#5-testes-de-análise-de-projeto)
6. [Testes de Leitura/Escrita](#6-testes-de-leituraescrita)
7. [Testes de Detecção de Intenções](#7-testes-de-detecção-de-intenções)
8. [Testes de Modos de Operação](#8-testes-de-modos-de-operação)
9. [Testes de Histórico e Contexto](#9-testes-de-histórico-e-contexto)
10. [Testes de Robustez](#10-testes-de-robustez)

---

## 1. Testes de Criação de Código

### TC-001: Criar Arquivo HTML Simples
**Descrição:** Validar criação de arquivo HTML básico
**Comando:**
```bash
./build/ollama-code ask "cria um arquivo HTML com header, nav e footer"
```

**Critérios de Sucesso:**
- [ ] Detecta intenção: `write_file`
- [ ] Gera arquivo .html com estrutura solicitada
- [ ] Código HTML é válido (DOCTYPE, head, body)
- [ ] Inclui elementos solicitados (header, nav, footer)
- [ ] Pede confirmação em modo interactive
- [ ] Registra arquivo em recentFiles

**Resultado:** ⬜ Não Testado | ✅ Passou | ❌ Falhou
**Notas:**

---

### TC-002: Criar Arquivo CSS
**Descrição:** Validar criação de CSS com estilos específicos
**Comando:**
```bash
./build/ollama-code ask "cria um arquivo CSS com estilo moderno, dark mode e responsivo"
```

**Critérios de Sucesso:**
- [ ] Gera arquivo .css
- [ ] Inclui media queries para responsividade
- [ ] Implementa dark mode
- [ ] CSS é válido (sem erros de sintaxe)

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-003: Criar Script Python
**Descrição:** Validar criação de script Python funcional
**Comando:**
```bash
./build/ollama-code ask "gera um script python que lê CSV e calcula médias"
```

**Critérios de Sucesso:**
- [ ] Gera arquivo .py
- [ ] Código Python sintaticamente correto
- [ ] Inclui imports necessários (csv, pandas, etc)
- [ ] Implementa lógica solicitada
- [ ] Inclui tratamento de erros
- [ ] Código é executável

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-004: Criar Projeto Multi-Arquivo
**Descrição:** Validar criação de projeto com múltiplos arquivos relacionados
**Comando:**
```bash
./build/ollama-code ask "cria uma landing page completa com HTML, CSS e JavaScript separados"
```

**Critérios de Sucesso:**
- [ ] Gera 3 arquivos: .html, .css, .js
- [ ] Arquivos estão corretamente linkados
- [ ] Cada arquivo tem conteúdo apropriado
- [ ] JavaScript funciona com HTML
- [ ] CSS estiliza corretamente

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-005: Criar Código Complexo
**Descrição:** Validar criação de código com lógica não-trivial
**Comando:**
```bash
./build/ollama-code ask "desenvolve uma API REST em Go com endpoints CRUD para usuários"
```

**Critérios de Sucesso:**
- [ ] Gera arquivo .go válido
- [ ] Inclui imports necessários
- [ ] Implementa todos endpoints (GET, POST, PUT, DELETE)
- [ ] Código compila sem erros
- [ ] Segue boas práticas de Go

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 2. Testes de Correção de Bugs

### TC-010: Corrigir Bug Funcional
**Descrição:** Validar correção de bug em arquivo recém-criado
**Comandos:**
```bash
# 1. Criar arquivo com bug intencional
./build/ollama-code ask "cria uma calculadora HTML mas sem eventos nos botões"

# 2. Reportar problema
./build/ollama-code ask "os botões não funcionam quando clico"
```

**Critérios de Sucesso:**
- [ ] Detecta que é correção de bug
- [ ] Identifica arquivo recente (calculadora.html)
- [ ] Lê arquivo atual
- [ ] Analisa problema corretamente
- [ ] Adiciona event listeners
- [ ] Mostra análise e correções
- [ ] Sobrescreve arquivo (não cria novo)

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-011: Corrigir Erro de Sintaxe
**Descrição:** Validar correção de erro de código
**Comandos:**
```bash
# 1. Criar
./build/ollama-code ask "faz um script que lista arquivos"

# 2. Reportar erro
./build/ollama-code ask "deu erro: NameError name 'os' is not defined"
```

**Critérios de Sucesso:**
- [ ] Detecta como bug fix
- [ ] Adiciona import faltante
- [ ] Corrige erro específico
- [ ] Explica o que foi corrigido

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-012: Corrigir Layout/CSS
**Descrição:** Validar correção de problemas visuais
**Comandos:**
```bash
# 1. Criar
./build/ollama-code ask "cria uma galeria de imagens responsiva"

# 2. Reportar
./build/ollama-code ask "o layout quebrou no mobile"
```

**Critérios de Sucesso:**
- [ ] Detecta problema de layout
- [ ] Adiciona/ajusta media queries
- [ ] Testa responsividade
- [ ] Grid/Flexbox corrigido

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 3. Testes de Pesquisa Web

### TC-020: Pesquisa de Informação Atual
**Descrição:** Validar busca de dados em tempo real
**Comando:**
```bash
./build/ollama-code ask "qual a temperatura atual em São Paulo"
```

**Critérios de Sucesso:**
- [ ] Detecta intenção: `web_search`
- [ ] Busca no DuckDuckGo
- [ ] Obtém conteúdo de sites
- [ ] Retorna temperatura atualizada
- [ ] Cita fontes

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-021: Pesquisa Técnica
**Descrição:** Validar busca de documentação técnica
**Comando:**
```bash
./build/ollama-code ask "pesquise as novidades do Python 3.12 na internet"
```

**Critérios de Sucesso:**
- [ ] Busca informações técnicas
- [ ] Acessa documentação oficial
- [ ] Resume novidades principais
- [ ] Informações corretas e atualizadas

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-022: Distinção: Pesquisa vs Criação
**Descrição:** Validar que não confunde pesquisa com criação
**Comandos:**
```bash
# Deve ser web_search
./build/ollama-code ask "pesquise tutoriais sobre React"

# Deve ser write_file
./build/ollama-code ask "cria um componente React"
```

**Critérios de Sucesso:**
- [ ] Primeiro comando: detecta web_search
- [ ] Segundo comando: detecta write_file
- [ ] Não cria arquivo no primeiro
- [ ] Não busca web no segundo

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 4. Testes de Busca em Código

### TC-030: Buscar Função
**Descrição:** Validar busca por função específica
**Comando:**
```bash
./build/ollama-code ask "busca a função handleWriteFile no código"
```

**Critérios de Sucesso:**
- [ ] Detecta intenção: `search_code`
- [ ] Executa code_searcher tool
- [ ] Retorna arquivos onde função aparece
- [ ] Mostra linha e trecho de código
- [ ] Limita resultados (top 10)

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-031: Buscar String
**Descrição:** Validar busca por string/padrão
**Comando:**
```bash
./build/ollama-code ask "procure por 'database connection' no projeto"
```

**Critérios de Sucesso:**
- [ ] Busca string literal
- [ ] Retorna todos matches
- [ ] Mostra contexto (linhas ao redor)

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 5. Testes de Análise de Projeto

### TC-040: Analisar Estrutura
**Descrição:** Validar análise completa do projeto
**Comando:**
```bash
./build/ollama-code ask "analisa este projeto"
```

**Critérios de Sucesso:**
- [ ] Detecta intenção: `analyze_project`
- [ ] Mostra nome do projeto
- [ ] Conta arquivos e diretórios
- [ ] Detecta linguagens usadas
- [ ] Mostra estrutura de pastas
- [ ] Informações corretas

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-041: Análise de Arquitetura
**Descrição:** Validar entendimento da arquitetura
**Comando:**
```bash
./build/ollama-code ask "qual a arquitetura deste projeto e como os componentes se relacionam"
```

**Critérios de Sucesso:**
- [ ] Identifica padrões arquiteturais
- [ ] Explica módulos principais
- [ ] Descreve relacionamentos
- [ ] Resposta coerente

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 6. Testes de Leitura/Escrita

### TC-050: Ler Arquivo Existente
**Descrição:** Validar leitura de arquivo específico
**Comando:**
```bash
./build/ollama-code ask "leia o arquivo README.md"
```

**Critérios de Sucesso:**
- [ ] Detecta intenção: `read_file`
- [ ] Executa file_reader tool
- [ ] Retorna conteúdo do arquivo
- [ ] Formata output adequadamente

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-051: Editar Arquivo Existente
**Descrição:** Validar edição de arquivo com replace
**Comando:**
```bash
./build/ollama-code ask "adiciona um novo método no arquivo agent.go"
```

**Critérios de Sucesso:**
- [ ] Lê arquivo atual
- [ ] Adiciona código no local correto
- [ ] Mantém código existente
- [ ] Não quebra sintaxe

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 7. Testes de Detecção de Intenções

### TC-060: Detecção com Contexto
**Descrição:** Validar uso de histórico para decisão
**Comandos:**
```bash
# Estabelecer contexto
./build/ollama-code chat
> quero criar meu próprio site de receitas
> desenvolve um usando HTML e CSS
```

**Critérios de Sucesso:**
- [ ] Segunda mensagem usa contexto da primeira
- [ ] Detecta write_file (não web_search)
- [ ] Gera site de receitas (não genérico)

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-061: Verbos de Criação
**Descrição:** Validar reconhecimento de múltiplos verbos
**Comandos:**
```bash
# Testar cada verbo
"cria um formulário"      → write_file
"desenvolve uma API"      → write_file
"faz um script"           → write_file
"gera um componente"      → write_file
"constrói uma app"        → write_file
"escreve uma função"      → write_file
"implementa um CRUD"      → write_file
```

**Critérios de Sucesso:**
- [ ] Todos detectados como write_file
- [ ] Nenhum detectado como web_search

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 8. Testes de Modos de Operação

### TC-070: Modo Read-Only
**Descrição:** Validar que bloqueia escritas
**Comando:**
```bash
./build/ollama-code chat --mode readonly
> cria um arquivo test.txt
```

**Critérios de Sucesso:**
- [ ] Detecta write_file
- [ ] Bloqueia operação
- [ ] Mostra mensagem: "modo somente leitura"
- [ ] NÃO cria arquivo

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-071: Modo Interactive (Padrão)
**Descrição:** Validar confirmações
**Comando:**
```bash
./build/ollama-code chat
> cria um arquivo test.txt
```

**Critérios de Sucesso:**
- [ ] Gera conteúdo
- [ ] Mostra preview
- [ ] Pede confirmação
- [ ] Aguarda resposta (s/n)
- [ ] Só cria se confirmar

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-072: Modo Autonomous
**Descrição:** Validar execução automática
**Comando:**
```bash
./build/ollama-code chat --mode autonomous
> cria 3 arquivos: index.html, style.css, script.js
```

**Critérios de Sucesso:**
- [ ] NÃO pede confirmação
- [ ] Cria todos arquivos automaticamente
- [ ] Executa sem intervenção

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 9. Testes de Histórico e Contexto

### TC-080: Referências Anafóricas
**Descrição:** Validar resolução de pronomes/referências
**Comandos:**
```bash
./build/ollama-code chat
> cria um site de portfólio
> adiciona uma seção de contato nele
> muda a cor de fundo dele para azul
```

**Critérios de Sucesso:**
- [ ] "nele" refere-se ao site criado
- [ ] "dele" também refere-se ao site
- [ ] Modifica arquivo correto
- [ ] Não cria arquivos novos

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-081: Continuidade de Conversa
**Descrição:** Validar manutenção de contexto longo
**Comandos:**
```bash
./build/ollama-code chat
> cria uma calculadora
> [... várias mensagens depois ...]
> volta para a calculadora e adiciona função de raiz quadrada
```

**Critérios de Sucesso:**
- [ ] Sistema lembra da calculadora
- [ ] Edita arquivo correto
- [ ] Adiciona funcionalidade solicitada

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 10. Testes de Robustez

### TC-090: Entrada Inválida
**Descrição:** Validar tratamento de erros
**Comando:**
```bash
./build/ollama-code ask ""  # Vazio
./build/ollama-code ask "xkjdflajsdflkjasdflkjasd"  # Gibberish
```

**Critérios de Sucesso:**
- [ ] Não crasha
- [ ] Retorna mensagem amigável
- [ ] Pede clarificação ou assume question

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-091: Arquivo Não Existe
**Descrição:** Validar erro ao ler arquivo inexistente
**Comando:**
```bash
./build/ollama-code ask "leia o arquivo naoexiste.txt"
```

**Critérios de Sucesso:**
- [ ] Não crasha
- [ ] Retorna erro claro
- [ ] Sugere verificar caminho

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-092: Timeout/Lentidão
**Descrição:** Validar comportamento com LLM lento
**Comando:**
```bash
./build/ollama-code ask "gera um arquivo gigante com 1000 linhas de código"
```

**Critérios de Sucesso:**
- [ ] Não trava indefinidamente
- [ ] Mostra progresso ou aguarda
- [ ] Respeita timeout se configurado
- [ ] Retorna resultado ou erro

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 📊 Resumo de Execução

### Estatísticas

| Categoria | Total | Passou | Falhou | Não Testado |
|-----------|-------|--------|--------|-------------|
| Criação de Código | 5 | 0 | 0 | 5 |
| Correção de Bugs | 3 | 0 | 0 | 3 |
| Pesquisa Web | 3 | 0 | 0 | 3 |
| Busca em Código | 2 | 0 | 0 | 2 |
| Análise de Projeto | 2 | 0 | 0 | 2 |
| Leitura/Escrita | 2 | 0 | 0 | 2 |
| Detecção de Intenções | 2 | 0 | 0 | 2 |
| Modos de Operação | 3 | 0 | 0 | 3 |
| Histórico/Contexto | 2 | 0 | 0 | 2 |
| Robustez | 3 | 0 | 0 | 3 |
| **TOTAL** | **27** | **0** | **0** | **27** |

### Bugs Encontrados

1. [Bug #001] Descrição do bug
   - Severidade: Alta/Média/Baixa
   - Passos para Reproduzir: ...
   - Comportamento Esperado: ...
   - Comportamento Atual: ...

---

## 🔧 Ambiente de Teste

- **OS:** Windows 11 / Linux / macOS
- **Go Version:** 1.21+
- **Ollama Version:** Latest
- **Modelo:** qwen2.5-coder:7b
- **Build:** `./build.sh` em [data]

---

## ✅ Critérios de Aprovação

Para considerar a aplicação pronta para produção:

- [ ] 100% dos casos de teste executados
- [ ] ≥ 95% de taxa de sucesso
- [ ] 0 bugs críticos
- [ ] ≤ 2 bugs médios
- [ ] Todos bugs documentados no GitHub Issues

---

## 📝 Próximos Passos

1. [ ] Executar todos os casos de teste
2. [ ] Documentar resultados detalhadamente
3. [ ] Reportar bugs encontrados
4. [ ] Criar testes automatizados para regressions
5. [ ] Preparar relatório final de QA

---

**Testador:** Claude Code (Assistente AI)
**Data de Criação:** 2024-12-19
**Última Atualização:** 2024-12-19
