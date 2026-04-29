/*
Copyright © 2025 Christina Sørensen <ces@fem.gg>
*/
package usecase

import (
	"context"
	"sync"

	"github.com/fatih/color"

	"github.com/lorne-luo/open-commit/internal/service"
)

type PRUsecase struct {
	gitService         *service.GitService
	aiService          *service.AIService
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
			aiService:          aiService,
			interactionService: interactionService,
		}
	})

	return prUsecaseInstance
}

func (p *PRUsecase) PRCommand(
	ctx context.Context,
	providers []service.ProviderConfig,
	model *string,
	noConfirm *bool,
	quiet *bool,
	dryRun *bool,
	showDiff *bool,
	maxLength *int,
	language *string,
	userContext *string,
	draft *bool,
	maxDiffLines *int,
) error {
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

	if maxDiffLines != nil && *maxDiffLines > 0 {
		original := data.Diff
		data.Diff = service.TruncateLargeDiffs(data.Diff, *maxDiffLines)
		if !*quiet && data.Diff != original {
			color.New(color.FgYellow).Printf(
				"⚠ Diff truncated to %d lines per file to save tokens\n",
				*maxDiffLines,
			)
		}
	}

	if *opts.ShowDiff && !*opts.Quiet {
		p.interactionService.DisplayDiff(data.Diff)
	}

	for {
		message, err := p.aiService.GenerateCommitMessage(providers, ctx, data, opts)
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
