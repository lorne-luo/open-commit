package usecase

import (
	"context"
	"sync"

	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"

	"github.com/tfkhdyt/geminicommit/internal/service"
)

type RootUsecase struct {
	gitService         *service.GitService
	aiService         *service.AIService
	interactionService *service.InteractionService
}

var (
	rootUsecaseInstance *RootUsecase
	rootUsecaseOnce     sync.Once
)

func NewRootUsecase() *RootUsecase {
	rootUsecaseOnce.Do(func() {
		gitService := service.NewGitService()
		aiService := service.NewAIService()
		interactionService := service.NewInteractionService()

		rootUsecaseInstance = &RootUsecase{
			gitService:         gitService,
			aiService:         aiService,
			interactionService: interactionService,
		}
	})

	return rootUsecaseInstance
}

func (r *RootUsecase) initializeAIClient(ctx context.Context, apiKey string, customBaseUrl *string) *openai.Client {
	config := openai.DefaultConfig(apiKey)
	if customBaseUrl != nil && *customBaseUrl != "" {
		config.BaseURL = *customBaseUrl
	}
	client := openai.NewClientWithConfig(config)
	return client
}

func (r *RootUsecase) RootCommand(
	ctx context.Context,
	apiKey string,
	stageAll *bool,
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
) error {
	// Initialize AI client
	client := r.initializeAIClient(ctx, apiKey, customBaseUrl)

	// Perform git verifications
	if err := r.gitService.VerifyGitInstallation(); err != nil {
		return err
	}

	if err := r.gitService.VerifyGitRepository(); err != nil {
		return err
	}

	// Prepare commit options
	opts := &service.CommitOptions{
		StageAll:    stageAll,
		UserContext: userContext,
		Model:       model,
		NoConfirm:   noConfirm,
		Quiet:       quiet,
		Push:        push,
		DryRun:      dryRun,
		ShowDiff:    showDiff,
		MaxLength:   maxLength,
		Language:    language,
		Issue:       issue,
		NoVerify:    noVerify,
	}

	// Detect and prepare changes
	data, err := r.gitService.DetectAndPrepareChanges(opts)
	if err != nil {
		return err
	}

	// Display detected files
	r.interactionService.DisplayDetectedFiles(data.Files, opts.Quiet)

	// Show diff if requested
	if *opts.ShowDiff && !*opts.Quiet {
		r.interactionService.DisplayDiff(data.Diff)
	}

	// Main generation loop
	for {
		message, err := r.aiService.GenerateCommitMessage(client, ctx, data, opts)
		if err != nil {
			return err
		}

		selectedAction, finalMessage, err := r.interactionService.HandleUserAction(message, opts)
		if err != nil {
			return err
		}

		switch selectedAction {
		case service.ActionConfirm:
			if err := r.gitService.ConfirmAction(finalMessage, opts.Quiet, opts.Push, opts.DryRun, opts.NoVerify); err != nil {
				return err
			}
			return nil
		case service.ActionRegenerate:
			continue
		case service.ActionEditContext:
			continue
		case service.ActionCancel:
			color.New(color.FgRed).Println("Commit cancelled")
			return nil
		}
	}
}
