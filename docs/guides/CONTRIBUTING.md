# 🤝 Contributing to Ollama Code

Obrigado por considerar contribuir com o Ollama Code! Este documento fornece diretrizes para contribuições.

## 📋 Índice

- [Código de Conduta](#código-de-conduta)
- [Como Contribuir](#como-contribuir)
- [Desenvolvimento Local](#desenvolvimento-local)
- [Padrões de Código](#padrões-de-código)
- [Testes](#testes)
- [Pull Requests](#pull-requests)

## 📜 Código de Conduta

- Seja respeitoso e inclusivo
- Aceite críticas construtivas
- Foque no que é melhor para a comunidade
- Mostre empatia com outros contribuidores

## 🚀 Como Contribuir

### Reportar Bugs

Encontrou um bug? Crie uma [issue](https://github.com/johnpitter/ollama-code/issues) com:

- **Título claro:** Descreva o problema em poucas palavras
- **Descrição:** Explique o bug em detalhes
- **Reprodução:** Passo a passo para reproduzir
- **Esperado vs Atual:** O que deveria acontecer vs o que acontece
- **Ambiente:** SO, versão do Go, modelo Ollama usado
- **Logs:** Cole logs relevantes (se houver)

**Exemplo:**
```markdown
**Bug:** Web search retorna erro 404

**Passos:**
1. Execute `ollama-code ask "pesquisar sobre Go"`
2. Aguarde busca

**Esperado:** Retornar resultados
**Atual:** Erro 404

**Ambiente:**
- OS: Windows 11
- Go: 1.21
- Modelo: qwen2.5-coder:7b
```

### Sugerir Funcionalidades

Tem uma ideia? Crie uma [issue](https://github.com/johnpitter/ollama-code/issues) com:

- **Título:** "Feature: [nome da funcionalidade]"
- **Descrição:** O que você quer adicionar
- **Motivação:** Por que isso seria útil
- **Exemplos:** Como funcionaria na prática

### Melhorar Documentação

Documentação sempre pode melhorar:

- Corrigir erros de digitação
- Adicionar exemplos
- Melhorar explicações
- Traduzir para outros idiomas

## 💻 Desenvolvimento Local

### Setup Inicial

```bash
# 1. Fork o repositório no GitHub

# 2. Clone seu fork
git clone https://github.com/SEU-USUARIO/ollama-code.git
cd ollama-code

# 3. Adicione o upstream
git remote add upstream https://github.com/johnpitter/ollama-code.git

# 4. Instale dependências
go mod download

# 5. Compile
./build.sh

# 6. Execute testes
go test ./...
```

### Workflow de Desenvolvimento

```bash
# 1. Crie uma branch para sua feature
git checkout -b feature/minha-feature

# 2. Faça suas mudanças

# 3. Adicione testes
go test ./internal/...

# 4. Execute linter
go vet ./...
golangci-lint run  # se tiver instalado

# 5. Compile e teste
./build.sh
./build/ollama-code ask "teste"

# 6. Commit suas mudanças
git add .
git commit -m "feat: Adiciona funcionalidade X"

# 7. Push para seu fork
git push origin feature/minha-feature

# 8. Abra um Pull Request no GitHub
```

## 📏 Padrões de Código

### Estilo Go

Seguimos o [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments):

```go
// ✅ BOM
func ProcessData(data []string) error {
    if len(data) == 0 {
        return fmt.Errorf("data is empty")
    }
    // ...
}

// ❌ RUIM
func processData(d []string) error {  // nome exportado deve começar com maiúscula
    if len(d)==0{  // falta espaços
        return errors.New("data is empty")  // use fmt.Errorf
    }
}
```

### Nomenclatura

```go
// Packages: minúsculas, uma palavra
package websearch  // ✅
package webSearch  // ❌

// Interfaces: -er suffix
type Reader interface {}  // ✅
type ReadInterface interface {}  // ❌

// Errors: Err prefix
var ErrNotFound = errors.New("not found")  // ✅
var NotFoundError = errors.New("not found")  // ❌
```

### Documentação

Toda função/tipo exportado deve ter comentário:

```go
// ProcessRequest processes an HTTP request and returns a response.
// It returns an error if the request is malformed.
func ProcessRequest(req *http.Request) (*Response, error) {
    // ...
}
```

### Tratamento de Erros

```go
// ✅ BOM - wrap errors com contexto
if err != nil {
    return fmt.Errorf("failed to process data: %w", err)
}

// ❌ RUIM - perde contexto
if err != nil {
    return err
}
```

## 🧪 Testes

### Escrevendo Testes

```go
func TestMyFunction(t *testing.T) {
    // Arrange
    input := "test data"
    expected := "expected result"

    // Act
    result := MyFunction(input)

    // Assert
    if result != expected {
        t.Errorf("MyFunction(%q) = %q, want %q", input, result, expected)
    }
}
```

### Table-Driven Tests

```go
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "hello", false},
        {"empty input", "", true},
        {"too long", strings.Repeat("a", 1000), true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Cobertura de Testes

```bash
# Executar testes com coverage
go test -coverprofile=coverage.out ./...

# Ver relatório
go tool cover -html=coverage.out

# Meta: >80% coverage para novos códigos
```

## 📝 Pull Requests

### Checklist

Antes de abrir um PR, verifique:

- [ ] Código compila sem erros
- [ ] Todos os testes passam
- [ ] Adicionei testes para novo código
- [ ] Documentação atualizada
- [ ] Mensagem de commit segue padrões
- [ ] Branch está atualizada com main

### Mensagens de Commit

Usamos [Conventional Commits](https://www.conventionalcommits.org/):

```
<tipo>[escopo opcional]: <descrição>

[corpo opcional]

[rodapé opcional]
```

**Tipos:**
- `feat`: Nova funcionalidade
- `fix`: Correção de bug
- `docs`: Apenas documentação
- `style`: Formatação, não afeta código
- `refactor`: Refatoração de código
- `test`: Adiciona/corrige testes
- `chore`: Tarefas de build, dependências

**Exemplos:**
```bash
feat: Adiciona suporte para GPT-4
fix: Corrige loop infinito no web search
docs: Atualiza README com novos exemplos
test: Adiciona testes para APISkill
```

### Processo de Review

1. **Abra o PR** com descrição clara
2. **Aguarde review** (pode demorar alguns dias)
3. **Responda comentários** e faça ajustes
4. **Aprovação** por mantenedor
5. **Merge** quando tudo estiver OK

### Template de PR

```markdown
## Descrição
Breve descrição das mudanças

## Motivação
Por que essa mudança é necessária

## Tipo de Mudança
- [ ] Bug fix
- [ ] Nova funcionalidade
- [ ] Breaking change
- [ ] Documentação

## Como Testar
Passos para testar as mudanças

## Checklist
- [ ] Testes passam
- [ ] Documentação atualizada
- [ ] Sem warnings de linter
```

## 🎯 Áreas para Contribuir

### Fácil (Good First Issue)

- Corrigir typos na documentação
- Adicionar exemplos no README
- Melhorar mensagens de erro
- Adicionar testes unitários

### Médio

- Implementar novos Skills
- Melhorar web search
- Adicionar suporte para mais modelos
- Otimizar performance

### Difícil

- Sistema de plugins
- Interface gráfica
- Integração com IDEs
- Workflow orchestration

## 📚 Recursos

- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Ollama Documentation](https://github.com/ollama/ollama)
- [Project Issues](https://github.com/johnpitter/ollama-code/issues)

## 💬 Comunicação

- **Issues:** Para bugs e features
- **Discussions:** Para perguntas gerais
- **Pull Requests:** Para contribuições de código

## 🙏 Agradecimentos

Obrigado por contribuir! Cada contribuição, por menor que seja, faz diferença.

---

**Dúvidas?** Abra uma [Discussion](https://github.com/johnpitter/ollama-code/discussions)
