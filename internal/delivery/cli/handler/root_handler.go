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

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type RootHandler struct {
	useCase *usecase.RootUsecase
}

var (
	rootHandlerInstance *RootHandler
	rootHandlerOnce     sync.Once
)

func NewRootHandler() *RootHandler {
	rootHandlerOnce.Do(func() {
		useCase := usecase.NewRootUsecase()

		rootHandlerInstance = &RootHandler{useCase}
	})

	return rootHandlerInstance
}

func (r *RootHandler) RootCommand(
	ctx context.Context,
	stageAll *bool,
	autoSelect *bool,
	userContext *string,
	model *string,
	noConfirm *bool,
	quiet *bool,
	push *bool,
	dryRun *bool,
	showDiff *bool,
	maxLength *int,
	language *string,
	issue *string,
	noVerify *bool,
	customBaseUrl *string,
	maxDiffLines *int,
) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		if *quiet && !*noConfirm {
			*quiet = false
		}

		// applyConfigDefaults has already merged config + CLI flags into
		// `*model` and `*customBaseUrl`, so they are the resolved primary values.
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

		// Reflect the resolved primary model back to the caller (for spinner display).
		*model = providers[0].Model

		err := r.useCase.RootCommand(
			ctx,
			providers,
			stageAll,
			autoSelect,
			userContext,
			model,
			noConfirm,
			quiet,
			push,
			dryRun,
			showDiff,
			maxLength,
			language,
			issue,
			noVerify,
			maxDiffLines,
		)
		cobra.CheckErr(err)
	}
}
