#!/bin/bash

###############################################################################
# Script de Download Direto de Modelos Ollama (Bypass Proxy Corporativo)
#
# Este script baixa modelos do Ollama diretamente do registry,
# sem usar 'ollama pull', contornando problemas de proxy corporativo.
###############################################################################

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configurações
OLLAMA_MODELS_DIR="${HOME}/.ollama/models"
REGISTRY_URL="https://registry.ollama.ai/v2/library"

# Criar diretórios necessários
mkdir -p "${OLLAMA_MODELS_DIR}/manifests/registry.ollama.ai/library"
mkdir -p "${OLLAMA_MODELS_DIR}/blobs"

###############################################################################
# FUNÇÕES AUXILIARES
###############################################################################

print_header() {
    echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Download Direto de Modelos Ollama${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}→${NC} $1"
}

print_step() {
    echo -e "\n${BLUE}▶${NC} $1"
}

# Função para verificar se comando existe
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Função para baixar arquivo com retry
download_file() {
    local url="$1"
    local output="$2"
    local max_retries=3
    local retry=0

    while [ $retry -lt $max_retries ]; do
        if command_exists wget; then
            if wget --no-check-certificate -O "$output" "$url" 2>/dev/null; then
                return 0
            fi
        elif command_exists curl; then
            if curl -k -L -o "$output" "$url" 2>/dev/null; then
                return 0
            fi
        fi

        retry=$((retry + 1))
        if [ $retry -lt $max_retries ]; then
            print_info "Tentativa $retry de $max_retries falhou. Tentando novamente..."
            sleep 2
        fi
    done

    return 1
}

###############################################################################
# FUNÇÃO PRINCIPAL DE DOWNLOAD
###############################################################################

download_model() {
    local model_name="$1"
    local model_tag="$2"

    print_step "Baixando modelo: ${model_name}:${model_tag}"

    # 1. Baixar manifesto
    print_info "Baixando manifesto..."
    local manifest_url="${REGISTRY_URL}/${model_name}/manifests/${model_tag}"
    local manifest_file="${OLLAMA_MODELS_DIR}/manifests/registry.ollama.ai/library/${model_name}/${model_tag}"

    mkdir -p "$(dirname "$manifest_file")"

    if ! download_file "$manifest_url" "$manifest_file"; then
        print_error "Falha ao baixar manifesto do modelo ${model_name}:${model_tag}"
        return 1
    fi

    print_success "Manifesto baixado"

    # 2. Extrair blobs do manifesto
    print_info "Processando camadas do modelo..."

    # Ler o manifesto e extrair digests dos blobs
    local blobs=$(grep -oP 'sha256:[a-f0-9]{64}' "$manifest_file" | sort -u)
    local blob_count=$(echo "$blobs" | wc -l)
    local current=0

    print_info "Total de camadas a baixar: $blob_count"

    # 3. Baixar cada blob
    for digest in $blobs; do
        current=$((current + 1))
        local short_digest=$(echo "$digest" | cut -c8-15)

        # Verificar se blob já existe
        local blob_file="${OLLAMA_MODELS_DIR}/blobs/${digest}"
        if [ -f "$blob_file" ]; then
            print_info "[$current/$blob_count] Camada $short_digest já existe (skip)"
            continue
        fi

        print_info "[$current/$blob_count] Baixando camada $short_digest..."

        local blob_url="${REGISTRY_URL}/${model_name}/blobs/${digest}"

        if download_file "$blob_url" "$blob_file"; then
            local size=$(du -h "$blob_file" | cut -f1)
            print_success "[$current/$blob_count] Camada $short_digest baixada ($size)"
        else
            print_error "Falha ao baixar camada $short_digest"
            return 1
        fi
    done

    print_success "Modelo ${model_name}:${model_tag} baixado com sucesso!"
    return 0
}

###############################################################################
# LINKS DIRETOS ALTERNATIVOS (Hugging Face Mirror)
###############################################################################

