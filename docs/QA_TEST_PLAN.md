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

### TC-006: Criar Projeto Com Estrutura de Pastas
**Descrição:** Validar criação de projeto com múltiplos arquivos e diretórios
**Comando:**
```bash
./build/ollama-code ask "cria um projeto React completo com estrutura de pastas: src/components, src/pages, src/styles, e arquivos package.json, README.md"
```

**Critérios de Sucesso:**
- [ ] Cria estrutura de diretórios correta
- [ ] Gera múltiplos arquivos em locais apropriados
- [ ] package.json com dependências corretas
- [ ] Componentes em src/components/
- [ ] Arquivos de configuração na raiz
- [ ] README.md com instruções
- [ ] Todos arquivos coerentes entre si

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-007: Criar App Full-Stack (Frontend + Backend)
**Descrição:** Validar criação de aplicação com múltiplas camadas
**Comando:**
```bash
./build/ollama-code ask "cria uma aplicação de todo list com frontend (HTML/CSS/JS) e backend (Node.js/Express) em arquivos separados"
```

**Critérios de Sucesso:**
- [ ] Cria arquivos de frontend: index.html, style.css, app.js
- [ ] Cria arquivos de backend: server.js, package.json
- [ ] API endpoints no backend
- [ ] Frontend faz chamadas para backend
- [ ] Arquivos integrados e funcionais
- [ ] README com instruções de setup

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-008: Adicionar Arquivo a Projeto Existente
**Descrição:** Validar adição de novo arquivo mantendo coerência
**Comandos:**
```bash
# 1. Criar projeto inicial
./build/ollama-code ask "cria um site com index.html e style.css"

# 2. Adicionar novo arquivo
./build/ollama-code ask "adiciona um arquivo script.js com validação de formulário e conecta no index.html"
```

**Critérios de Sucesso:**
- [ ] Cria novo arquivo script.js
- [ ] Adiciona <script> tag no index.html existente
- [ ] JavaScript funcional
- [ ] Mantém código existente do HTML
- [ ] Integração perfeita

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-009: Editar Múltiplos Arquivos Coordenadamente
**Descrição:** Validar edição de vários arquivos relacionados
**Comandos:**
```bash
# 1. Criar projeto
./build/ollama-code ask "cria um blog simples com header.html, footer.html, style.css"

# 2. Modificar todos
./build/ollama-code ask "muda o tema para dark mode em todos os arquivos (HTML e CSS)"
```

**Critérios de Sucesso:**
- [ ] Identifica todos arquivos que precisam mudança
- [ ] Modifica style.css (cores, backgrounds)
- [ ] Atualiza classes nos HTMLs se necessário
- [ ] Mantém consistência visual
- [ ] Não quebra layout existente
- [ ] Todos arquivos em harmonia

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-010: Refatorar Entre Arquivos
**Descrição:** Validar movimentação de código entre arquivos
**Comandos:**
```bash
# 1. Criar arquivo monolítico
./build/ollama-code ask "cria um app.js com todas funções: login, signup, dashboard"

# 2. Refatorar
./build/ollama-code ask "separa as funções em 3 arquivos: auth.js (login/signup), dashboard.js, e atualiza app.js para importar deles"
```

**Critérios de Sucesso:**
- [ ] Cria auth.js com funções de autenticação
- [ ] Cria dashboard.js com funções de dashboard
- [ ] Atualiza app.js com imports
- [ ] Remove código duplicado
- [ ] Mantém funcionalidade
- [ ] Exports/imports corretos

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-011: Criar Projeto Com Dependências Entre Arquivos
**Descrição:** Validar criação onde arquivos dependem uns dos outros
**Comando:**
```bash
./build/ollama-code ask "cria um projeto Python com: main.py, database.py (classe Database), models.py (User model), e utils.py (helper functions). Main importa todos"
```

