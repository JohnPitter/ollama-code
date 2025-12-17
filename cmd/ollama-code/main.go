package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/johnpitter/ollama-code/internal/agent"
	"github.com/johnpitter/ollama-code/internal/modes"
	"github.com/spf13/cobra"
)

var (
	flagMode    string
	flagModel   string
	flagURL     string
	flagWorkDir string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ollama-code",
		Short: "AI code assistant using Ollama",
		Long:  "Ollama Code - Assistente de código AI rodando 100% local",
	}

	// Chat command
	chatCmd := &cobra.Command{
		Use:   "chat [message]",
		Short: "Start interactive chat",
		Long:  "Inicia chat interativo com o assistente",
		Run:   runChat,
	}

	chatCmd.Flags().StringVarP(&flagMode, "mode", "m", "interactive", "Operation mode: readonly, interactive, autonomous")
	chatCmd.Flags().StringVar(&flagModel, "model", "qwen2.5-coder:32b-instruct-q6_K", "Ollama model to use")
	chatCmd.Flags().StringVar(&flagURL, "url", "http://localhost:11434", "Ollama server URL")
	chatCmd.Flags().StringVarP(&flagWorkDir, "workdir", "w", ".", "Working directory")

	// Ask command (one-shot)
	askCmd := &cobra.Command{
		Use:   "ask <question>",
		Short: "Ask a single question",
		Long:  "Faz uma pergunta única e retorna resposta",
		Args:  cobra.ExactArgs(1),
		Run:   runAsk,
	}

	askCmd.Flags().StringVar(&flagModel, "model", "qwen2.5-coder:32b-instruct-q6_K", "Ollama model to use")
	askCmd.Flags().StringVar(&flagURL, "url", "http://localhost:11434", "Ollama server URL")

	rootCmd.AddCommand(chatCmd, askCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runChat(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// Criar agente
	cfg := agent.Config{
		OllamaURL: flagURL,
		Model:     flagModel,
		Mode:      modes.ParseMode(flagMode),
		WorkDir:   flagWorkDir,
	}

	ag, err := agent.NewAgent(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating agent: %v\n", err)
		os.Exit(1)
	}

	// Banner
	blue := color.New(color.FgBlue, color.Bold)
	yellow := color.New(color.FgYellow)

	blue.Println("\n🤖 Ollama Code - AI Code Assistant")
	fmt.Printf("Modelo: %s\n", flagModel)
	fmt.Printf("Modo: %s (%s)\n", ag.GetMode(), ag.GetMode().Description())
	fmt.Printf("Diretório: %s\n", ag.GetWorkDir())
	yellow.Println("\nDigite 'exit' para sair, 'help' para ajuda\n")

	// Se tem mensagem inicial
	if len(args) > 0 {
		initialMessage := strings.Join(args, " ")
		if err := ag.ProcessMessage(ctx, initialMessage); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	// Loop interativo
	reader := bufio.NewReader(os.Stdin)
	green := color.New(color.FgGreen, color.Bold)

	for {
		green.Print("\n💬 Você: ")

		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			break
		}

		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}

		// Comandos especiais
		switch strings.ToLower(message) {
		case "exit", "quit":
			blue.Println("\n👋 Até logo!")
			return

		case "help":
			showHelp()
			continue

		case "clear":
			ag.ClearHistory()
			fmt.Println("\n✓ Histórico limpo")
			continue

		case "mode":
			fmt.Printf("\nModo atual: %s (%s)\n", ag.GetMode(), ag.GetMode().Description())
			continue

		case "pwd":
			fmt.Printf("\nDiretório: %s\n", ag.GetWorkDir())
			continue
		}

		// Processar mensagem
		if err := ag.ProcessMessage(ctx, message); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
}

func runAsk(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	cfg := agent.Config{
		OllamaURL: flagURL,
		Model:     flagModel,
		Mode:      modes.ModeReadOnly, // Ask é sempre readonly
		WorkDir:   ".",
	}

	ag, err := agent.NewAgent(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating agent: %v\n", err)
		os.Exit(1)
	}

	question := args[0]

	if err := ag.ProcessMessage(ctx, question); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showHelp() {
	blue := color.New(color.FgBlue, color.Bold)
	blue.Println("\n📚 Comandos Disponíveis:\n")

	fmt.Println("  exit/quit     - Sair do chat")
	fmt.Println("  help          - Mostrar esta ajuda")
	fmt.Println("  clear         - Limpar histórico")
	fmt.Println("  mode          - Mostrar modo atual")
	fmt.Println("  pwd           - Mostrar diretório atual")
	fmt.Println("\n💡 Exemplos de uso:\n")
	fmt.Println("  - Leia o arquivo main.go")
	fmt.Println("  - Mostre a estrutura do projeto")
	fmt.Println("  - Execute os testes")
	fmt.Println("  - Busque por 'handleRequest' no código")
	fmt.Println("  - Pesquise na internet como fazer X em Go")
	fmt.Println()
}
