package ai

import "context"

type Service interface {
	SummarizeTasks(ctx context.Context, input TaskSummaryInput) (TaskSummaryOutput, error)
}