**Critérios de Sucesso:**
- [ ] Cria 4 arquivos Python
- [ ] database.py tem classe Database
- [ ] models.py usa Database
- [ ] main.py importa todos corretamente
- [ ] Sem imports circulares
- [ ] Código executável sem erros
- [ ] Estrutura modular correta

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-012: Sincronizar Mudanças Em Config Files
**Descrição:** Validar propagação de mudanças em arquivos de configuração
**Comandos:**
```bash
# 1. Criar projeto
./build/ollama-code ask "cria um projeto Node com package.json, .env.example, config.js"

# 2. Adicionar nova dependência
./build/ollama-code ask "adiciona axios como dependência e configura em todos os arquivos necessários"
```

**Critérios de Sucesso:**
- [ ] Adiciona axios no package.json
- [ ] Atualiza config.js se necessário
- [ ] Adiciona variáveis em .env.example se relevante
- [ ] Mantém estrutura de todos arquivos
- [ ] Mudanças coerentes

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 2. Testes de Correção de Bugs

### TC-020: Corrigir Bug Funcional
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

### TC-021: Corrigir Erro de Sintaxe
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

### TC-022: Corrigir Layout/CSS
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

### TC-023: Corrigir Bug em Múltiplos Arquivos
**Descrição:** Validar correção de bug que afeta vários arquivos
**Comandos:**
```bash
# 1. Criar projeto
./build/ollama-code ask "cria um site com index.html que importa style.css e script.js"

# 2. Reportar bug complexo
./build/ollama-code ask "o botão de submit não funciona e as cores estão erradas"
```

**Critérios de Sucesso:**
- [ ] Identifica que problema afeta múltiplos arquivos
- [ ] Corrige JavaScript (event listeners)
- [ ] Corrige CSS (cores)
- [ ] Possivelmente ajusta HTML (se necessário)
- [ ] Todos arquivos sincronizados
- [ ] Bug completamente resolvido

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 3. Testes de Pesquisa Web

### TC-030: Pesquisa de Informação Atual
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

### TC-031: Pesquisa Técnica
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

### TC-032: Distinção: Pesquisa vs Criação
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

### TC-040: Buscar Função
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

### TC-041: Buscar String
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

### TC-050: Analisar Estrutura
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

### TC-051: Análise de Arquitetura
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

### TC-060: Ler Arquivo Existente
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

### TC-061: Editar Arquivo Existente
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

### TC-070: Detecção com Contexto
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

### TC-071: Verbos de Criação
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

### TC-080: Modo Read-Only
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

### TC-081: Modo Interactive (Padrão)
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

### TC-082: Modo Autonomous
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

### TC-090: Referências Anafóricas
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

### TC-091: Continuidade de Conversa
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

### TC-100: Entrada Inválida
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

### TC-101: Arquivo Não Existe
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

### TC-102: Timeout/Lentidão
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

## 11. Testes de Skills Especializados

### TC-110: Research Skill - Pesquisa Avançada
**Descrição:** Validar skill de pesquisa especializado
**Comando:**
```bash
./build/ollama-code ask "use o research skill para comparar React vs Vue vs Angular com prós e contras"
```

**Critérios de Sucesso:**
- [ ] Ativa ResearchSkill corretamente
- [ ] Busca informações de múltiplas fontes
- [ ] Compara tecnologias objetivamente
- [ ] Apresenta prós e contras estruturados
- [ ] Cita fontes de pesquisa

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-111: API Skill - Testar Endpoints
**Descrição:** Validar skill de testes de API
**Comando:**
```bash
./build/ollama-code ask "use o API skill para testar os endpoints da API pública do GitHub (https://api.github.com)"
```

**Critérios de Sucesso:**
- [ ] Ativa APISkill
- [ ] Faz requisições HTTP reais
- [ ] Analisa respostas
- [ ] Reporta status codes
- [ ] Valida JSON responses
- [ ] Identifica problemas se houver

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-112: Code Analysis Skill - Análise Profunda
**Descrição:** Validar skill de análise de código especializado
**Comando:**
```bash
./build/ollama-code ask "use o code analysis skill para analisar complexidade, bugs e segurança do arquivo internal/agent/handlers.go"
```

