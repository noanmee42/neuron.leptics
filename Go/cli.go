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

var (
	queryFlag    string
	responseFlag string
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a single response for AI hallucinations",
	Long: `The verify command checks for AI hallucinations in a given response
by extracting claims and saving them to a JSON file.
Requires GEMINI_API_KEY environment variable to be set.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Check for GEMINI_API_KEY
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			fmt.Println("❌ Ошибка: Переменная окружения GEMINI_API_KEY не установлена.")
			fmt.Println("💡 Пожалуйста, установите ее перед использованием команды verify.")
			os.Exit(1)
		}

		// 2. Initialize Python client
		client := NewPythonClient("http://localhost:8000")

		// 3. Health Check
		fmt.Println("🔍 Проверка Python API...")
		if err := client.HealthCheck(); err != nil {
			fmt.Printf("❌ Python API недоступен: %v\n", err)
			fmt.Println("\n💡 Запустите Python сервер:")
			fmt.Println("   cd Python && python app.py")
			os.Exit(1)
		}
		fmt.Println("✅ Python API работает!")

		// 4. Input validation
		if queryFlag == "" || responseFlag == "" {
			fmt.Println("❌ Ошибка: Необходимо указать и запрос (-q) и ответ (-r).")
			cmd.Help()
			os.Exit(1)
		}

		// 5. Extract and Save
		fmt.Println("🚀 Извлечение утверждений и сохранение...")
		result, err := client.ExtractAndSave(queryFlag, responseFlag)
		if err != nil {
			fmt.Printf("❌ Ошибка при извлечении и сохранении: %v\n", err)
			os.Exit(1)
		}

		// 6. Output Results
		fmt.Printf("✅ Утверждения успешно извлечены и сохранены в файл: %s\n", result.Filename)
		fmt.Printf("   Количество извлеченных утверждений: %d\n", result.ClaimsCount)

		// 7. Next Step Hint
		fmt.Println("\n💡 Следующий шаг: Проверьте извлеченные утверждения с помощью Fact Check команды.")
		fmt.Println("   Пример: go run . fact-check -f " + result.Filename)
	},
}

func init() {
	verifyCmd.Flags().StringVarP(&queryFlag, "query", "q", "", "The query string provided to the AI.")
	verifyCmd.Flags().StringVarP(&responseFlag, "response", "r", "", "The AI's response to be verified.")
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
