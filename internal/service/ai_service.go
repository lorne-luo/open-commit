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

//go:embed file_selection_prompt.md
var fileSelectionPrompt string

//go:embed combined_prompt.md
var combinedPrompt string

type AIService struct {
	systemPrompt string
}

// CommitOptions contains options for commit generation
type CommitOptions struct {
	StageAll    *bool
	AutoSelect  *bool
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
	MaxDiffLines *int
}

// PreCommitData contains data about the changes to be committed
type PreCommitData struct {
	Files        []string
	Diff         string
	RelatedFiles map[string]string
	Issue        string
}

// SelectFilesAndGenerateCommitOptions contains optional parameters for SelectFilesAndGenerateCommit
type SelectFilesAndGenerateCommitOptions struct {
	UserContext  *string
	RelatedFiles *map[string]string
	MaxLength    *int
	Language     *string
	Issue        *string
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
	providers []ProviderConfig,
	ctx context.Context,
	data *PreCommitData,
	opts *CommitOptions,
) (string, error) {
	messageChan := make(chan string, 1)

	if !*opts.Quiet {
		if err := spinner.New().
			Title(fmt.Sprintf("AI is analyzing your changes. (Model: %s)", *opts.Model)).
			Action(func() {
				a.analyzeToChannel(providers, ctx, data, opts, messageChan)
			}).
			Run(); err != nil {
			return "", err
		}
	} else {
		a.analyzeToChannel(providers, ctx, data, opts, messageChan)
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
	providers []ProviderConfig,
	ctx context.Context,
	data *PreCommitData,
	opts *CommitOptions,
	messageChan chan string,
) {
	message, err := a.AnalyzeChanges(
		providers,
		ctx,
		data.Diff,
		opts.UserContext,
		&data.RelatedFiles,
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

// formatRelatedFiles formats a map of directory to files into a slice of strings
// in the format "dir/file"
func formatRelatedFiles(dirToFiles map[string]string) []string {
	relatedFilesArray := make([]string, 0, len(dirToFiles))
	for dir, ls := range dirToFiles {
		relatedFilesArray = append(relatedFilesArray, fmt.Sprintf("%s/%s", dir, ls))
	}
	return relatedFilesArray
}

// chatCompleteOnce calls the OpenAI chat completion API on a single client with
// internal retry. It is the per-provider primitive used by chatCompleteFallback.
func chatCompleteOnce(
	client *openai.Client,
	ctx context.Context,
	model string,
	systemPrompt string,
	userPrompt string,
) (string, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
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
			lastErr = err
			continue
		}
		if len(resp.Choices) == 0 {
			lastErr = fmt.Errorf("no response from AI model")
			continue
		}
		text := strings.TrimSpace(resp.Choices[0].Message.Content)
		if text == "" {
			lastErr = fmt.Errorf("empty response text from model")
			continue
		}
		return text, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("failed to get response from AI model")
	}
	return "", lastErr
}

func (a *AIService) AnalyzeChanges(
	providers []ProviderConfig,
	ctx context.Context,
	diff string,
	userContext *string,
	relatedFiles *map[string]string,
	maxLength *int,
	language *string,
	issue *string,
) (string, error) {
	relatedFilesArray := formatRelatedFiles(*relatedFiles)

	userPrompt, err := a.GetUserPrompt(userContext, diff, relatedFilesArray, maxLength, language, issue)
	if err != nil {
		return "", err
	}

	enhancedSystemPrompt := a.systemPrompt
	if *language != "english" {
		enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Generate the commit message in %s language.", *language)
	}
	enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Keep the commit message under %d characters.", *maxLength)
	if *issue != "" {
		enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Reference issue %s in the commit message.", *issue)
	}

	result, err := chatCompleteFallback(ctx, providers, enhancedSystemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	result = strings.ReplaceAll(result, "```", "")
	result = strings.TrimSpace(result)

	return result, nil
}

// SelectFilesUsingAI lets the AI determine which files to stage based on the diff and context
func (a *AIService) SelectFilesUsingAI(
	providers []ProviderConfig,
	ctx context.Context,
	diff string,
	userContext *string,
) ([]string, error) {
	prompt := fmt.Sprintf(
		`%s
Here's the code diff:
%s`,
		*userContext,
		diff,
	)

	result, err := chatCompleteFallback(ctx, providers, fileSelectionPrompt, prompt)
	if err != nil {
		return nil, err
	}

	// Look for the file list in the response with flexible matching
	var filesStr string
	if after, ok := strings.CutPrefix(result, "FILES:"); ok {
		filesStr = after
	} else {
		lines := strings.Split(result, "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if after, ok := strings.CutPrefix(trimmedLine, "FILES:"); ok {
				filesStr = after
				break
			}
		}

		if filesStr == "" {
			if idx := strings.Index(result, ":"); idx != -1 {
				potential := result[idx+1:]
				if strings.Contains(potential, ".") || strings.Contains(potential, "/") || strings.Contains(potential, ",") {
					filesStr = potential
				}
			}
		}
	}

	if filesStr == "" {
		return nil, fmt.Errorf("AI response did not include file list in expected format. Response was: %s", result)
	}

	files := strings.Split(filesStr, ",")
	for i, f := range files {
		f = strings.Trim(f, "` \t\n\r")
		files[i] = strings.TrimSpace(f)
	}

	var validFiles []string
	for _, f := range files {
		if f != "" {
			validFiles = append(validFiles, f)
		}
	}

	return validFiles, nil
}

// SelectFilesAndGenerateCommit combines file selection and commit message generation in a single AI request
func (a *AIService) SelectFilesAndGenerateCommit(
	providers []ProviderConfig,
	ctx context.Context,
	diff string,
	opts *SelectFilesAndGenerateCommitOptions,
) ([]string, string, error) {
	if opts == nil {
		return nil, "", fmt.Errorf("options cannot be nil")
	}
	if opts.MaxLength == nil {
		return nil, "", fmt.Errorf("MaxLength cannot be nil")
	}
	if opts.Language == nil {
		return nil, "", fmt.Errorf("Language cannot be nil")
	}
	if opts.RelatedFiles == nil {
		return nil, "", fmt.Errorf("RelatedFiles cannot be nil")
	}

	relatedFilesArray := formatRelatedFiles(*opts.RelatedFiles)

	contextStr := ""
	if opts.UserContext != nil && *opts.UserContext != "" {
		contextStr = fmt.Sprintf("Use the following context to understand intent: %s\n\n", *opts.UserContext)
	}

	prompt := fmt.Sprintf(
		`%sHere's the code diff:
%s

Neighboring files:
%s

Requirements:
- Maximum commit message length: %d characters
- Language: %s`,
		contextStr,
		diff,
		strings.Join(relatedFilesArray, ", "),
		*opts.MaxLength,
		*opts.Language,
	)

	if opts.Issue != nil && *opts.Issue != "" {
		prompt += fmt.Sprintf("\n- Reference issue: %s", *opts.Issue)
	}

	enhancedSystemPrompt := combinedPrompt
	if *opts.Language != "english" {
		enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Generate the commit message in %s language.", *opts.Language)
	}
	enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Keep the commit message under %d characters.", *opts.MaxLength)
	if opts.Issue != nil && *opts.Issue != "" {
		enhancedSystemPrompt += fmt.Sprintf("\n\nIMPORTANT: Reference issue %s in the commit message.", *opts.Issue)
	}

	result, err := chatCompleteFallback(ctx, providers, enhancedSystemPrompt, prompt)
	if err != nil {
		return nil, "", err
	}

	// Parse files from response
	var filesStr string
	lines := strings.Split(result, "\n")
	foundFilesSection := false
	var filesLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "FILES:") {
			foundFilesSection = true
			afterPrefix := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "FILES:"))
			if afterPrefix != "" {
				filesLines = append(filesLines, afterPrefix)
			}
			continue
		}

		if foundFilesSection {
			if strings.HasPrefix(trimmedLine, "COMMIT_MESSAGE:") {
				break
			}
			filesLines = append(filesLines, line)
		}
	}

	if len(filesLines) > 0 {
		filesStr = strings.TrimSpace(strings.Join(filesLines, " "))
	} else {
		if after, ok := strings.CutPrefix(result, "FILES:"); ok {
			if idx := strings.Index(after, "COMMIT_MESSAGE:"); idx != -1 {
				filesStr = strings.TrimSpace(after[:idx])
			} else {
				filesStr = strings.TrimSpace(after)
			}
		}
	}

	if filesStr == "" {
		return nil, "", fmt.Errorf("AI response did not include file list in expected format. Response was: %s", result)
	}

	files := strings.Split(filesStr, ",")
	for i, f := range files {
		f = strings.Trim(f, "` \t\n\r")
		files[i] = strings.TrimSpace(f)
	}

	var validFiles []string
	for _, f := range files {
		if f != "" {
			validFiles = append(validFiles, f)
		}
	}

	// Parse commit message from response
	var commitMessage string
	foundCommitSection := false
	var commitLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "COMMIT_MESSAGE:") {
			foundCommitSection = true
			afterPrefix := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "COMMIT_MESSAGE:"))
			if afterPrefix != "" {
				commitLines = append(commitLines, afterPrefix)
			}
			continue
		}

		if foundCommitSection {
			if strings.HasPrefix(trimmedLine, "FILES:") {
				break
			}
			commitLines = append(commitLines, line)
		}
	}

	if len(commitLines) > 0 {
		commitMessage = strings.Join(commitLines, "\n")
		commitMessage = strings.ReplaceAll(commitMessage, "```", "")
		commitMessage = strings.TrimSpace(commitMessage)
	}

	if commitMessage == "" {
		return nil, "", fmt.Errorf("AI response did not include commit message in expected format. Response was: %s", result)
	}

	return validFiles, commitMessage, nil
}
