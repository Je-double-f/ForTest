package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// formatAndValidateKey обрабатывает ключ
func formatAndValidateKey(input string) (string, error) {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, " ", "_")
	input = strings.ToUpper(input)

	match, _ := regexp.MatchString(`^[A-Z_]+$`, input)
	if !match {
		return "", fmt.Errorf("ключ должен содержать только латинские буквы (без цифр и спецсимволов)")
	}

	if !strings.HasSuffix(input, "_KEY") {
		input += "_KEY"
	}

	return input, nil
}

// validateValue проверяет, что значение не содержит кириллицу
func validateValue(input string) (string, error) {
	input = strings.TrimSpace(input)
	match, _ := regexp.MatchString(`[а-яА-ЯёЁ]`, input)
	if match {
		return "", fmt.Errorf("значение должно быть только на латинице (без кириллицы)")
	}
	return input, nil
}

// promptValidInput — валидированный пользовательский ввод
func promptValidInput(promptText string, validator func(string) (string, error)) string {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(promptText)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		validated, err := validator(input)
		if err != nil {
			fmt.Println("❌ Ошибка:", err)
			continue
		}
		return validated
	}
}

// readEnvFileToMap читает .env в карту ключ-значение
func readEnvFileToMap(filePath string) (map[string]string, error) {
	env := make(map[string]string)

	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return env, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return env, nil
}

// writeEnvMap сохраняет карту переменных в .env
func writeEnvMap(filePath string, envMap map[string]string) error {
	var newLines []string
	for k, v := range envMap {
		newLines = append(newLines, fmt.Sprintf("%s=%s", k, v))
	}
	return os.WriteFile(filePath, []byte(strings.Join(newLines, "\n")+"\n"), 0644)
}

// promptOverwriteConfirmation спрашивает подтверждение и даёт 3 попытки
func promptOverwriteConfirmation(originalInput, currentValue string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("⚠️  Ключ \"%s\" уже существует. Вы хотите изменить значение? (yes/no): ", originalInput)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.ToLower(strings.TrimSpace(confirm))

	if confirm != "yes" {
		fmt.Println("⛔ Отмена обновления.")
		return false
	}

	const maxAttempts = 3
	for attempts := 1; attempts <= maxAttempts; attempts++ {
		fmt.Printf("Введите текущее значение для подтверждения (попытка %d из %d): ", attempts, maxAttempts)
		entered, _ := reader.ReadString('\n')
		entered = strings.TrimSpace(entered)

		if entered == currentValue {
			return true
		}
		fmt.Println("❌ Значение не совпадает.")
	}

	fmt.Println("⛔ Превышено количество попыток. Обновление отклонено.")
	return false
}

// AddOrUpdateEnvVarSecure безопасно добавляет или обновляет переменнуюe
func AddOrUpdateEnvVarSecure(filePath, originalInput, key, value string) (added, updated bool, err error) {
	envMap, err := readEnvFileToMap(filePath)
	if err != nil {
		return false, false, err
	}

	currentValue, exists := envMap[key]
	if exists {
		if !promptOverwriteConfirmation(originalInput, currentValue) {
			return false, false, nil
		}
		envMap[key] = value
		err = writeEnvMap(filePath, envMap)
		return false, true, err
	}

	envMap[key] = value
	err = writeEnvMap(filePath, envMap)
	return true, false, err
}

func main() {
	fmt.Println("🔐 Безопасное добавление переменной в .env")

	// Получаем оригинальный ввод (до форматирования)
	originalKeyInput := promptValidInput("Введите ключ переменной (например: db password): ", func(s string) (string, error) {
		return s, nil
	})

	// Форматируем ключ отдельно
	formattedKey, err := formatAndValidateKey(originalKeyInput)
	if err != nil {
		fmt.Println("❌ Ошибка:", err)
		return
	}

	value := promptValidInput("Введите значение переменной (только латиница): ", validateValue)

	added, updated, err := AddOrUpdateEnvVarSecure(".env", originalKeyInput, formattedKey, value)
	if err != nil {
		fmt.Println("❌ Ошибка при обновлении:", err)
		return
	}

	if added {
		fmt.Println("✅ Переменная успешно добавлена.")
	} else if updated {
		fmt.Println("✅ Переменная успешно обновлена.")
	}
}
