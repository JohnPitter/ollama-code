# 🚀 CI/CD Configuration Guide

## Data de Implementação: 22/12/2024

---

## 📋 Resumo

Este documento descreve a configuração completa de CI/CD implementada para o projeto Ollama Code usando GitHub Actions, GoReleaser e outras ferramentas de automação.

---

## 1. 🔄 GitHub Actions Workflow

### Arquivo: `.github/workflows/ci.yml`

O workflow principal executa em cada push e pull request nas branches `main` e `develop`, e em tags de versão.

### Jobs Configurados

#### 1.1 **Test Job**
- **Plataformas**: Ubuntu, Windows, macOS
- **Versões Go**: 1.21, 1.22
- **Execução**: Testes com race detector e cobertura
- **Upload**: Coverage para Codecov (apenas Ubuntu + Go 1.22)

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, windows-latest, macos-latest]
    go: ['1.21', '1.22']
```

#### 1.2 **Lint Job**
- **Plataforma**: Ubuntu
- **Ferramenta**: golangci-lint
- **Timeout**: 5 minutos
- **Versão**: Latest

#### 1.3 **Build Job**
- **Dependências**: test, lint (só executa se ambos passarem)
- **Plataformas**: Ubuntu, Windows, macOS
- **Artefatos**: Binários compilados para cada plataforma

#### 1.4 **Release Job**
- **Trigger**: Tags começando com `v` (ex: v1.0.0)
- **Dependências**: build
- **Ferramenta**: GoReleaser
- **Saída**: Releases no GitHub com binários para todas as plataformas

---

## 2. 📦 GoReleaser Configuration

### Arquivo: `.goreleaser.yml`

Configuração para automatizar releases multi-plataforma.

### Features

#### 2.1 **Build Matrix**
```yaml
goos:
  - linux
  - windows
  - darwin
goarch:
  - amd64
  - arm64
