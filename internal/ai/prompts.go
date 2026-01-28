package ai

import "strings"

func TaskSummaryPrompt(titles []string) string {
	return "Summarize the following tasks and suggest what to focus on:\n\n- " +
		strings.Join(titles, "\n- ")
}