**Critérios de Sucesso:**
- [ ] Ativa CodeAnalysisSkill
- [ ] Mede complexidade ciclomática
- [ ] Detecta code smells
- [ ] Identifica vulnerabilidades de segurança
- [ ] Sugere refatorações
- [ ] Gera relatório estruturado

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 12. Testes do Sistema OLLAMA.md

### TC-120: Carregar Hierarquia OLLAMA.md
**Descrição:** Validar carregamento de configurações hierárquicas
**Setup:**
```bash
# Criar arquivos OLLAMA.md em diferentes níveis
echo "# Enterprise Rules" > ~/.ollama/OLLAMA.md
echo "- Always use MIT license" >> ~/.ollama/OLLAMA.md

echo "# Project Rules" > OLLAMA.md
echo "- Use Clean Architecture" >> OLLAMA.md

mkdir -p .ollama/go
echo "# Go Rules" > .ollama/go/OLLAMA.md
echo "- Use golangci-lint" >> .ollama/go/OLLAMA.md
```

**Comando:**
```bash
./build/ollama-code chat
> cria um novo projeto Go
```

**Critérios de Sucesso:**
- [ ] Carrega OLLAMA.md de todos os níveis
- [ ] Aplica regras enterprise (MIT license)
- [ ] Aplica regras de projeto (Clean Architecture)
- [ ] Aplica regras de linguagem (golangci-lint)
- [ ] Mostra quantos arquivos OLLAMA.md foram carregados
- [ ] Código gerado segue todas as regras

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-121: Merge de Configurações
**Descrição:** Validar merge correto de configs conflitantes
**Setup:**
```bash
# Enterprise diz uma coisa
echo "# Default: 80% coverage" > ~/.ollama/OLLAMA.md

# Project sobrescreve
echo "# This project: 95% coverage" > OLLAMA.md
```

**Comando:**
```bash
./build/ollama-code ask "qual a cobertura de testes necessária?"
```

**Critérios de Sucesso:**
- [ ] Nível mais específico (Project) sobrescreve enterprise
- [ ] Responde: 95% coverage
- [ ] Usa configuração correta ao gerar código

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-122: OLLAMA.md Por Linguagem
**Descrição:** Validar aplicação de regras específicas por linguagem
**Setup:**
```bash
mkdir -p .ollama/python .ollama/javascript

echo "# Python: Use type hints" > .ollama/python/OLLAMA.md
echo "# JS: Use strict mode" > .ollama/javascript/OLLAMA.md
```

**Comandos:**
```bash
./build/ollama-code ask "cria um script Python"
# Deve ter type hints

./build/ollama-code ask "cria um script JavaScript"
# Deve ter 'use strict'
```

**Critérios de Sucesso:**
- [ ] Python gerado usa type hints
- [ ] JavaScript gerado usa strict mode
- [ ] Regras aplicadas automaticamente por linguagem

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 13. Testes de Git Operations

### TC-130: Git Status e Diff
**Descrição:** Validar operações git básicas
**Comando:**
```bash
./build/ollama-code ask "mostra o git status e diff do projeto"
```

**Critérios de Sucesso:**
- [ ] Executa git status
- [ ] Executa git diff
- [ ] Mostra arquivos modificados
- [ ] Mostra mudanças detalhadas
- [ ] Formata output de forma legível

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-131: Git Commit com Mensagem Inteligente
**Descrição:** Validar criação de commit com mensagem automática
**Comandos:**
```bash
# Fazer alguma mudança
./build/ollama-code ask "adiciona comentário no README.md"

# Commitar
./build/ollama-code ask "cria um commit com essas mudanças"
```

**Critérios de Sucesso:**
- [ ] Analisa mudanças feitas
- [ ] Gera mensagem de commit descritiva
- [ ] Segue Conventional Commits
- [ ] Pede confirmação
- [ ] Executa git add + git commit

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

