package usecase

import (
	"context"
	"fmt"
	"sync"

	"github.com/charmbracelet/huh/spinner"
	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"

	"github.com/lorne-luo/opencommit/internal/service"
)

type RootUsecase struct {
	gitService         *service.GitService
	aiService          *service.AIService
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
			aiService:          aiService,
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
	return openai.NewClientWithConfig(config)
}

func (r *RootUsecase) RootCommand(
	ctx context.Context,
	apiKey string,
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
		AutoSelect:  autoSelect,
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

	// Display detected files (skip this in auto mode since AI will select a subset later)
	if !*opts.AutoSelect {
		r.interactionService.DisplayDetectedFiles(data.Files, opts.Quiet)
	}

	// Show diff if requested
	if *opts.ShowDiff && !*opts.Quiet {
		r.interactionService.DisplayDiff(data.Diff)
	}

	// Check if auto-select flag is set and handle accordingly
	var initialCommitMessage string
	if *opts.AutoSelect {
		// Auto flow: Select files with AI and generate commit message in one request
		autoResult, err := r.handleAutoFlow(client, ctx, data, opts)
		if err != nil {
			return err
		}
		data = autoResult.Data
		initialCommitMessage = autoResult.CommitMessage

		// In auto mode, stage only the selected files for the commit
		if err := r.gitService.ResetStaged(); err != nil {
			return fmt.Errorf("failed to reset staged files: %v", err)
		}

		if err := r.gitService.StageFiles(data.Files); err != nil {
			return fmt.Errorf("failed to stage selected files: %v", err)
		}
	}

	// Main generation loop
	message := initialCommitMessage
	for {
		if message == "" {
			var err error
			message, err = r.aiService.GenerateCommitMessage(client, ctx, data, opts)
			if err != nil {
				return err
			}
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
			message = ""
			continue
		case service.ActionEditContext:
			message = ""
			continue
		case service.ActionCancel:
			color.New(color.FgRed).Println("Commit cancelled")
			return nil
		}
	}
}

// AutoFlowResult contains both selected files and generated commit message
type AutoFlowResult struct {
	Data          *service.PreCommitData
	CommitMessage string
}

// handleAutoFlow implements the complete auto flow as per the flowchart
func (r *RootUsecase) handleAutoFlow(
	client *openai.Client,
	ctx context.Context,
	data *service.PreCommitData,
	opts *service.CommitOptions,
) (*AutoFlowResult, error) {
	selectFilesAndGenerateCommit := func() ([]string, string, error) {
		selectOpts := &service.SelectFilesAndGenerateCommitOptions{
			UserContext:  opts.UserContext,
			RelatedFiles: &data.RelatedFiles,
			ModelName:    opts.Model,
			MaxLength:    opts.MaxLength,
			Language:     opts.Language,
			Issue:        &data.Issue,
		}
		selectedFiles, commitMessage, err := r.aiService.SelectFilesAndGenerateCommit(
			client,
			ctx,
			data.Diff,
			selectOpts,
		)
		if err != nil {
			return nil, "", err
		}
		if commitMessage == "" {
			return nil, "", fmt.Errorf("AI returned an empty commit message")
		}
		return selectedFiles, commitMessage, nil
	}

	var selectedFiles []string
	var commitMessage string
	var err error

	if !*opts.Quiet {
		err = spinner.New().
			Title(fmt.Sprintf("AI is analyzing your changes. (Model: %s)", *opts.Model)).
			Action(func() {
				selectedFiles, commitMessage, err = selectFilesAndGenerateCommit()
			}).
			Run()
		if err != nil {
			return nil, err
		}
	} else {
		selectedFiles, commitMessage, err = selectFilesAndGenerateCommit()
		if err != nil {
			return nil, err
		}
	}

	action, confirmedFiles, err := r.interactionService.ConfirmAutoSelectedFiles(selectedFiles)
	if err != nil {
		return nil, err
	}

	switch action {
	case service.ActionCancel:
		return nil, fmt.Errorf("operation cancelled")
	case service.ActionEdit:
		editedFiles, err := r.interactionService.EditFileList(selectedFiles)
		if err != nil {
			return nil, err
		}
		newData := *data
		newData.Files = editedFiles
		return &AutoFlowResult{
			Data:          &newData,
			CommitMessage: commitMessage,
		}, nil
	case service.ActionConfirm, service.ActionAutoSelect:
		newData := *data
		newData.Files = confirmedFiles
		return &AutoFlowResult{
			Data:          &newData,
			CommitMessage: commitMessage,
		}, nil
	default:
		return nil, fmt.Errorf("unknown action: %v", action)
	}
}
