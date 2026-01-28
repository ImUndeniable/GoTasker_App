package ai

import (
	"context"
	"errors"
)

type OpenAIService struct{}

func NewOpenAIService() *OpenAIService {
	return &OpenAIService{}
}

func (s *OpenAIService) SummarizeTasks(
	ctx context.Context,
	input TaskSummaryInput,
) (TaskSummaryOutput, error) {

	if len(input.Titles) == 0 {
		return TaskSummaryOutput{}, errors.New("no tasks to summarize")
	}

	// TEMP MOCK (replace with real OpenAI call later)
	summary := "You have " + string(len(input.Titles)+'0') + " tasks. Focus on the most urgent ones."

	return TaskSummaryOutput{Summary: summary}, nil
}
