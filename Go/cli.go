package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

const asciiLogo = `
██      ███████ ██████  ████████ ██ ██   ██ ██   ██ 
██      ██      ██   ██    ██    ██  ██ ██   ██ ██  
██      █████   ██████     ██    ██   ███     ███  
██      ██      ██         ██    ██  ██ ██   ██ ██ 
███████ ███████ ██         ██    ██ ██   ██ ██   ██ `

func printGradientLogo() {
	p := termenv.ColorProfile()
	lines := strings.Split(asciiLogo, "\n")

	startColor, _ := colorful.Hex("#00BFFF")
	endColor, _ := colorful.Hex("#00008B")

	for i, line := range lines {
		ratio := float64(i) / float64(len(lines))
		resColor := startColor.BlendLuv(endColor, ratio).Hex()
		fmt.Println(termenv.String(line).Foreground(p.Color(resColor)))
	}
	fmt.Println(termenv.String("   CLI App for detecting AI hallucinations.").Italic().Foreground(p.Color("#808080")))
	fmt.Println()
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "CLI Application for AI Hallucinations",
	Run: func(cmd *cobra.Command, args []string) {
		printGradientLogo()
		cmd.Help()
	},
}

var (
	queryFlag    string
	responseFlag string
	fileFlag     string
	keyFlag      string
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a single response for AI hallucinations",
	Long: `The verify command checks for AI hallucinations in a given response
by extracting claims and saving them to a JSON file.
Requires GEMINI_API_KEY environment variable to be set.`,
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("GEMINI_API_KEY") == "" {
			fmt.Println("❌ Ошибка: Переменная окружения GEMINI_API_KEY не установлена.")
			fmt.Println("💡 Получите ключ: https://aistudio.google.com/app/apikey")
			os.Exit(1)
		}

		client := NewPythonClient("http://localhost:8000")

		fmt.Println("🔍 Проверка Python API...")
		if err := client.HealthCheck(); err != nil {
			fmt.Printf("❌ Python API недоступен: %v\n", err)
			fmt.Println("\n💡 Запустите Python сервер:")
			fmt.Println("   cd Python && python app.py")
			os.Exit(1)
		}
		fmt.Println("✅ Python API работает!")

		fmt.Println("🚀 Извлечение утверждений и сохранение...")
		result, err := client.ExtractAndSave(queryFlag, responseFlag)
		if err != nil {
			fmt.Printf("❌ Ошибка при извлечении и сохранении: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Утверждения сохранены в файл: %s\n", result.Filename)
		fmt.Printf("   Количество извлеченных утверждений: %d\n", result.ClaimsCount)
		fmt.Println("\n💡 Следующий шаг:")
		fmt.Println("   go run . check -f " + result.Filename)
	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check extracted claims via Jina AI Grounding API",
	Long: `Reads a JSON file with extracted claims and verifies each one
using the Jina AI Grounding API.
Requires JINA_API_KEY environment variable or -k flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := keyFlag
		if apiKey == "" {
			apiKey = os.Getenv("JINA_API_KEY")
		}
		if apiKey == "" {
			fmt.Println("❌ Jina AI API ключ не найден!")
			fmt.Println("\n💡 Как получить ключ:")
			fmt.Println("   1. Перейдите на https://jina.ai/")
			fmt.Println("   2. Нажмите 'Get API Key' — бесплатно 1M токенов")
			fmt.Println("\n   Затем: set JINA_API_KEY=ваш_ключ")
			os.Exit(1)
		}

		fmt.Printf("📂 Чтение файла: %s\n", fileFlag)
		data, err := os.ReadFile(fileFlag)
		if err != nil {
			fmt.Printf("❌ Не удалось прочитать файл: %v\n", err)
			fmt.Println("💡 Убедитесь что путь к файлу правильный")
			os.Exit(1)
		}

		var claimsData ClaimsData
		if err := json.Unmarshal(data, &claimsData); err != nil {
			fmt.Printf("❌ Ошибка парсинга JSON: %v\n", err)
			os.Exit(1)
		}

		if claimsData.Count == 0 {
			fmt.Println("⚠️  В файле нет утверждений для проверки")
			os.Exit(0)
		}

		fmt.Printf("✅ Загружено %d утверждений\n\n", claimsData.Count)

		fmt.Println("🔎 Проверка через Jina AI Grounding API...")
		api := NewJinaClient(apiKey)
		results, err := api.CheckClaims(claimsData.Claims)
		if err != nil {
			fmt.Printf("❌ Ошибка при проверке: %v\n", err)
			os.Exit(1)
		}

		printResults(claimsData, results)
	},
}

var fullCmd = &cobra.Command{
	Use:   "full",
	Short: "Full pipeline: extract claims and verify via Jina AI",
	Long: `Runs the complete hallucination detection pipeline:
1. Extracts claims from the LLM response via Python API (langextract + Gemini)
2. Saves claims to a JSON file
3. Verifies each claim via Jina AI Grounding API
4. Prints results with sources and confidence scores

Requires GEMINI_API_KEY and JINA_API_KEY environment variables.

Example:
  detector full -q "Когда была битва?" -r "Куликовская битва была в 1480 году"`,
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("GEMINI_API_KEY") == "" {
			fmt.Println("❌ GEMINI_API_KEY не установлен")
			fmt.Println("💡 Получите ключ: https://aistudio.google.com/app/apikey")
			os.Exit(1)
		}

		jinaKey := os.Getenv("JINA_API_KEY")
		if jinaKey == "" {
			fmt.Println("❌ JINA_API_KEY не установлен")
			fmt.Println("💡 Получите ключ: https://jina.ai/")
			os.Exit(1)
		}

		client := NewPythonClient("http://localhost:8000")
		fmt.Println("🔍 Проверка Python API...")
		if err := client.HealthCheck(); err != nil {
			fmt.Printf("❌ Python API недоступен: %v\n", err)
			fmt.Println("💡 Запустите: cd Python && python app.py")
			os.Exit(1)
		}
		fmt.Println("✅ Python API работает!")

		fmt.Println("\n📝 Извлечение утверждений...")
		result, err := client.ExtractAndSave(queryFlag, responseFlag)
		if err != nil {
			fmt.Printf("❌ Ошибка извлечения: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Сохранено в: %s\n", result.Filename)
		fmt.Printf("   Извлечено утверждений: %d\n\n", result.ClaimsCount)

		if result.ClaimsCount == 0 {
			fmt.Println("⚠️  Утверждений не найдено, завершение")
			os.Exit(0)
		}

		data, err := os.ReadFile(result.Filename)
		if err != nil {
			fmt.Printf("❌ Не удалось прочитать файл: %v\n", err)
			os.Exit(1)
		}

		var claimsData ClaimsData
		if err := json.Unmarshal(data, &claimsData); err != nil {
			fmt.Printf("❌ Ошибка парсинга JSON: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("🔎 Проверка через Jina AI Grounding API...")
		api := NewJinaClient(jinaKey)
		results, err := api.CheckClaims(claimsData.Claims)
		if err != nil {
			fmt.Printf("❌ Ошибка проверки: %v\n", err)
			os.Exit(1)
		}

		printResults(claimsData, results)
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

func printResults(claimsData ClaimsData, results []FactCheckResult) {
	fmt.Println("\n══════════════════════════════════════════════")
	fmt.Println("              РЕЗУЛЬТАТЫ ПРОВЕРКИ             ")
	fmt.Println("══════════════════════════════════════════════")

	if claimsData.Query != "" {
		fmt.Printf("\n📌 Запрос:   %s\n", claimsData.Query)
	}
	fmt.Printf("💬 Ответ:    %s\n", claimsData.Response)
	fmt.Println("\n──────────────────────────────────────────────")

	for i, result := range results {
		fmt.Printf("\n[%d] %s\n", i+1, result.Claim)

		if result.Found && result.Result {
			fmt.Printf("    ✅ ФАКТ ПОДТВЕРЖДЁН (достоверность: %.0f%%)\n", result.Factuality*100)
		} else if result.Found && !result.Result {
			fmt.Printf("    ❌ ГАЛЛЮЦИНАЦИЯ (достоверность: %.0f%%)\n", result.Factuality*100)
		} else {
			fmt.Println("    ⚠️  Не удалось проверить")
		}

		if result.Reason != "" {
			fmt.Printf("    💬 Объяснение: %s\n", result.Reason)
		}
		if result.ReviewURL != "" {
			fmt.Printf("    🔗 Источник:   %s\n", result.ReviewURL)
		}
		if result.KeyQuote != "" {
			fmt.Printf("    📝 Цитата:     \"%s\"\n", result.KeyQuote)
		}
	}

	summary := BuildSummary(results)

	fmt.Println("\n══════════════════════════════════════════════")
	fmt.Println("                    СВОДКА                    ")
	fmt.Println("══════════════════════════════════════════════")
	fmt.Printf("📊 Всего утверждений:      %d\n", summary.TotalClaims)
	fmt.Printf("✅ Подтверждено:            %d\n", summary.ClaimsFound)
	fmt.Printf("❌ Не подтверждено:         %d\n", summary.ClaimsNotFound)

	if summary.TotalClaims > 0 {
		pct := float64(summary.PotentialHallucinations) / float64(summary.TotalClaims) * 100
		fmt.Printf("⚠️  Возможных галлюцинаций: %d (%.1f%%)\n", summary.PotentialHallucinations, pct)
	}

	fmt.Println("══════════════════════════════════════════════")
}

func init() {
	verifyCmd.Flags().StringVarP(&queryFlag, "query", "q", "", "Запрос пользователя")
	verifyCmd.Flags().StringVarP(&responseFlag, "response", "r", "", "Ответ LLM для проверки (обязательно)")
	verifyCmd.MarkFlagRequired("response")

	checkCmd.Flags().StringVarP(&fileFlag, "file", "f", "", "Путь к JSON файлу с утверждениями (обязательно)")
	checkCmd.Flags().StringVarP(&keyFlag, "key", "k", "", "Jina AI API ключ (или JINA_API_KEY)")
	checkCmd.MarkFlagRequired("file")

	fullCmd.Flags().StringVarP(&queryFlag, "query", "q", "", "Запрос пользователя")
	fullCmd.Flags().StringVarP(&responseFlag, "response", "r", "", "Ответ LLM для проверки (обязательно)")
	fullCmd.MarkFlagRequired("response")
}

func cli() {
	rootCmd.AddCommand(verifyCmd, batchCmd, evaluateCmd, buildIndexCmd, healthCmd, checkCmd, fullCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
