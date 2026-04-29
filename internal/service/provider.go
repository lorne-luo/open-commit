package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

// ProviderConfig describes one AI provider candidate.
type ProviderConfig struct {
	ID      int // 1 = primary (api.*), 2 = secondary (api2.*)
	Key     string
	BaseURL string
	Model   string
}

// BuildProviders returns the ordered provider list. `primary` is the
// fully-resolved api1 configuration (already merged with CLI flags by the
// caller). The secondary is loaded from viper (api2.*) and api.last_success
// determines order when both providers have a key.
//
// Providers with an empty Key are filtered out; the returned slice may be
// empty, in which case the caller should report a missing-key error.
func BuildProviders(primary ProviderConfig) []ProviderConfig {
	primary.ID = 1
	if primary.Model == "" {
		primary.Model = DefaultModel
	}

	secondary := ProviderConfig{
		ID:      2,
		Key:     viper.GetString("api2.key"),
		BaseURL: viper.GetString("api2.baseurl"),
		Model:   viper.GetString("api2.model"),
	}
	if secondary.Model == "" {
		secondary.Model = DefaultModel
	}

	var providers []ProviderConfig
	if primary.Key != "" {
		providers = append(providers, primary)
	}
	if secondary.Key != "" {
		providers = append(providers, secondary)
	}

	if len(providers) == 2 && viper.GetInt("api.last_success") == 2 {
		providers[0], providers[1] = providers[1], providers[0]
	}

	return providers
}

// newOpenAIClient builds a configured openai client for a single provider.
func newOpenAIClient(p ProviderConfig) *openai.Client {
	cfg := openai.DefaultConfig(p.Key)
	if p.BaseURL != "" {
		cfg.BaseURL = p.BaseURL
	}
	return openai.NewClientWithConfig(cfg)
}

// persistLastSuccess writes api.last_success to the config file when the value
// has changed. Write errors are reported on stderr but not returned — losing
// the hint is not fatal.
func persistLastSuccess(usedID int) {
	if viper.GetInt("api.last_success") == usedID {
		return
	}
	viper.Set("api.last_success", usedID)
	if err := viper.WriteConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to persist api.last_success: %v\n", err)
	}
}

// chatCompleteFallback runs the same chat-completion request against an ordered
// list of providers, falling back to the next one on error. It records
// api.last_success when a provider succeeds.
func chatCompleteFallback(
	ctx context.Context,
	providers []ProviderConfig,
	systemPrompt string,
	userPrompt string,
) (string, error) {
	if len(providers) == 0 {
		return "", fmt.Errorf("no AI providers configured")
	}

	var errs []string
	for _, p := range providers {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		client := newOpenAIClient(p)
		text, err := chatCompleteOnce(client, ctx, p.Model, systemPrompt, userPrompt)
		if err == nil {
			persistLastSuccess(p.ID)
			return text, nil
		}
		errs = append(errs, fmt.Sprintf("api%d: %v", p.ID, err))
	}
	return "", fmt.Errorf("all providers failed (in order): %s", strings.Join(errs, "; "))
}
