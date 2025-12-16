#!/bin/bash

# Setup Otimizado Ollama para RTX Ada 2000 (16GB VRAM)
# Hardware: i9 14th gen, 64GB RAM, RTX Ada 2000

echo "🚀 Configurando Ollama para Performance Máxima"
echo ""

# Variáveis de ambiente para Ollama
export OLLAMA_NUM_PARALLEL=4          # Processar 4 requests simultâneos
export OLLAMA_MAX_LOADED_MODELS=2     # Manter 2 modelos em VRAM
export OLLAMA_FLASH_ATTENTION=1       # Ativar Flash Attention (mais rápido)
export OLLAMA_GPU_LAYERS=999          # Forçar TODAS camadas na GPU
export OLLAMA_HOST=0.0.0.0:11434      # Aceitar conexões de qualquer IP

# CUDA optimizations
export CUDA_VISIBLE_DEVICES=0
export CUDA_LAUNCH_BLOCKING=0

echo "✅ Variáveis de ambiente configuradas"
echo ""

# Criar arquivo de configuração persistente
mkdir -p ~/.config/ollama

cat > ~/.config/ollama/env.conf << 'EOF'
# Ollama Performance Configuration
OLLAMA_NUM_PARALLEL=4
OLLAMA_MAX_LOADED_MODELS=2
OLLAMA_FLASH_ATTENTION=1
OLLAMA_GPU_LAYERS=999
OLLAMA_ORIGINS=*
EOF

echo "✅ Arquivo de configuração criado em ~/.config/ollama/env.conf"
echo ""

# Criar systemd service (se não existir)
if [ -d "/etc/systemd/system" ]; then
    echo "📝 Criando systemd service..."
    sudo tee /etc/systemd/system/ollama.service > /dev/null << 'EOSERVICE'
[Unit]
Description=Ollama Service
After=network-online.target

[Service]
Type=simple
User=$USER
EnvironmentFile=/home/$USER/.config/ollama/env.conf
ExecStart=/usr/local/bin/ollama serve
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
EOSERVICE

    sudo systemctl daemon-reload
    sudo systemctl enable ollama
    sudo systemctl restart ollama
    
    echo "✅ Systemd service configurado e iniciado"
else
    echo "⚠️  Systemd não disponível, inicie manualmente com:"
    echo "   source ~/.config/ollama/env.conf && ollama serve"
fi

echo ""
echo "🎯 Modelos Recomendados para seu Hardware:"
echo ""
echo "   Para CÓDIGO (principal uso):"
echo "   ├─ deepseek-coder-v2:16b-lite-instruct-q8_0  # Rápido, 16GB total"
echo "   ├─ qwen2.5-coder:32b-instruct-q6_K           # Melhor qualidade"
echo "   └─ codestral:22b-v0.1-q8_0                   # Balanceado"
echo ""
echo "   Para CHAT/Explicações:"
echo "   ├─ qwen2.5:32b-instruct-q6_K                 # Excelente reasoning"
echo "   └─ llama3.1:70b-instruct-q4_K_M              # Máxima capacidade"
echo ""
echo "   Para EMBEDDINGS (contexto):"
echo "   └─ nomic-embed-text                          # Leve e eficiente"
echo ""

echo "💡 Comandos para instalar:"
echo ""
echo "# Modelo principal (CODING) - Recomendado"
echo "ollama pull qwen2.5-coder:32b-instruct-q6_K"
echo ""
echo "# Modelo secundário (CHAT)"
echo "ollama pull qwen2.5:32b-instruct-q6_K"
echo ""
echo "# Embeddings"
echo "ollama pull nomic-embed-text"
echo ""

echo "🔧 Verificar se GPU está sendo usada:"
echo "   nvidia-smi -l 1"
echo ""
echo "📊 Testar performance:"
echo "   ollama run qwen2.5-coder:32b-instruct-q6_K 'Write a fibonacci function in Python'"
echo ""

echo "✅ Setup completo!"