### TC-132: Git Workflow Completo
**Descrição:** Validar workflow git completo (branch → commit → push)
**Comandos:**
```bash
./build/ollama-code chat --mode interactive
> cria uma nova branch feature/test
> faz mudanças no código
> commita com mensagem apropriada
> faz push da branch
```

**Critérios de Sucesso:**
- [ ] Cria branch corretamente
- [ ] Faz mudanças solicitadas
- [ ] Gera commit com mensagem clara
- [ ] Push para remote funciona
- [ ] Todos passos confirmados pelo usuário

**Resultado:** ⬜ | ✅ | ❌
**Notas:**

---

## 📊 Resumo de Execução

### Estatísticas

| Categoria | Total | Passou | Falhou | Não Testado |
|-----------|-------|--------|--------|-------------|
| Criação de Código | 12 | 0 | 0 | 12 |
| Correção de Bugs | 4 | 0 | 0 | 4 |
| Pesquisa Web | 3 | 0 | 0 | 3 |
| Busca em Código | 2 | 0 | 0 | 2 |
| Análise de Projeto | 2 | 0 | 0 | 2 |
| Leitura/Escrita | 2 | 0 | 0 | 2 |
| Detecção de Intenções | 2 | 0 | 0 | 2 |
| Modos de Operação | 3 | 0 | 0 | 3 |
| Histórico/Contexto | 2 | 0 | 0 | 2 |
| Robustez | 3 | 0 | 0 | 3 |
| Skills Especializados | 3 | 0 | 0 | 3 |
| Sistema OLLAMA.md | 3 | 0 | 0 | 3 |
| Git Operations | 3 | 0 | 0 | 3 |
| **TOTAL** | **44** | **0** | **0** | **44** |

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

## 🎯 Comparação com Claude Code CLI

### ✅ Funcionalidades Implementadas (Paridade)

| Funcionalidade | Ollama Code | Claude Code | Status |
|----------------|-------------|-------------|--------|
| **Criação de Código** | ✅ | ✅ | ✅ Paridade |
| **Edição Inteligente** | ✅ | ✅ | ✅ Paridade |
| **Detecção Contextual** | ✅ | ✅ | ✅ Paridade |
| **Multi-file Operations** | ✅ | ✅ | ✅ Paridade |
| **Web Search** | ✅ | ✅ | ✅ Paridade |
| **Code Search** | ✅ | ✅ | ✅ Paridade |
| **Project Analysis** | ✅ | ✅ | ✅ Paridade |
| **Bug Fixing** | ✅ | ✅ | ✅ Paridade |
| **Skills System** | ✅ Research, API, CodeAnalysis | ✅ | ✅ Paridade |
| **Hierarchical Config** | ✅ OLLAMA.md (4 níveis) | ✅ CLAUDE.md | ✅ Paridade |
| **Git Operations** | ✅ status, diff, commit | ✅ | ✅ Paridade |
| **Modes** | ✅ readonly, interactive, autonomous | ✅ | ✅ Paridade |
| **Privacy** | ✅ 100% Local (Ollama) | ❌ Envia para servers | ✅ **Vantagem** |
| **Cost** | ✅ Grátis | ❌ Pago | ✅ **Vantagem** |
| **Offline** | ✅ Funciona offline | ❌ Requer internet | ✅ **Vantagem** |

### ⚠️ Funcionalidades Parciais

| Funcionalidade | Ollama Code | Claude Code | Gap |
|----------------|-------------|-------------|-----|
| **Test Integration** | ⚠️ Via command_executor | ✅ Integrado | Falta execução automática |
| **Refactoring** | ⚠️ Básico | ✅ Avançado | Falta rename/extract/inline |
| **Debugging** | ⚠️ Via analysis | ✅ Integrado | Falta breakpoints/watch |

### ❌ Funcionalidades Não Implementadas