download_from_huggingface() {
    local model_name="$1"
    local hf_repo="$2"

    print_step "Baixando de Hugging Face: ${model_name}"

    local temp_dir="/tmp/ollama-model-${model_name}"
    mkdir -p "$temp_dir"

    print_info "Repositório: ${hf_repo}"

    # URLs do Hugging Face
    local base_url="https://huggingface.co/${hf_repo}/resolve/main"

    # Tentar baixar via git clone ou wget
    if command_exists git-lfs; then
        print_info "Usando git-lfs para download..."
        cd "$temp_dir"
        if git clone "https://huggingface.co/${hf_repo}" .; then
            print_success "Modelo baixado via git"
            # Importar para Ollama
            if command_exists ollama; then
                print_info "Importando modelo para Ollama..."
                # Nota: requer Modelfile customizado
                print_info "Você precisará criar um Modelfile para este modelo"
            fi
        fi
    else
        print_error "git-lfs não instalado. Instale com: apt install git-lfs"
    fi
}

###############################################################################
# LINKS DIRETOS PARA MODELOS (Backup)
###############################################################################

download_from_direct_links() {
    print_step "Downloads Diretos (Links Alternativos)"

    cat << 'EOF'

╔════════════════════════════════════════════════════════════════════╗
║  LINKS DIRETOS PARA DOWNLOAD MANUAL                                ║
╚════════════════════════════════════════════════════════════════════╝

Caso o script automático falhe, você pode baixar manualmente:

┌─────────────────────────────────────────────────────────────────────┐
│ QWen2.5-Coder 32B (Recomendado)                                     │
├─────────────────────────────────────────────────────────────────────┤
│ Hugging Face: https://huggingface.co/Qwen/Qwen2.5-Coder-32B-Instruct│
│ Tamanho: ~19GB (Q6_K quantizado)                                    │
│                                                                      │
│ Download direto (GGUF):                                             │
│ https://huggingface.co/Qwen/Qwen2.5-Coder-32B-Instruct-GGUF/resolve/main/qwen2.5-coder-32b-instruct-q6_k.gguf
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│ DeepSeek-Coder-V2 16B                                               │
├─────────────────────────────────────────────────────────────────────┤
│ Hugging Face: https://huggingface.co/deepseek-ai/DeepSeek-Coder-V2-Lite-Instruct
│ Tamanho: ~9GB (Q8_0 quantizado)                                     │
│                                                                      │
│ Download direto (GGUF):                                             │
│ https://huggingface.co/bartowski/DeepSeek-Coder-V2-Lite-Instruct-GGUF/resolve/main/DeepSeek-Coder-V2-Lite-Instruct-Q8_0.gguf
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│ Nomic Embed Text (Embeddings)                                       │
├─────────────────────────────────────────────────────────────────────┤
│ Hugging Face: https://huggingface.co/nomic-ai/nomic-embed-text-v1.5 │
│ Tamanho: ~274MB                                                     │
│                                                                      │
│ Download direto:                                                    │
│ https://huggingface.co/nomic-ai/nomic-embed-text-v1.5/resolve/main/model.safetensors
└─────────────────────────────────────────────────────────────────────┘

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

INSTRUÇÕES PARA USO MANUAL:

1. Baixe o arquivo .gguf usando wget ou navegador:

   wget --no-check-certificate -O qwen2.5-coder-32b.gguf \
     "https://huggingface.co/Qwen/Qwen2.5-Coder-32B-Instruct-GGUF/resolve/main/qwen2.5-coder-32b-instruct-q6_k.gguf"

2. Crie um Modelfile:

   cat > Modelfile << 'MODELFILE'
   FROM ./qwen2.5-coder-32b.gguf
   TEMPLATE """{{ if .System }}<|im_start|>system
   {{ .System }}<|im_end|>
   {{ end }}{{ if .Prompt }}<|im_start|>user
   {{ .Prompt }}<|im_end|>
   <|im_start|>assistant
   {{ end }}"""
   PARAMETER stop "<|im_start|>"
   PARAMETER stop "<|im_end|>"
   PARAMETER temperature 0.7
   PARAMETER num_gpu 999
   MODELFILE

3. Importar para Ollama:

   ollama create qwen2.5-coder:32b-instruct-q6_K -f Modelfile

4. Testar:

   ollama run qwen2.5-coder:32b-instruct-q6_K "Hello, write a function to reverse a string in Go"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

EOF
}

