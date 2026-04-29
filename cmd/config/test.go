package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lorne-luo/open-commit/internal/service"
)

var testOnlyProvider int

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test each configured AI provider",
	Long: `Send a minimal request to each configured AI provider and report
whether it responded successfully. Does not modify api.last_success.

Examples:
  opencommit config test
  opencommit config test --provider 2`,
	Run: func(cmd *cobra.Command, args []string) {
		primary := service.ProviderConfig{
			ID:      1,
			Key:     viper.GetString("api.key"),
			BaseURL: viper.GetString("api.baseurl"),
			Model:   viper.GetString("api.model"),
		}
		if primary.Model == "" {
			primary.Model = service.DefaultModel
		}
		secondary := service.ProviderConfig{
			ID:      2,
			Key:     viper.GetString("api2.key"),
			BaseURL: viper.GetString("api2.baseurl"),
			Model:   viper.GetString("api2.model"),
		}
		if secondary.Model == "" {
			secondary.Model = service.DefaultModel
		}

		var candidates []service.ProviderConfig
		if testOnlyProvider == 0 || testOnlyProvider == 1 {
			candidates = append(candidates, primary)
		}
		if testOnlyProvider == 0 || testOnlyProvider == 2 {
			candidates = append(candidates, secondary)
		}
		if testOnlyProvider != 0 && testOnlyProvider != 1 && testOnlyProvider != 2 {
			fmt.Fprintln(os.Stderr, "Error: --provider must be 1 or 2")
			os.Exit(1)
		}

		anyFailed := false
		for _, p := range candidates {
			label := fmt.Sprintf("api%d", p.ID)
			if p.Key == "" {
				color.New(color.FgYellow).Printf("%s ⊘ skipped (no key configured)\n", label)
				continue
			}
			fmt.Printf("%s → model=%s baseurl=%s\n", label, p.Model, displayBaseURL(p.BaseURL))
			if err := pingProvider(p); err != nil {
				color.New(color.FgRed).Printf("  ✗ failed: %v\n", err)
				anyFailed = true
			} else {
				color.New(color.FgGreen).Println("  ✓ ok")
			}
		}

		if anyFailed {
			os.Exit(1)
		}
	},
}

func displayBaseURL(u string) string {
	if u == "" {
		return "(default)"
	}
	return u
}

func pingProvider(p service.ProviderConfig) error {
	cfg := openai.DefaultConfig(p.Key)
	if p.BaseURL != "" {
		cfg.BaseURL = p.BaseURL
	}
	client := openai.NewClientWithConfig(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.Model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: "ping"},
		},
		MaxTokens:   5,
		Temperature: 0,
	})
	if err != nil {
		return err
	}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("empty response (no choices)")
	}
	return nil
}

func init() {
	testCmd.Flags().IntVarP(&testOnlyProvider, "provider", "p", 0, "test only provider 1 or 2 (default: both)")
	ConfigCmd.AddCommand(testCmd)
}
