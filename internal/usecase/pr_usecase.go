/*
Copyright © 2025 Christina Sørensen <ces@fem.gg>
*/
package usecase

import (
	"context"
	"sync"

	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"

	"github.com/tfkhdyt/geminicommit/internal/service"
)

type PRUsecase struct {
	gitService         *service.GitService
	aiService         *service.AIService
	interactionService *service.InteractionService
}

var (
	prUsecaseInstance *PRUsecase
	prUsecaseOnce     sync.Once
)

func NewPRUsecase() *PRUsecase {
	prUsecaseOnce.Do(func() {
		gitService := service.NewGitService()
		aiService := service.NewAIService()
		interactionService := service.NewInteractionService()

		prUsecaseInstance = &PRUsecase{
			gitService:         gitService,
			aiService:         aiService,
			interactionService: interactionService,
		}
	})

	return prUsecaseInstance
}

func (p *PRUsecase) initializeAIClient(
	ctx context.Context,
	apiKey string,
	customBaseUrl *string,
) *openai.Client {
	config := openai.DefaultConfig(apiKey)
	if customBaseUrl != nil && *customBaseUrl != "" {
		config.BaseURL = *customBaseUrl
	}
	client := openai.NewClientWithConfig(config)
	return client
}

func (p *PRUsecase) PRCommand(
	ctx context.Context,
	apiKey string,
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
) error {
	client := p.initializeAIClient(ctx, apiKey, customBaseUrl)

	if err := p.gitService.VerifyGitInstallation(); err != nil {
		return err
	}

	if err := p.gitService.VerifyGitRepository(); err != nil {
		return err
	}

	opts := &service.CommitOptions{
		Model:       model,
		NoConfirm:   noConfirm,
		Quiet:       quiet,
		DryRun:      dryRun,
		ShowDiff:    showDiff,
		MaxLength:   maxLength,
		Language:    language,
		UserContext: userContext,
	}

	data, err := p.gitService.GetDiff()
	if err != nil {
		return err
	}

	if *opts.ShowDiff && !*opts.Quiet {
		p.interactionService.DisplayDiff(data.Diff)
	}

	for {
		message, err := p.aiService.GenerateCommitMessage(client, ctx, data, opts)
		if err != nil {
			return err
		}

		selectedAction, finalMessage, err := p.interactionService.HandleUserAction(
			message,
			opts,
		)
		if err != nil {
			return err
		}

		switch selectedAction {
		case service.ActionConfirm:
			if err := p.gitService.CreatePullRequest(
				finalMessage,
				opts.Quiet,
				opts.DryRun,
				draft,
			); err != nil {
				return err
			}
			return nil
		case service.ActionRegenerate:
			continue
		case service.ActionEditContext:
			continue
		case service.ActionCancel:
			color.New(color.FgRed).Println("Pull request cancelled")
			return nil
		}
	}
}