| Funcionalidade | Prioridade | Impacto | Complexidade |
|----------------|-----------|---------|--------------|
| **MCP Plugin System** | 🔴 Alta | Alto | Alta |
| **Background Tasks** | 🟡 Média | Médio | Média |
| **IDE Integration** | 🟡 Média | Alto | Alta |
| **Real-time Collaboration** | 🟢 Baixa | Médio | Alta |
| **Code Review Features** | 🔴 Alta | Alto | Média |
| **Documentation Generation** | 🟡 Média | Médio | Baixa |
| **Performance Profiling** | 🟢 Baixa | Baixo | Alta |
| **Security Scanning** | 🔴 Alta | Alto | Média |
| **Dependency Management** | 🟡 Média | Médio | Média |
| **CI/CD Integration** | 🟡 Média | Alto | Média |

### 📈 Roadmap para Paridade Completa

#### Fase 1: Funcionalidades Críticas (4-6 semanas)
1. **MCP Plugin System** - Suporte para plugins externos
   - [ ] Arquitetura de plugins
   - [ ] API de integração
   - [ ] Marketplace de plugins

2. **Code Review Features** - Review automatizado
   - [ ] Análise de diffs
   - [ ] Sugestões de melhorias
   - [ ] Checklist automático

3. **Security Scanning** - Detecção de vulnerabilidades
   - [ ] Scan de dependências (CVEs)
   - [ ] Análise de código (SAST)
   - [ ] Secrets detection

#### Fase 2: Produtividade (4-6 semanas)
4. **Test Integration** - Testes automáticos
   - [ ] Auto-detecção de framework de testes
   - [ ] Execução automática após mudanças
   - [ ] Coverage tracking

5. **Advanced Refactoring** - Refatorações complexas
   - [ ] Rename symbol (cross-file)
   - [ ] Extract method/class
   - [ ] Inline variable/method
   - [ ] Move to file

6. **Documentation Generation** - Docs automáticos
   - [ ] JSDoc/GoDoc/Docstrings
   - [ ] README.md generation
   - [ ] API documentation

#### Fase 3: Integrações (6-8 semanas)
7. **IDE Integration** - VS Code, JetBrains
   - [ ] Extension para VS Code
   - [ ] Plugin para IntelliJ
   - [ ] LSP server

8. **CI/CD Integration** - GitHub Actions, GitLab CI
   - [ ] Workflow templates
   - [ ] Auto-fix em PRs
   - [ ] Quality gates

9. **Dependency Management** - Gerenciamento de deps
   - [ ] Auto-update de dependências
   - [ ] Compatibility checking
   - [ ] License compliance

#### Fase 4: Avançado (Opcional)
10. **Background Tasks** - Operações assíncronas
11. **Real-time Collaboration** - Pair programming
12. **Performance Profiling** - Análise de performance

### 🏆 Diferenciadores do Ollama Code

Enquanto busca paridade com Claude Code, Ollama Code já tem vantagens únicas:

1. **100% Local e Privado** 🔒
   - Código proprietário nunca sai da máquina
   - GDPR/LGPD compliant por design
   - Ideal para empresas com dados sensíveis

2. **Grátis e Open Source** 💰
   - Sem custos mensais
   - Sem limites de uso
   - Comunidade pode contribuir

3. **Funciona Offline** ✈️
   - Não precisa de internet após setup
   - Perfeito para avião, cafés sem WiFi
   - Sem latência de rede

4. **Customizável** ⚙️
   - OLLAMA.md totalmente flexível
   - Skills personalizados
   - Modelos Ollama intercambiáveis

5. **Hardware Otimizado** 🚀
   - Auto-detecção de GPU
   - Configuração otimizada automática
   - Performa bem até em máquinas modestas

### 📊 Score de Paridade

**Funcionalidades Core:** 15/15 (100%) ✅
**Funcionalidades Avançadas:** 3/12 (25%) ⚠️
**Integrações:** 0/5 (0%) ❌

**Score Total:** 18/32 (**56% de paridade**)

**Meta para v2.0:** 90% de paridade (29/32 funcionalidades)

---

**Testador:** Claude Code (Assistente AI)
**Data de Criação:** 2024-12-19
**Última Atualização:** 2024-12-19
**Próxima Revisão:** Após implementação da Fase 1
