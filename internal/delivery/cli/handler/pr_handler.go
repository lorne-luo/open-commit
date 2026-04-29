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

	"github.com/lorne-luo/open-commit/internal/service"
	"github.com/lorne-luo/open-commit/internal/usecase"
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
	maxDiffLines *int,
) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		if *quiet && !*noConfirm {
			*quiet = false
		}

		primary := service.ProviderConfig{
			Key:     viper.GetString("api.key"),
			BaseURL: derefString(customBaseUrl),
			Model:   derefString(model),
		}
		providers := service.BuildProviders(primary)

		if len(providers) == 0 {
			fmt.Println(
				"Error: API key is still empty, run this command to set your API key",
			)
			fmt.Print("\n")
			color.New(color.Bold).Print("opencommit config set api.key ")
			color.New(color.Italic, color.Bold).Print("your_api_key\n\n")
			os.Exit(1)
		}

		*model = providers[0].Model

		err := p.useCase.PRCommand(
			ctx,
			providers,
			model,
			noConfirm,
			quiet,
			dryRun,
			showDiff,
			maxLength,
			language,
			userContext,
			draft,
			maxDiffLines,
		)
		cobra.CheckErr(err)
	}
}