```

#### 2.2 **Optimizações**
- CGO desabilitado para portabilidade máxima
- Flags de otimização: `-s -w` (reduz tamanho do binário)
- LDFlags para injetar version, commit, date

#### 2.3 **Archives**
- **Linux/macOS**: tar.gz
- **Windows**: zip
- Inclui README, LICENSE, e docs de features

#### 2.4 **Changelog Automatizado**
Agrupa commits por tipo:
- 🆕 New Features (feat:)
- 🐛 Bug Fixes (fix:)
- ⚡ Performance Improvements (perf:)
- ♻️ Refactors (refactor:)
- 📝 Other Changes

---

## 3. 🔍 golangci-lint Configuration

### Arquivo: `.golangci.yml`

Configuração de linting para garantir qualidade de código.

### Linters Habilitados

#### Core Linters
- `gofmt` - Formatação
- `goimports` - Imports organizados
- `govet` - Análise estática padrão Go
- `errcheck` - Verifica erros não tratados
- `staticcheck` - Análise estática avançada

#### Security Linters
- `gosec` - Vulnerabilidades de segurança

#### Code Quality Linters
- `gocyclo` - Complexidade ciclomática (max: 15)
- `dupl` - Código duplicado (threshold: 100)
- `goconst` - Strings constantes não extraídas
- `misspell` - Erros de ortografia
- `lll` - Linhas muito longas (max: 120)

#### Performance Linters
- `prealloc` - Slices que podem ser pré-alocados
- `unparam` - Parâmetros não utilizados

### Exclusões
- Testes são mais flexíveis (excluem: gocyclo, dupl, gosec, lll)
- Comandos CLI excluem lll

---

## 4. 🛠️ Makefile Enhancements

### Arquivo: `Makefile`

Novos targets adicionados para CI/CD:

#### CI Targets

**`make ci`**
```bash
Executes: deps -> lint -> test -> build
```
Pipeline básico de CI.

**`make ci-full`**
```bash
Executes: deps -> lint -> test-coverage -> build-all
```
Pipeline completo com coverage e builds multi-plataforma.

**`make ci-tools`**
```bash
Installs: golangci-lint, goreleaser, goimports
```
Instala todas as ferramentas de CI/CD.

**`make check`**
```bash
Executes: lint -> vet -> test
```
Validação completa de código.

**`make release-dry-run`**
```bash
Executes: goreleaser release --snapshot --skip-publish --clean
```
Testa processo de release sem publicar.

#### New Test Targets

**`make test-tools`**
```bash
Runs: go test -v ./internal/tools/...
```
Testa apenas os tools (143 testes).

**`make vet`**
```bash
Runs: go vet ./...
```
Análise estática padrão do Go.

---

## 5. 📊 Badges no README

### Arquivo: `README.md`

Badges adicionados:

```markdown
[![CI/CD](https://github.com/johnpitter/ollama-code/workflows/CI/CD/badge.svg)]
[![Tests](https://img.shields.io/badge/Tests-143_passing-success)]
[![Coverage](https://img.shields.io/badge/Coverage-Codecov-blue)]
[![Go Report Card](https://goreportcard.com/badge/github.com/johnpitter/ollama-code)]
```

---

## 6. 🔐 Secrets Necessários

### GitHub Secrets

Para o workflow funcionar completamente, configure:

1. **CODECOV_TOKEN** (opcional)
   - Para: Upload de coverage
   - Obtido em: https://codecov.io/

2. **GITHUB_TOKEN** (automático)
   - Fornecido automaticamente pelo GitHub Actions
   - Usado para: Criar releases

---

## 7. 📖 Como Usar

### Local Development

#### Rodar CI Localmente
```bash
# Pipeline completo
make ci

# Com coverage
make ci-full

# Apenas checks
make check
```

#### Testar Release Localmente
```bash
# Dry-run do GoReleaser
make release-dry-run

# Verifica artefatos em ./dist/
```

#### Instalar Ferramentas de CI
```bash
make ci-tools
```

### GitHub Actions

#### Trigger Manual
1. Vá para Actions tab no GitHub
2. Selecione workflow "CI/CD"
3. Click "Run workflow"

#### Criar Release
```bash
# Criar tag de versão
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GitHub Actions automaticamente:
# 1. Roda todos os testes
# 2. Faz lint do código
# 3. Builda para todas as plataformas
# 4. Cria release com binários
```

---

## 8. 🔄 Workflow de Development

### Branch Strategy

```
main (production)
  ↑
  Pull Request (CI runs)
  ↑
develop (staging)
  ↑
  Pull Request (CI runs)
  ↑
feature/* (development)
```

### CI em cada etapa

1. **Push para feature branch**
   - Nenhum CI (economiza recursos)

2. **Pull Request para develop**
   - ✅ Tests
   - ✅ Lint
   - ✅ Build (single platform)

3. **Pull Request para main**
   - ✅ Tests (todas plataformas)
   - ✅ Lint
   - ✅ Build (todas plataformas)

4. **Tag de release (v*)**
   - ✅ Tests
   - ✅ Lint
   - ✅ Build + Release multi-plataforma

---

## 9. 📈 Métricas e Monitoramento

### Coverage Reports

- **Codecov**: https://codecov.io/gh/johnpitter/ollama-code
- Upload automático em cada push para main
- Comentários automáticos em PRs com diff de coverage

### Go Report Card

- **URL**: https://goreportcard.com/report/github.com/johnpitter/ollama-code
- Atualização automática
- Avalia: gofmt, go vet, gocyclo, golint, ineffassign, license, misspell

### GitHub Actions Insights

- **URL**: https://github.com/johnpitter/ollama-code/actions
- Histórico de builds
- Tempos de execução
- Taxa de sucesso

---

## 10. 🐛 Troubleshooting

### CI Failing

#### Teste Falhando
```bash
# Rodar localmente
make test

# Com verbose
make test-verbose

# Apenas tools
make test-tools
```

#### Lint Falhando
```bash
# Rodar lint localmente
make lint

# Auto-fix formataçõa
make fmt
```

#### Build Falhando
```bash
# Testar build localmente
make build

# Todas as plataformas
make build-all
```

### GoReleaser Issues

#### Dry-run Falhando
```bash
# Debug
make release-dry-run

# Verificar .goreleaser.yml
goreleaser check
```

#### Falta de Tag
```bash
# GoReleaser precisa de uma tag
git tag -a v0.1.0 -m "Test release"

# Dry-run com tag
make release-dry-run
```

---

## 11. ✨ Próximas Melhorias

### Planejadas

1. **Testes de Integração**
   - [ ] Testes E2E automatizados
   - [ ] Testes de performance
   - [ ] Benchmark tracking

2. **Deploy Automático**
   - [ ] Docker images
   - [ ] Homebrew formula
   - [ ] Chocolatey package (Windows)
   - [ ] APT repository (Ubuntu/Debian)

3. **Documentação**
   - [ ] Auto-gerar docs do código
   - [ ] Changelog automático
   - [ ] API documentation

4. **Segurança**
   - [ ] Dependabot para updates automáticos
   - [ ] Security scanning (Snyk/Trivy)
   - [ ] SBOM generation

---

## 12. 📚 Recursos e Referências

### Documentação Oficial

- [GitHub Actions](https://docs.github.com/en/actions)
- [GoReleaser](https://goreleaser.com/intro/)
- [golangci-lint](https://golangci-lint.run/)
- [Codecov](https://docs.codecov.com/)

### Arquivos de Configuração

- `.github/workflows/ci.yml` - GitHub Actions workflow
- `.goreleaser.yml` - GoReleaser config
- `.golangci.yml` - Linter config
- `Makefile` - Build automation

---

## 13. 🎉 Status Atual

### ✅ Implementado

- ✅ GitHub Actions CI/CD
- ✅ Testes automatizados (143 tests)
- ✅ Linting automatizado
- ✅ Builds multi-plataforma
- ✅ Release automatizado com GoReleaser
- ✅ Coverage tracking
- ✅ Makefile enhancements
- ✅ README badges

### 📊 Estatísticas

- **Plataformas**: Linux, Windows, macOS
- **Arquiteturas**: amd64, arm64
- **Versões Go**: 1.21, 1.22
- **Total de Testes**: 143
- **Linters Ativos**: 21

---

## 🎯 Conclusão

O projeto agora possui um pipeline CI/CD completo e profissional que:

1. ✅ **Garante Qualidade**: Testes e linting em todas as mudanças
2. ✅ **Multi-Plataforma**: Builds automáticos para Linux, Windows, macOS
3. ✅ **Releases Automáticos**: Tags geram releases completos
4. ✅ **Developer Friendly**: Makefile para desenvolvimento local
5. ✅ **Monitoramento**: Badges e metrics tracking

**Data de Conclusão**: 22/12/2024
**Status**: ✅ Completo e Funcional
