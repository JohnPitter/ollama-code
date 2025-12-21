# 🧪 Execução de Testes QA - Ollama Code

**Data de Execução:** 2024-12-19
**Testador:** A definir
**Build:** `./build/ollama-code` (compilado em 2024-12-19)
**Modelo Ollama:** qwen2.5-coder:7b

---

## 📋 Instruções de Execução

### Preparação

1. **Compilar a aplicação:**
```bash
cd /c/Users/joaop/Desenvolvimento/ollama-code
./build.sh
```

2. **Verificar que Ollama está rodando:**
```bash
ollama list
# Deve mostrar qwen2.5-coder:7b
```

3. **Criar diretório de teste:**
```bash
mkdir -p test-execution
cd test-execution
```

4. **Iniciar sessão de testes:**
```bash
# Anotar horário de início
date
```

---

## ✅ Casos de Teste Prioritários

### 🔥 Testes Críticos (Executar Primeiro)

#### TC-001: Criar Arquivo HTML Simples
**Status:** ⬜ Não Executado | ⏳ Em Execução | ✅ Passou | ❌ Falhou

**Comando:**
```bash
../build/ollama-code ask "cria um arquivo HTML com header, nav e footer"
```

**Checklist:**
- [ ] Detectou intenção: `write_file`?
- [ ] Gerou arquivo .html?
- [ ] HTML tem DOCTYPE, head, body?
- [ ] Inclui header, nav, footer?
- [ ] Pediu confirmação (modo interactive)?
- [ ] Arquivo foi criado com sucesso?

**Resultado:**
```
[Copie e cole aqui o output completo do comando]
```

**Arquivo Gerado:** [nome do arquivo]
**Observações:**

---

#### TC-020: Corrigir Bug Funcional
**Status:** ⬜ | ⏳ | ✅ | ❌

**Passo 1 - Criar arquivo:**
```bash
../build/ollama-code ask "cria uma calculadora HTML mas sem eventos nos botões"
```

**Verificar:** Arquivo criado sem event listeners?

**Passo 2 - Reportar bug:**
```bash
../build/ollama-code ask "os botões não funcionam quando clico"
```

**Checklist:**
- [ ] Detectou como correção de bug?
- [ ] Identificou arquivo recente (calculadora.html)?
- [ ] Leu arquivo atual?
- [ ] Analisou problema?
- [ ] Adicionou event listeners?
- [ ] Mostrou análise e correções?
- [ ] Sobrescreveu arquivo (não criou novo)?

**Resultado Passo 1:**
```
[Output do passo 1]
```

**Resultado Passo 2:**
```
[Output do passo 2]
```

**Observações:**

---

#### TC-030: Pesquisa Web
**Status:** ⬜ | ⏳ | ✅ | ❌

**Comando:**
```bash
../build/ollama-code ask "qual a temperatura atual em São Paulo"
```

**Checklist:**
- [ ] Detectou intenção: `web_search`?
- [ ] Buscou no DuckDuckGo?
- [ ] Obteve conteúdo de sites?
- [ ] Retornou temperatura atualizada?
- [ ] Citou fontes?

**Resultado:**
```
[Output completo]
```

**Temperatura Reportada:** ___°C
**Fontes Citadas:**
1.
2.
3.

**Observações:**

---

#### TC-070: Detecção com Contexto
**Status:** ⬜ | ⏳ | ✅ | ❌

**Passo 1:**
```bash
../build/ollama-code chat
> quero criar meu próprio site de receitas
```

**Passo 2:**
```bash
> desenvolve um usando HTML e CSS
```

**Checklist:**
- [ ] Segunda mensagem usou contexto da primeira?
- [ ] Detectou `write_file` (não web_search)?
- [ ] Gerou site de RECEITAS (não genérico)?
- [ ] Conteúdo condiz com "site de receitas"?

**Resultado:**
```
[Copie toda a conversa]
```

**Observações:**

---

#### TC-004: Criar Projeto Multi-Arquivo
**Status:** ⬜ | ⏳ | ✅ | ❌

**Comando:**
```bash
../build/ollama-code ask "cria uma landing page completa com HTML, CSS e JavaScript separados"
```

**Checklist:**
- [ ] Gerou 3 arquivos (`.html`, `.css`, `.js`)?
- [ ] Arquivos estão linkados (link tag, script tag)?
- [ ] HTML referencia CSS e JS corretamente?
- [ ] CSS tem estilos aplicáveis?
- [ ] JavaScript é funcional?

