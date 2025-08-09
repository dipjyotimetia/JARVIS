package prompt

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
)

type specType string

const (
	AVRO     specType = "avro"
	PROTOBUF specType = "protobuf"
	SWAGGER  specType = "swagger"
	OAS3     specType = "OpenAPI 3"
)

type PromptContent struct {
	ErrorMsg string
	Label    string
	ItemType string
	Items    []string
}

// Note: CompareSpecFiles function was removed as it was specific to the old Gemini API
// If spec comparison functionality is needed in the future, it can be implemented
// using the new Ollama API with string-based prompts.

func SelectLanguage(pc PromptContent) string {
	var items []string
	switch pc.ItemType {
	case "language":
		items = []string{"Go", "Python", "JavaScript", "java", "TypeScript"}
	case "spec":
		items = []string{"avro", "protobuf", "swagger", "openapi"}
	default:
		items = []string{}
		slog.Error("Invalid prompt type")
	}
	index := -1
	var result string
	var err error

	for index < 0 {
		prompt := promptui.SelectWithAdd{
			Label:    pc.Label,
			Items:    items,
			AddLabel: "Other",
		}

		index, result, err = prompt.Run()

		if index == -1 {
			items = append(items, result)
		}
	}

	if err != nil {
		color.Red(pc.ErrorMsg)
		os.Exit(1)
	}

	color.Green("Input: %s\n", result)

	return result
}

func setFrameworksForLanguage(language string) []string {
	switch language {
	case "Go":
		return []string{"Gin", "Echo", "Fiber", "gRPC"}
	case "Python":
		return []string{"Django", "Flask", "FastAPI", "gRPC"}
	case "Java":
		return []string{"Spring", "JAX-RS", "restassured", "gRPC"}
	case "JavaScript":
		return []string{"supertest", "axios", "http", "gRPC"}
	case "TypeScript":
		return []string{"supertest", "axios", "http"}
	default:
		return []string{}
	}
}

// SelectFramework allows user to select a framework based on the chosen language
func SelectFramework(language string) string {
	frameworks := setFrameworksForLanguage(language)
	if len(frameworks) == 0 {
		return ""
	}

	prompt := promptui.Select{
		Label: fmt.Sprintf("Select %s Framework", language),
		Items: frameworks,
		Size:  10,
	}

	index, result, err := prompt.Run()
	if err != nil {
		color.Red("Failed to select framework: %v", err)
		os.Exit(1)
	}

	color.Green("Selected framework: %s (index: %d)\n", result, index)
	return result
}

// Confirm presents a yes/no confirmation prompt to the user
func Confirm(label string, defaultValue bool) bool {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
		Default:   map[bool]string{true: "y", false: "n"}[defaultValue],
	}

	result, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrAbort {
			return false
		}
		color.Red("Prompt failed: %v", err)
		return defaultValue
	}

	// Convert result to boolean
	result = strings.ToLower(result)
	return result == "y" || result == "yes"
}

// Input prompts the user for text input with optional validation
func Input(label string, defaultValue string, validate func(string) error) string {
	prompt := promptui.Prompt{
		Label:    label,
		Default:  defaultValue,
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		color.Red("Prompt failed: %v", err)
		return defaultValue
	}

	return result
}

// InputNumber prompts for a number with validation
func InputNumber(label string, defaultValue int) int {
	validate := func(input string) error {
		_, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return fmt.Errorf("please enter a valid number")
		}
		return nil
	}

	defaultStr := ""
	if defaultValue != 0 {
		defaultStr = strconv.Itoa(defaultValue)
	}

	prompt := promptui.Prompt{
		Label:    label,
		Default:  defaultStr,
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		color.Red("Number input failed: %v", err)
		return defaultValue
	}

	num, _ := strconv.Atoi(result)
	return num
}

// SelectWithSearch creates a searchable select prompt
func SelectWithSearch(label string, items []string) (string, error) {
	searcher := func(input string, index int) bool {
		item := strings.ToLower(items[index])
		input = strings.ToLower(input)
		return strings.Contains(item, input)
	}

	prompt := promptui.Select{
		Label:    label,
		Items:    items,
		Size:     10,
		Searcher: searcher,
	}

	_, result, err := prompt.Run()
	return result, err
}

// Password prompts for a password with masking
func Password(label string) string {
	prompt := promptui.Prompt{
		Label:    label,
		Mask:     '*',
		Validate: nil,
	}

	result, err := prompt.Run()
	if err != nil {
		color.Red("Password input failed: %v", err)
		return ""
	}

	return result
}
