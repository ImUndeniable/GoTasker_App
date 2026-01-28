package ai

import (
	"context"
	"gotasker/internal/models"
	"strings"
	"time"

	"database/sql"
)

type Worker struct {
	db        *sql.DB
	aiService Service
}

func NewWorker(db *sql.DB, aiService Service) *Worker {
	return &Worker{
		db:        db,
		aiService: aiService,
	}
}

func (w *Worker) AnalyzeTask(ctx context.Context, task models.Task, totalTasks int, pendingTasks int) string {
	title := strings.ToLower(task.Title)

	var insights []string

	// 1. Status-based insight
	if task.Done {
		insights = append(insights, "Task completed successfully.")
	} else {
		insights = append(insights, "Task is still pending.")
	}

	// 2. Keyword-based categorization
	switch {
	case strings.Contains(title, "jwt") || strings.Contains(title, "auth") || strings.Contains(title, "login"):
		insights = append(insights, "This is an authentication-related task.")
	case strings.Contains(title, "db") || strings.Contains(title, "sql") || strings.Contains(title, "postgres"):
		insights = append(insights, "This task involves database work.")
	case strings.Contains(title, "redis") || strings.Contains(title, "cache"):
		insights = append(insights, "This task focuses on caching or performance.")
	case strings.Contains(title, "ai") || strings.Contains(title, "ml"):
		insights = append(insights, "This task relates to AI functionality.")
	}

	// 3. Time-based analysis
	age := time.Since(task.CreatedAt)
	if !task.Done && age > 48*time.Hour {
		insights = append(insights, "This task has been pending for a long time.")
	} else if age < 2*time.Hour {
		insights = append(insights, "This is a newly created task.")
	}

	// 4. User workload context
	if pendingTasks > 5 {
		insights = append(insights, "You have many pending tasks. Prioritization is recommended.")
	} else if pendingTasks <= 2 {
		insights = append(insights, "Your task load is manageable. Keep going!")
	}

	return strings.Join(insights, " ")
}