**Resultado:**
```
[Output]
```

**Arquivos Gerados:**
1.
2.
3.

**Verificação de Links:**
- [ ] HTML inclui `<link rel="stylesheet" href="...">`?
- [ ] HTML inclui `<script src="...">`?

**Observações:**

---

### 🎯 Testes Adicionais (Se Tempo Permitir)

#### TC-032: Distinção Pesquisa vs Criação
**Status:** ⬜ | ⏳ | ✅ | ❌

**Teste A (deve ser web_search):**
```bash
../build/ollama-code ask "pesquise tutoriais sobre React"
```
Intenção detectada: ____________

**Teste B (deve ser write_file):**
```bash
../build/ollama-code ask "cria um componente React"
```
Intenção detectada: ____________

**Passou?** [ ] Sim [ ] Não

**Observações:**

---

#### TC-080: Modo Read-Only
**Status:** ⬜ | ⏳ | ✅ | ❌

**Comando:**
```bash
../build/ollama-code chat --mode readonly
> cria um arquivo test.txt
```

**Checklist:**
- [ ] Detectou `write_file`?
- [ ] Bloqueou operação?
- [ ] Mostrou mensagem de modo readonly?
- [ ] NÃO criou arquivo?

**Resultado:**
```
[Output]
```

**Observações:**

---

#### TC-006: Criar Projeto Com Estrutura de Pastas
**Status:** ⬜ | ⏳ | ✅ | ❌

**Comando:**
```bash
../build/ollama-code ask "cria um projeto React completo com estrutura de pastas: src/components, src/pages, src/styles, e arquivos package.json, README.md"
```

**Checklist:**
- [ ] Criou estrutura de diretórios?
- [ ] Gerou múltiplos arquivos nos lugares certos?
- [ ] package.json com dependências?
- [ ] Componentes em src/components/?
- [ ] README.md criado?
- [ ] Arquivos coerentes entre si?

**Resultado:**
```
[Output]
```

**Estrutura Criada:**
```
[Executar: tree ou ls -R para ver estrutura]
```

**Observações:**

---

## 📊 Resumo de Execução

**Data de Execução:** _______________
**Horário Início:** _______________
**Horário Fim:** _______________
**Duração Total:** _______________

### Estatísticas

| Teste | ID | Status | Observações |
|-------|----|----|-------------|
| Criar HTML Simples | TC-001 | ⬜ | |
| Corrigir Bug | TC-020 | ⬜ | |
| Pesquisa Web | TC-030 | ⬜ | |
| Detecção Contexto | TC-070 | ⬜ | |
| Projeto Multi-Arquivo | TC-004 | ⬜ | |
| Distinção Pesquisa/Criação | TC-032 | ⬜ | |
| Modo Read-Only | TC-080 | ⬜ | |
| Estrutura de Pastas | TC-006 | ⬜ | |

**Total Executados:** ___ / 8
**Total Passou:** ___ / 8
**Total Falhou:** ___ / 8
**Taxa de Sucesso:** ____%

---

## 🐛 Bugs Encontrados

### Bug #001
**Título:**
**Severidade:** 🔴 Alta | 🟡 Média | 🟢 Baixa
**Teste:** TC-___
**Descrição:**

**Passos para Reproduzir:**
1.
2.
3.

**Comportamento Esperado:**

**Comportamento Atual:**

**Logs/Screenshots:**
```
```

**Ação Necessária:**
- [ ] Criar issue no GitHub
- [ ] Corrigir imediatamente
- [ ] Adicionar ao backlog

---

### Bug #002
(Repetir template acima para cada bug encontrado)

---

## ✅ Conclusão

### Pontos Positivos
-
-
-

### Pontos de Melhoria
-
-
-

### Recomendações
-
-
-

### Próximos Passos
- [ ]
- [ ]
- [ ]

---

## 📝 Notas Adicionais

**Observações Gerais:**


**Performance:**
- Tempo médio de resposta: ___ segundos
- Uso de memória: OK / Alto / Crítico
- CPU durante execução: ___%

**Usabilidade:**
- Interface clara: Sim / Não
- Mensagens de erro úteis: Sim / Não
- Confirmações apropriadas: Sim / Não

---

**Assinatura do Testador:** _______________
**Data:** _______________
