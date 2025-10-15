/*
Copyright © 2025 Christina Sørensen <ces@fem.gg>
*/
package handler

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lorne-luo/opencommit/internal/service"
	"github.com/lorne-luo/opencommit/internal/usecase"
)

type PRHandler struct {
	useCase *usecase.PRUsecase
}

var (
	prHandlerInstance *PRHandler
	prHandlerOnce     sync.Once
)

func NewPRHandler() *PRHandler {
	prHandlerOnce.Do(func() {
		useCase := usecase.NewPRUsecase()

		prHandlerInstance = &PRHandler{useCase}
	})

	return prHandlerInstance
}

func (p *PRHandler) PRCommand(
	ctx context.Context,
	model *string,
	noConfirm *bool,
	quiet *bool,
	dryRun *bool,
	showDiff *bool,
	maxLength *int,
	language *string,
	userContext *string,
	draft *bool,
	customBaseUrl *string,
) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		modelFromConfig := viper.GetString("api.model")
		if modelFromConfig != "" && *model == service.DefaultModel {
			*model = modelFromConfig
		}

		baseUrlFromConfig := viper.GetString("api.baseurl")
		if baseUrlFromConfig != "" && *customBaseUrl == service.DefaultBaseUrl {
			*customBaseUrl = baseUrlFromConfig
		}

		if *quiet && !*noConfirm {
			*quiet = false
		}

		apiKey := viper.GetString("api.key")
		if apiKey == "" {
			fmt.Println(
				"Error: API key is still empty, run this command to set your API key",
			)
			fmt.Print("\n")
			color.New(color.Bold).Print("opencommit config key set ")
			color.New(color.Italic, color.Bold).Print("your_api_key\n\n")
			os.Exit(1)
		}

		err := p.useCase.PRCommand(
			ctx,
			apiKey,
			model,
			noConfirm,
			quiet,
			dryRun,
			showDiff,
			maxLength,
			language,
			userContext,
			draft,
			customBaseUrl,
		)
		cobra.CheckErr(err)
	}
}