###############################################################################
# DOWNLOAD COM PROXY BYPASS (aria2c)
###############################################################################

download_with_aria2() {
    local url="$1"
    local output="$2"

    if ! command_exists aria2c; then
        print_error "aria2c não instalado. Instalando..."
        if command_exists apt-get; then
            sudo apt-get update && sudo apt-get install -y aria2
        elif command_exists yum; then
            sudo yum install -y aria2
        else
            print_error "Não foi possível instalar aria2 automaticamente"
            return 1
        fi
    fi

    print_info "Baixando com aria2c (suporta proxy corporativo)..."

    # aria2c com configurações otimizadas para proxy corporativo
    aria2c \
        --max-connection-per-server=16 \
        --split=16 \
        --min-split-size=1M \
        --check-certificate=false \
        --allow-overwrite=true \
        --auto-file-renaming=false \
        --continue=true \
        --out="$output" \
        "$url"
}

###############################################################################
# SCRIPT PRINCIPAL
###############################################################################

main() {
    print_header

    # Verificar se Ollama está instalado
    if ! command_exists ollama; then
        print_error "Ollama não está instalado!"
        echo ""
        echo "Instale o Ollama primeiro:"
        echo "  curl -fsSL https://ollama.ai/install.sh | sh"
        echo ""
        echo "Ou baixe manualmente de: https://ollama.ai/download"
        exit 1
    fi

    echo ""
    echo "Escolha o método de download:"
    echo ""
    echo "  1) Download via Ollama Registry (requer rede sem proxy)"
    echo "  2) Download direto via Hugging Face (bypass proxy)"
    echo "  3) Mostrar links para download manual"
    echo "  4) Download com aria2c (melhor para proxy corporativo)"
    echo ""
    read -p "Opção [1-4]: " choice

    case $choice in
        1)
            print_step "Baixando via Ollama Registry"
            download_model "qwen2.5-coder" "32b-instruct-q6_K"
            download_model "deepseek-coder-v2" "16b-lite-instruct-q8_0"
            download_model "nomic-embed-text" "latest"
            ;;
        2)
            print_step "Baixando via Hugging Face"
            download_from_huggingface "qwen2.5-coder-32b" "Qwen/Qwen2.5-Coder-32B-Instruct-GGUF"
            ;;
        3)
            download_from_direct_links
            ;;
        4)
            print_step "Download com aria2c"
            echo ""
            echo "Baixando QWen2.5-Coder 32B..."
            download_with_aria2 \
                "https://huggingface.co/Qwen/Qwen2.5-Coder-32B-Instruct-GGUF/resolve/main/qwen2.5-coder-32b-instruct-q6_k.gguf" \
                "qwen2.5-coder-32b.gguf"

            if [ $? -eq 0 ]; then
                print_success "Download concluído!"
                print_info "Importando para Ollama..."

                # Criar Modelfile
                cat > Modelfile << 'MODELFILE'
FROM ./qwen2.5-coder-32b.gguf
TEMPLATE """{{ if .System }}<|im_start|>system
{{ .System }}<|im_end|>
{{ end }}{{ if .Prompt }}<|im_start|>user
{{ .Prompt }}<|im_end|>
<|im_start|>assistant
{{ end }}"""
PARAMETER stop "<|im_start|>"
PARAMETER stop "<|im_end|>"
PARAMETER temperature 0.7
PARAMETER num_gpu 999
MODELFILE

                ollama create qwen2.5-coder:32b-instruct-q6_K -f Modelfile

                print_success "Modelo importado com sucesso!"

                # Cleanup
                rm -f Modelfile qwen2.5-coder-32b.gguf
            fi
            ;;
        *)
            print_error "Opção inválida"
            exit 1
            ;;
    esac

    echo ""
    print_success "Processo concluído!"
    echo ""
    echo "Para verificar os modelos instalados:"
    echo "  ollama list"
    echo ""
    echo "Para testar um modelo:"
    echo "  ollama run qwen2.5-coder:32b-instruct-q6_K"
}

# Executar script principal
main "$@"
