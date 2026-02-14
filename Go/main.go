// Go/main.go

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/muesli/termenv"
)

func main() {
	p := termenv.ColorProfile()

	colorPrompt := p.Color("#00BFFF")
	colorBg := p.Color("#0D1117")
	colorError := p.Color("#FF6B6B")
	colorDim := p.Color("#8B949E")

	printGradientLogo()

	fmt.Println(termenv.String("  Введите /help для списка команд. Ctrl+C для выхода.").Foreground(colorDim))
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Строка ввода с подложкой
		prompt := termenv.String(" > ").Foreground(colorPrompt).Background(colorBg).Bold()
		inputArea := termenv.String("                                                  ").Background(colorBg)
		fmt.Print(prompt, inputArea, "\r", prompt)

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := splitArgs(input)
		if len(parts) == 0 {
			continue
		}
		command := parts[0]

		switch command {
		case "help":
			printHelp(p)

		case "check":
			response := extractFlag(parts, "-r")
			if response == "" {
				fmt.Println(termenv.String("  ❌ Укажите ответ ИИ: /check -r \"текст ответа\"").Foreground(colorError))
				continue
			}
			runFull(response, p)

		case "verify":
			runVerify(p)

		case "exit", "quit":
			fmt.Println(termenv.String("\n  До свидания! 👋\n").Foreground(colorDim))
			os.Exit(0)

		default:
			fmt.Println(termenv.String(fmt.Sprintf("  ❌ Неизвестная команда: %s. Введите /help", command)).Foreground(colorError))
		}

		fmt.Println()
	}
}

