package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

const asciiLogo = `
██      ███████ ██████  ████████ ██ ██   ██ ██   ██ 
██      ██      ██   ██    ██    ██  ██ ██   ██ ██  
██      █████   ██████     ██    ██   ███     ███  
██      ██      ██         ██    ██  ██ ██   ██ ██ 
███████ ███████ ██         ██    ██ ██   ██ ██   ██ `

// Функция для печати градиентного логотипа
func printGradientLogo() {
	p := termenv.ColorProfile()
	lines := strings.Split(asciiLogo, "\n")

	// Определяем цвета: от ярко-синего к глубокому темно-синему
	startColor, _ := colorful.Hex("#00BFFF") // DeepSkyBlue
	endColor, _ := colorful.Hex("#00008B")   // DarkBlue

	for i, line := range lines {
		// Вычисляем шаг градиента для каждой строки
		ratio := float64(i) / float64(len(lines))
		resColor := startColor.BlendLuv(endColor, ratio).Hex()

		// Печатаем строку
		fmt.Println(termenv.String(line).Foreground(p.Color(resColor)))
	}
	fmt.Println(termenv.String("   CLI App for detecting AI hallucinations.").Italic().Foreground(p.Color("#808080")))
	fmt.Println()
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "CLI Application for AI Hallucinations",
	// Мы убрали Long, так как выведем логотип сами в Run или PersistentPreRun
	Run: func(cmd *cobra.Command, args []string) {
		printGradientLogo()
		cmd.Help() // Показывает подсказку, если запущен просто app
	},
}

// ... твои остальные команды (verify, batch и т.д.) остаются такими же ...

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a single response",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Placeholder for verify command")
	},
}

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Process a batch of inputs",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📦 Placeholder for batch command")
	},
}

var evaluateCmd = &cobra.Command{
	Use:   "evaluate",
	Short: "Evaluate on a test dataset",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📊 Placeholder for evaluate command")
	},
}

var buildIndexCmd = &cobra.Command{
	Use:   "build-index",
	Short: "Build an index",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔨 Placeholder for build-index command")
	},
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check Python API status",
	Run: func(cmd *cobra.Command, args []string) {
		client := NewPythonClient("http://localhost:8000")

		fmt.Println("🔍 Проверка Python API...")

		if err := client.HealthCheck(); err != nil {
			fmt.Printf("❌ Python API недоступен: %v\n", err)
			fmt.Println("\n💡 Запустите Python сервер:")
			fmt.Println("   cd Python && python app.py")
			os.Exit(1)
		}

		fmt.Println("✅ Python API работает!")
	},
}

func cli() {
	// Добавляем команды
	rootCmd.AddCommand(verifyCmd, batchCmd, evaluateCmd, buildIndexCmd, healthCmd) // <- добавили healthCmd

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
