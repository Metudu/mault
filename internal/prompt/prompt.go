package prompt

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"mault/internal/cerror"
	"os"
	"regexp"
	"strings"
	"time"
)

// Config holds configuration for prompts
type Config struct {
	MaxLength    int
	MinLength    int
	AllowEmpty   bool
	Timeout      time.Duration
	Retries      int
	ValidateFunc func(string) error
}

// DefaultConfig returns default prompt configuration
func DefaultConfig() *Config {
	return &Config{
		MaxLength:  255,
		MinLength:  1,
		AllowEmpty: false,
		Timeout:    30 * time.Second,
		Retries:    3,
	}
}

// KeyConfig returns configuration specific to secret keys
func KeyConfig() *Config {
	config := DefaultConfig()
	config.MaxLength = 100
	config.MinLength = 1
	config.ValidateFunc = validateSecretKey
	return config
}

// Prompter handles user input prompting
type Prompter struct {
	reader io.Reader
	writer io.Writer
	config *Config
}

// NewPrompter creates a new prompter with default configuration
func NewPrompter(reader io.Reader, writer io.Writer) *Prompter {
	return &Prompter{
		reader: reader,
		writer: writer,
		config: DefaultConfig(),
	}
}

// NewPrompterWithConfig creates a new prompter with custom configuration
func NewPrompterWithConfig(reader io.Reader, writer io.Writer, config *Config) *Prompter {
	return &Prompter{
		reader: reader,
		writer: writer,
		config: config,
	}
}

// GetKey prompts for a secret key with validation
func (p *Prompter) GetKey(prompt string) (string, error) {
	if prompt == "" {
		prompt = "Enter the name of the secret"
	}
	
	config := KeyConfig()
	return p.promptWithValidation(prompt, config)
}

// GetInput prompts for generic input with validation
func (p *Prompter) GetInput(prompt string) (string, error) {
	return p.promptWithValidation(prompt, p.config)
}

// GetInputWithConfig prompts for input with custom configuration
func (p *Prompter) GetInputWithConfig(prompt string, config *Config) (string, error) {
	return p.promptWithValidation(prompt, config)
}

