# Download Direto de Modelos Ollama - PowerShell Script
# Para ambientes corporativos com proxy

param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("1","2","3","4")]
    [string]$Method = "4"
)

# Configurações
$OllamaModelsDir = "$env:USERPROFILE\.ollama\models"
$TempDir = "$env:TEMP\ollama-models"

# Criar diretórios
New-Item -ItemType Directory -Force -Path $OllamaModelsDir | Out-Null
New-Item -ItemType Directory -Force -Path $TempDir | Out-Null

# Cores
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Header {
    Write-ColorOutput "`n════════════════════════════════════════════════════" "Cyan"
    Write-ColorOutput "  Download Direto de Modelos Ollama (Windows)" "Cyan"
    Write-ColorOutput "════════════════════════════════════════════════════`n" "Cyan"
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "✓ $Message" "Green"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "✗ $Message" "Red"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "→ $Message" "Yellow"
}

function Write-Step {
    param([string]$Message)
    Write-ColorOutput "`n▶ $Message" "Blue"
}

# Função de download com retry
function Download-FileWithRetry {
    param(
        [string]$Url,
        [string]$OutputPath,
        [int]$MaxRetries = 3
    )

    $retries = 0
    $success = $false

    # Configurar para ignorar certificados SSL (para proxy corporativo)
    [System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12

    while (-not $success -and $retries -lt $MaxRetries) {
        try {
            Write-Info "Tentativa $($retries + 1) de $MaxRetries..."

            # Usar WebClient com configurações de proxy
            $webClient = New-Object System.Net.WebClient

            # Usar proxy do sistema
            $webClient.Proxy = [System.Net.WebRequest]::GetSystemWebProxy()
            $webClient.Proxy.Credentials = [System.Net.CredentialCache]::DefaultNetworkCredentials

            # Progress bar
            Register-ObjectEvent -InputObject $webClient -EventName DownloadProgressChanged -SourceIdentifier WebClient.DownloadProgressChanged -Action {
                Write-Progress -Activity "Downloading" -Status "$($EventArgs.ProgressPercentage)% Complete" -PercentComplete $EventArgs.ProgressPercentage
            } | Out-Null

            $webClient.DownloadFile($Url, $OutputPath)
            Unregister-Event -SourceIdentifier WebClient.DownloadProgressChanged -ErrorAction SilentlyContinue
            Write-Progress -Activity "Downloading" -Completed

            $success = $true
            Write-Success "Download concluído"
            return $true

        } catch {
            $retries++
            Write-Error "Falha no download: $_"

            if ($retries -lt $MaxRetries) {
                Write-Info "Aguardando 5 segundos antes de tentar novamente..."
                Start-Sleep -Seconds 5
            }
        }
    }

    return $false
}

# Método 1: Download via aria2c (recomendado para proxy corporativo)
function Download-WithAria2 {
    Write-Step "Download com aria2c (Bypass Proxy)"

    # Verificar se aria2c está instalado
    $aria2Path = (Get-Command aria2c -ErrorAction SilentlyContinue)

    if (-not $aria2Path) {
        Write-Error "aria2c não está instalado!"
        Write-Info "Instalando via Chocolatey..."

        # Tentar instalar via Chocolatey
        if (Get-Command choco -ErrorAction SilentlyContinue) {
            choco install aria2 -y
        } else {
            Write-Error "Chocolatey não instalado. Instale manualmente:"
            Write-Info "1. Baixe: https://github.com/aria2/aria2/releases"
            Write-Info "2. Ou instale via Scoop: scoop install aria2"
            return
        }
    }

    # URLs dos modelos
    $models = @{
        "qwen2.5-coder-32b" = "https://huggingface.co/Qwen/Qwen2.5-Coder-32B-Instruct-GGUF/resolve/main/qwen2.5-coder-32b-instruct-q6_k.gguf"
        "deepseek-coder-v2-16b" = "https://huggingface.co/bartowski/DeepSeek-Coder-V2-Lite-Instruct-GGUF/resolve/main/DeepSeek-Coder-V2-Lite-Instruct-Q8_0.gguf"
    }

    foreach ($model in $models.GetEnumerator()) {
        Write-Step "Baixando $($model.Key)..."
        $outputFile = Join-Path $TempDir "$($model.Key).gguf"

        # aria2c com configurações otimizadas
        $aria2Args = @(
            "--max-connection-per-server=16",
            "--split=16",
            "--min-split-size=1M",
            "--check-certificate=false",
            "--allow-overwrite=true",
            "--auto-file-renaming=false",
            "--continue=true",
            "--out=$outputFile",
            $model.Value
        )

        & aria2c $aria2Args

        if ($LASTEXITCODE -eq 0) {
            Write-Success "Download concluído: $($model.Key)"

            # Importar para Ollama
            Import-ModelToOllama -ModelFile $outputFile -ModelName $model.Key
        } else {
            Write-Error "Falha no download de $($model.Key)"
        }
    }
}

# Método 2: Download via PowerShell (WebClient)
function Download-WithPowerShell {
    Write-Step "Download via PowerShell WebClient"

    $url = "https://huggingface.co/Qwen/Qwen2.5-Coder-32B-Instruct-GGUF/resolve/main/qwen2.5-coder-32b-instruct-q6_k.gguf"
    $output = Join-Path $TempDir "qwen2.5-coder-32b.gguf"

    Write-Info "Baixando QWen2.5-Coder 32B (19GB)..."
    Write-Info "Isso pode levar alguns minutos..."

    if (Download-FileWithRetry -Url $url -OutputPath $output) {
        Write-Success "Download concluído!"
        Import-ModelToOllama -ModelFile $output -ModelName "qwen2.5-coder-32b"
    } else {
        Write-Error "Falha no download. Tente outro método."
    }
}

# Método 3: Download via curl (se disponível)
function Download-WithCurl {
    Write-Step "Download via curl"

    if (-not (Get-Command curl -ErrorAction SilentlyContinue)) {
        Write-Error "curl não está disponível!"
        return
    }

    $url = "https://huggingface.co/Qwen/Qwen2.5-Coder-32B-Instruct-GGUF/resolve/main/qwen2.5-coder-32b-instruct-q6_k.gguf"
    $output = Join-Path $TempDir "qwen2.5-coder-32b.gguf"

    Write-Info "Baixando com curl..."

    curl.exe -L -k -o $output $url

    if ($LASTEXITCODE -eq 0) {
        Write-Success "Download concluído!"
        Import-ModelToOllama -ModelFile $output -ModelName "qwen2.5-coder-32b"
    } else {
        Write-Error "Falha no download"
    }
}

# Importar modelo para Ollama
function Import-ModelToOllama {
    param(
        [string]$ModelFile,
        [string]$ModelName
    )

    Write-Step "Importando modelo para Ollama..."

    # Criar Modelfile
    $modelfileContent = @"
FROM $ModelFile
TEMPLATE `"`"`"{{ if .System }}<|im_start|>system
{{ .System }}<|im_end|>
{{ end }}{{ if .Prompt }}<|im_start|>user
{{ .Prompt }}<|im_end|>
<|im_start|>assistant
{{ end }}`"`"`"
PARAMETER stop "<|im_start|>"
PARAMETER stop "<|im_end|>"
PARAMETER temperature 0.7
PARAMETER num_gpu 999
"@

    $modelfilePath = Join-Path $TempDir "Modelfile"
    $modelfileContent | Out-File -FilePath $modelfilePath -Encoding utf8

    # Importar para Ollama
    Write-Info "Executando: ollama create $ModelName -f $modelfilePath"

    try {
        & ollama create "$ModelName" -f $modelfilePath

        if ($LASTEXITCODE -eq 0) {
            Write-Success "Modelo importado com sucesso!"
            Write-Info "Para testar: ollama run $ModelName"
        } else {
            Write-Error "Falha ao importar modelo"
        }
    } catch {
        Write-Error "Erro ao executar ollama: $_"
    }

    # Cleanup
    Remove-Item -Path $modelfilePath -Force -ErrorAction SilentlyContinue
}

# Mostrar links diretos
function Show-DirectLinks {
    Write-Step "Links Diretos para Download Manual"

    $links = @"

╔═══════════════════════════════════════════════════════════════════╗
║  LINKS DIRETOS PARA DOWNLOAD MANUAL (Windows)                     ║
╚═══════════════════════════════════════════════════════════════════╝

┌────────────────────────────────────────────────────────────────────┐
│ QWen2.5-Coder 32B (RECOMENDADO) - 19GB                             │
├────────────────────────────────────────────────────────────────────┤
│ https://huggingface.co/Qwen/Qwen2.5-Coder-32B-Instruct-GGUF/resolve/main/qwen2.5-coder-32b-instruct-q6_k.gguf
└────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│ DeepSeek-Coder-V2 16B - 9GB                                        │
├────────────────────────────────────────────────────────────────────┤
│ https://huggingface.co/bartowski/DeepSeek-Coder-V2-Lite-Instruct-GGUF/resolve/main/DeepSeek-Coder-V2-Lite-Instruct-Q8_0.gguf
└────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│ Nomic Embed Text (Embeddings) - 274MB                             │
├────────────────────────────────────────────────────────────────────┤
│ https://huggingface.co/nomic-ai/nomic-embed-text-v1.5/resolve/main/model.safetensors
└────────────────────────────────────────────────────────────────────┘

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

INSTRUÇÕES PARA DOWNLOAD MANUAL (Windows):

1. Abra PowerShell como Administrador

2. Baixe o modelo:

   `$url = "https://huggingface.co/Qwen/Qwen2.5-Coder-32B-Instruct-GGUF/resolve/main/qwen2.5-coder-32b-instruct-q6_k.gguf"
   `$output = "`$env:TEMP\qwen2.5-coder-32b.gguf"

   # Ignorar SSL (proxy corporativo)
   [System.Net.ServicePointManager]::ServerCertificateValidationCallback = {`$true}

   # Download
   Invoke-WebRequest -Uri `$url -OutFile `$output -UseBasicParsing

3. Crie um Modelfile:

   @"
   FROM `$env:TEMP\qwen2.5-coder-32b.gguf
   TEMPLATE `"`"`"{{ if .System }}<|im_start|>system
   {{ .System }}<|im_end|>
   {{ end }}{{ if .Prompt }}<|im_start|>user
   {{ .Prompt }}<|im_end|>
   <|im_start|>assistant
   {{ end }}`"`"`"
   PARAMETER temperature 0.7
   PARAMETER num_gpu 999
   "@ | Out-File -FilePath "`$env:TEMP\Modelfile" -Encoding utf8

4. Importar para Ollama:

   ollama create qwen2.5-coder:32b -f `$env:TEMP\Modelfile

5. Testar:

   ollama run qwen2.5-coder:32b "Write a function to reverse a string in Go"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ALTERNATIVA: Download via Navegador

1. Copie o link acima e cole no navegador
2. Aguarde o download (19GB)
3. Siga os passos 3-5 acima

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

"@

    Write-Output $links
}

# Menu principal
function Show-Menu {
    Write-Header

    Write-Output "Escolha o método de download:"
    Write-Output ""
    Write-Output "  1) Download via PowerShell WebClient (proxy do sistema)"
    Write-Output "  2) Download via curl (se disponível)"
    Write-Output "  3) Mostrar links para download manual"
    Write-Output "  4) Download via aria2c (RECOMENDADO para proxy corporativo)"
    Write-Output ""

    if (-not $Method) {
        $Method = Read-Host "Opção [1-4]"
    }

    switch ($Method) {
        "1" { Download-WithPowerShell }
        "2" { Download-WithCurl }
        "3" { Show-DirectLinks }
        "4" { Download-WithAria2 }
        default {
            Write-Error "Opção inválida"
            exit 1
        }
    }

    Write-Success "`nProcesso concluído!"
    Write-Output ""
    Write-Output "Para verificar modelos instalados:"
    Write-Output "  ollama list"
    Write-Output ""
    Write-Output "Para testar:"
    Write-Output "  ollama run qwen2.5-coder:32b"
}

# Executar
Show-Menu