// splitArgs разбивает строку с учётом кавычек
// /check -r "текст с пробелами" → ["/check", "-r", "текст с пробелами"]
func splitArgs(input string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, ch := range input {
		switch {
		case (ch == '"' || ch == '\'') && !inQuotes:
			inQuotes = true
			quoteChar = ch
		case ch == quoteChar && inQuotes:
			inQuotes = false
			quoteChar = 0
		case ch == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// extractFlag извлекает значение флага
func extractFlag(parts []string, flag string) string {
	for i, part := range parts {
		if part == flag && i+1 < len(parts) {
			return strings.Join(parts[i+1:], " ")
		}
	}
	return ""
}

// printHelp выводит список команд
func printHelp(p termenv.Profile) {
	colorCmd := p.Color("#00BFFF")
	colorFlag := p.Color("#79C0FF")
	colorDesc := p.Color("#E6EDF3")
	colorDim := p.Color("#8B949E")

	fmt.Println()
	fmt.Println(termenv.String("  ══════════════════════════════════════════").Foreground(colorDim))
	fmt.Println(termenv.String("                   КОМАНДЫ                  ").Foreground(colorDesc))
	fmt.Println(termenv.String("  ══════════════════════════════════════════").Foreground(colorDim))
	fmt.Println()
	fmt.Print(termenv.String("  /check").Foreground(colorCmd))
	fmt.Print(termenv.String(" -r").Foreground(colorFlag))
	fmt.Println(termenv.String(" \"<ответ ИИ>\"").Foreground(colorDim))
	fmt.Println(termenv.String("      Полный пайплайн: извлечь утверждения и проверить факты").Foreground(colorDesc))
	fmt.Println(termenv.String("      Пример: /check -r \"Куликовская битва была в 1480 году\"").Foreground(colorDim))
	fmt.Println()
	fmt.Println(termenv.String("  /verify").Foreground(colorCmd))
	fmt.Println(termenv.String("      Проверить готовность: API ключи и Python сервер").Foreground(colorDesc))
	fmt.Println()
	fmt.Println(termenv.String("  /help").Foreground(colorCmd))
	fmt.Println(termenv.String("      Показать этот список команд").Foreground(colorDesc))
	fmt.Println()
	fmt.Println(termenv.String("  /exit").Foreground(colorCmd))
	fmt.Println(termenv.String("      Выйти из программы").Foreground(colorDesc))
	fmt.Println()
	fmt.Println(termenv.String("  ══════════════════════════════════════════").Foreground(colorDim))
	fmt.Println(termenv.String("  Переменные окружения:").Foreground(colorDim))
	fmt.Println(termenv.String("    GEMINI_API_KEY  — для извлечения утверждений (langextract)").Foreground(colorDim))
	fmt.Println(termenv.String("    JINA_API_KEY    — для проверки фактов (Jina Grounding API)").Foreground(colorDim))
	fmt.Println(termenv.String("  ══════════════════════════════════════════").Foreground(colorDim))
}

// runVerify проверяет готовность системы
func runVerify(p termenv.Profile) {
	colorOk := p.Color("#3FB950")
	colorErr := p.Color("#FF6B6B")
	colorWarn := p.Color("#D29922")
	colorText := p.Color("#E6EDF3")

	fmt.Println()
	fmt.Println(termenv.String("  🔍 Проверка готовности системы...").Foreground(colorText))
	fmt.Println()

	if os.Getenv("GEMINI_API_KEY") != "" {
		fmt.Println(termenv.String("  ✅ GEMINI_API_KEY    — установлен").Foreground(colorOk))
	} else {
		fmt.Println(termenv.String("  ❌ GEMINI_API_KEY    — не установлен").Foreground(colorErr))
		fmt.Println(termenv.String("     💡 https://aistudio.google.com/app/apikey").Foreground(colorWarn))
	}

	if os.Getenv("JINA_API_KEY") != "" {
		fmt.Println(termenv.String("  ✅ JINA_API_KEY      — установлен").Foreground(colorOk))
	} else {
		fmt.Println(termenv.String("  ❌ JINA_API_KEY      — не установлен").Foreground(colorErr))
		fmt.Println(termenv.String("     💡 https://jina.ai/").Foreground(colorWarn))
	}

	client := NewPythonClient("http://localhost:8000")
	if err := client.HealthCheck(); err != nil {
		fmt.Println(termenv.String("  ❌ Python API        — недоступен").Foreground(colorErr))
		fmt.Println(termenv.String("     💡 cd Python && python app.py").Foreground(colorWarn))
	} else {
		fmt.Println(termenv.String("  ✅ Python API        — работает").Foreground(colorOk))
	}
}

// runFull запускает полный пайплайн
func runFull(response string, p termenv.Profile) {
	colorErr := p.Color("#FF6B6B")
	colorOk := p.Color("#3FB950")
	colorWarn := p.Color("#D29922")

	if os.Getenv("GEMINI_API_KEY") == "" {
		fmt.Println(termenv.String("  ❌ GEMINI_API_KEY не установлен").Foreground(colorErr))
		fmt.Println(termenv.String("  💡 https://aistudio.google.com/app/apikey").Foreground(colorWarn))
		return
	}

	jinaKey := os.Getenv("JINA_API_KEY")
	if jinaKey == "" {
		fmt.Println(termenv.String("  ❌ JINA_API_KEY не установлен").Foreground(colorErr))
		fmt.Println(termenv.String("  💡 https://jina.ai/").Foreground(colorWarn))
		return
	}

	client := NewPythonClient("http://localhost:8000")
	fmt.Println("  🔍 Проверка Python API...")
	if err := client.HealthCheck(); err != nil {
		fmt.Println(termenv.String(fmt.Sprintf("  ❌ Python API недоступен: %v", err)).Foreground(colorErr))
		fmt.Println(termenv.String("  💡 cd Python && python app.py").Foreground(colorWarn))
		return
	}
	fmt.Println(termenv.String("  ✅ Python API работает!").Foreground(colorOk))

	fmt.Println("\n  📝 Извлечение утверждений...")
	result, err := client.ExtractAndSave("", response)
	if err != nil {
		fmt.Println(termenv.String(fmt.Sprintf("  ❌ Ошибка извлечения: %v", err)).Foreground(colorErr))
		return
	}
	fmt.Println(termenv.String(fmt.Sprintf("  ✅ Сохранено в: %s", result.Filename)).Foreground(colorOk))
	fmt.Printf("     Извлечено утверждений: %d\n\n", result.ClaimsCount)

	if result.ClaimsCount == 0 {
		fmt.Println(termenv.String("  ⚠️  Утверждений не найдено").Foreground(colorWarn))
		return
	}

	data, err := os.ReadFile(result.Filename)
	if err != nil {
		fmt.Println(termenv.String(fmt.Sprintf("  ❌ Не удалось прочитать файл: %v", err)).Foreground(colorErr))
		return
	}

	var claimsData ClaimsData
	if err := json.Unmarshal(data, &claimsData); err != nil {
		fmt.Println(termenv.String(fmt.Sprintf("  ❌ Ошибка парсинга JSON: %v", err)).Foreground(colorErr))
		return
	}

	fmt.Println("  🔎 Проверка через Jina AI Grounding API...")
	api := NewJinaClient(jinaKey)
	results, err := api.CheckClaims(claimsData.Claims)
	if err != nil {
		fmt.Println(termenv.String(fmt.Sprintf("  ❌ Ошибка проверки: %v", err)).Foreground(colorErr))
		return
	}

	printResults(claimsData, results)
}
