package service

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/huh/spinner"
	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
)

//go:embed system_prompt.md
var systemPrompt string

type AIService struct {
	systemPrompt string
}

// CommitOptions contains options for commit generation
type CommitOptions struct {
	StageAll    *bool
	UserContext *string
	Model       *string
	NoConfirm   *bool
	Quiet       *bool
	Push        *bool
	DryRun      *bool
	ShowDiff    *bool
	MaxLength   *int
	Language    *string
	Issue       *string
	NoVerify    *bool
}

// PreCommitData contains data about the changes to be committed
type PreCommitData struct {
	Files        []string
	Diff         string
	RelatedFiles map[string]string
	Issue        string
}

var (
	aiService *AIService
	aiOnce    sync.Once
)

func NewAIService() *AIService {
	aiOnce.Do(func() {
		aiService = &AIService{
			systemPrompt: systemPrompt,
		}
	})

	return aiService
}

// GenerateCommitMessage creates a commit message using AI analysis with UI feedback
func (a *AIService) GenerateCommitMessage(
	client *openai.Client,
	ctx context.Context,
	data *PreCommitData,
	opts *CommitOptions,
) (string, error) {
	messageChan := make(chan string, 1)

	if !*opts.Quiet {
		if err := spinner.New().
			Title(fmt.Sprintf("AI is analyzing your changes. (Model: %s)", *opts.Model)).
			Action(func() {
				a.analyzeToChannel(client, ctx, data, opts, messageChan)
			}).
			Run(); err != nil {
			return "", err
		}
	} else {
		a.analyzeToChannel(client, ctx, data, opts, messageChan)
	}

	message := <-messageChan
	if !*opts.Quiet {
		underline := color.New(color.Underline)
		underline.Println("\nChanges analyzed!")
	}

	message = strings.TrimSpace(message)
	if message == "" {
		return "", fmt.Errorf("no commit messages were generated. try again")
	}

	return message, nil
}

// analyzeToChannel performs the actual AI analysis and sends result to channel
func (a *AIService) analyzeToChannel(
	client *openai.Client,
	ctx context.Context,
	data *PreCommitData,
	opts *CommitOptions,
	messageChan chan string,
) {
	message, err := a.AnalyzeChanges(
		client,
		ctx,
		data.Diff,
		opts.UserContext,
		&data.RelatedFiles,
		opts.Model,
		opts.MaxLength,
		opts.Language,
		&data.Issue,
	)
	if err != nil {
		messageChan <- ""
	} else {
		messageChan <- message
	}
}

func (a *AIService) GetUserPrompt(
	context *string,
	diff string,
	files []string,
	maxLength *int,
	language *string,
	issue *string,
	// lastCommits []string,
) (string, error) {
	if *context != "" {
		temp := fmt.Sprintf("Use the following context to understand intent: %s", *context)
		context = &temp
	} else {
		*context = ""
	}

	prompt := fmt.Sprintf(
		`%s

Code diff:
%s

Neighboring files:
%s

Requirements:
- Maximum commit message length: %d characters
- Language: %s`,
		*context,
		diff,
		strings.Join(files, ", "),
		*maxLength,
		*language,
	)

	if *issue != "" {
		prompt += fmt.Sprintf("\n- Reference issue: %s", *issue)
	}

	return prompt, nil
}

func (a *AIService) AnalyzeChanges(
	openaiClient *openai.Client,
	ctx context.Context,
	diff string,
	userContext *string,
	relatedFiles *map[string]string,
	modelName *string,
	maxLength *int,
	language *string,
	issue *string,
	// lastCommits []string,
) (string, error) {
	// format relatedFiles to be dir : files
	relatedFilesArray := make([]string, 0, len(*relatedFiles))
	for dir, ls := range *relatedFiles {
		relatedFilesArray = append(relatedFilesArray, fmt.Sprintf("%s/%s", dir, ls))
	}

	userPrompt, err := a.GetUserPrompt(userContext, diff, relatedFilesArray, maxLength, language, issue)
	if err != nil {
		return "", err
	}

	// Update system prompt to include language and length requirements
	enhancedSystemPrompt := a.systemPrompt
	if *language != "english" {
		enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Generate the commit message in %s language.", *language)
	}
	enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Keep the commit message under %d characters.", *maxLength)
	if *issue != "" {
		enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Reference issue %s in the commit message.", *issue)
	}

	resp, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: *modelName,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: enhancedSystemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Temperature: 0.2,
		MaxTokens:   1000,
	})
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI model")
	}

	result := resp.Choices[0].Message.Content
	result = strings.ReplaceAll(result, "```", "")
	result = strings.TrimSpace(result)

	return result, nil
}