// GetConfirmation prompts for yes/no confirmation
func (p *Prompter) GetConfirmation(prompt string) (bool, error) {
	fullPrompt := fmt.Sprintf("%s (y/N)", prompt)
	
	config := &Config{
		MaxLength:  10,
		MinLength:  1,
		AllowEmpty: true,
		Timeout:    p.config.Timeout,
		Retries:    p.config.Retries,
	}
	
	response, err := p.promptWithValidation(fullPrompt, config)
	if err != nil {
		return false, err
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// promptWithValidation handles the core prompting logic with validation
func (p *Prompter) promptWithValidation(prompt string, config *Config) (string, error) {
	scanner := bufio.NewScanner(p.reader)
	
	for attempt := 0; attempt < config.Retries; attempt++ {
		if attempt > 0 {
			fmt.Fprintf(p.writer, "Invalid input. Please try again.\n")
		}
		
		fmt.Fprintf(p.writer, "%s: ", prompt)
		
		// Handle timeout if reader supports it
		if config.Timeout > 0 {
			if err := p.scanWithTimeout(scanner, config.Timeout); err != nil {
				return "", &cerror.Error{
					Operation: "Get user input",
					Cause:     fmt.Sprintf("timeout after %v: %v", config.Timeout, err),
				}
			}
		} else {
			if !scanner.Scan() {
				return "", &cerror.Error{
					Operation: "Get user input",
					Cause:     "failed to scan input",
				}
			}
		}
		
		if err := scanner.Err(); err != nil {
			return "", &cerror.Error{
				Operation: "Get user input",
				Cause:     fmt.Sprintf("scanner error: %v", err),
			}
		}
		
		input := strings.TrimSpace(scanner.Text())
		
		// Validate input
		if err := p.validateInput(input, config); err != nil {
			if attempt == config.Retries-1 {
				return "", err
			}
			fmt.Fprintf(p.writer, "Error: %v\n", err)
			continue
		}
		
		return input, nil
	}
	
	return "", &cerror.Error{
		Operation: "Get user input",
		Cause:     fmt.Sprintf("exceeded maximum retries (%d)", config.Retries),
	}
}

// scanWithTimeout attempts to implement timeout for scanning
func (p *Prompter) scanWithTimeout(scanner *bufio.Scanner, timeout time.Duration) error {
	// Note: This is a simplified implementation. In a real-world scenario,
	// you might want to use a more sophisticated approach with goroutines
	// and channels for true timeout handling with os.Stdin
	
	done := make(chan bool, 1)
	go func() {
		scanner.Scan()
		done <- true
	}()
	
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("input timeout")
	}
}

// validateInput validates user input against configuration
func (p *Prompter) validateInput(input string, config *Config) error {
	// Check if empty input is allowed
	if input == "" {
		if !config.AllowEmpty {
			return cerror.ErrEmptyKey
		}
		return nil
	}
	
	// Check length constraints
	if len(input) < config.MinLength {
		return &cerror.Error{
			Operation: "Validate input",
			Cause:     fmt.Sprintf("input too short (minimum %d characters)", config.MinLength),
		}
	}
	
	if len(input) > config.MaxLength {
		return &cerror.Error{
			Operation: "Validate input",
			Cause:     fmt.Sprintf("input too long (maximum %d characters)", config.MaxLength),
		}
	}
	
	// Apply custom validation function if provided
	if config.ValidateFunc != nil {
		if err := config.ValidateFunc(input); err != nil {
			return err
		}
	}
	
	return nil
}

// validateSecretKey validates secret key format and content
func validateSecretKey(key string) error {
	// Check for invalid characters
	validKeyRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validKeyRegex.MatchString(key) {
		return &cerror.Error{
			Operation: "Validate secret key",
			Cause:     "key can only contain letters, numbers, underscores, and hyphens",
		}
	}
	
	// Check if key starts with a letter or underscore
	if !regexp.MustCompile(`^[a-zA-Z_]`).MatchString(key) {
		return &cerror.Error{
			Operation: "Validate secret key",
			Cause:     "key must start with a letter or underscore",
		}
	}
	
	// Reserved key names
	reservedKeys := []string{"help", "version", "init", "list", "delete", "update", "create"}
	keyLower := strings.ToLower(key)
	for _, reserved := range reservedKeys {
		if keyLower == reserved {
			return &cerror.Error{
				Operation: "Validate secret key",
				Cause:     fmt.Sprintf("'%s' is a reserved key name", key),
			}
		}
	}
	
	return nil
}

// Legacy functions for backward compatibility

// GetKey is a backward-compatible function that uses the original interface
func GetKey(reader io.Reader) (string, error) {
	prompter := NewPrompter(reader, os.Stdout)
	return prompter.GetKey("")
}

// GetKeyWithPrompt allows custom prompt message
func GetKeyWithPrompt(reader io.Reader, prompt string) (string, error) {
	prompter := NewPrompter(reader, os.Stdout)
	return prompter.GetKey(prompt)
}

// GetInput prompts for generic input
func GetInput(reader io.Reader, prompt string) (string, error) {
	prompter := NewPrompter(reader, os.Stdout)
	return prompter.GetInput(prompt)
}

// GetConfirmation prompts for yes/no confirmation
func GetConfirmation(reader io.Reader, prompt string) (bool, error) {
	prompter := NewPrompter(reader, os.Stdout)
	return prompter.GetConfirmation(prompt)
}

// ContextualPrompter supports context-aware prompting
type ContextualPrompter struct {
	*Prompter
}

// NewContextualPrompter creates a context-aware prompter
func NewContextualPrompter(reader io.Reader, writer io.Writer, config *Config) *ContextualPrompter {
	return &ContextualPrompter{
		Prompter: NewPrompterWithConfig(reader, writer, config),
	}
}

// GetKeyWithContext prompts for a key with context cancellation support
func (cp *ContextualPrompter) GetKeyWithContext(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		prompt = "Enter the name of the secret"
	}
	
	return cp.getInputWithContext(ctx, prompt, KeyConfig())
}

// GetInputWithContext prompts for input with context cancellation support
func (cp *ContextualPrompter) GetInputWithContext(ctx context.Context, prompt string) (string, error) {
	return cp.getInputWithContext(ctx, prompt, cp.config)
}

// getInputWithContext handles context-aware input with cancellation
func (cp *ContextualPrompter) getInputWithContext(ctx context.Context, prompt string, config *Config) (string, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	
	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	
	// Run the prompt in a goroutine
	go func() {
		result, err := cp.promptWithValidation(prompt, config)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- result
		}
	}()
	
	// Wait for either result or context cancellation
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}