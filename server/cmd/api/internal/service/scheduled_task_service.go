package service

import (
	"cmd/api/internal/domain"
	"cmd/api/internal/repository"
	"errors"
	"fmt"
	"time"
)

type ScheduledTaskService struct {
	repo repository.ScheduledTaskRepository
}

func NewScheduledTaskService(repo repository.ScheduledTaskRepository) *ScheduledTaskService {
	return &ScheduledTaskService{repo: repo}
}

func (s *ScheduledTaskService) CreateTask(task *domain.ScheduledTask) error {
	if task.UserID <= 0 {
		return errors.New("user_id is required")
	}
	if task.Description == "" {
		return errors.New("description is required")
	}
	if task.Prompt == "" {
		return errors.New("prompt is required")
	}
	if task.DayOfMonth < 1 || task.DayOfMonth > 31 {
		return errors.New("day_of_month must be between 1 and 31")
	}

	// Calculate next run date
	nextRun := calculateNextRun(task.DayOfMonth)
	nextRunStr := nextRun.Format("2006-01-02T15:04:05")
	task.NextRunAt = &nextRunStr

	return s.repo.CreateTask(task)
}

func (s *ScheduledTaskService) GetTaskByID(taskID int) (*domain.ScheduledTask, error) {
	return s.repo.GetTaskByID(taskID)
}

func (s *ScheduledTaskService) GetTasksByUserID(userID int) ([]*domain.ScheduledTask, error) {
	return s.repo.GetTasksByUserID(userID)
}

func (s *ScheduledTaskService) GetDueTasks() ([]*domain.ScheduledTask, error) {
	return s.repo.GetDueTasks()
}

func (s *ScheduledTaskService) UpdateTask(task *domain.ScheduledTask) error {
	return s.repo.UpdateTask(task)
}

func (s *ScheduledTaskService) ToggleTask(taskID int) error {
	task, err := s.repo.GetTaskByID(taskID)
	if err != nil {
		return err
	}
	task.Active = !task.Active
	if task.Active {
		nextRun := calculateNextRun(task.DayOfMonth)
		nextRunStr := nextRun.Format("2006-01-02T15:04:05")
		task.NextRunAt = &nextRunStr
	}
	return s.repo.UpdateTask(task)
}

func (s *ScheduledTaskService) DeleteTask(taskID int) error {
	return s.repo.DeleteTask(taskID)
}

func (s *ScheduledTaskService) MarkTaskRun(taskID int) error {
	now := time.Now().Format("2006-01-02T15:04:05")
	task, err := s.repo.GetTaskByID(taskID)
	if err != nil {
		return err
	}
	nextRun := calculateNextRun(task.DayOfMonth)
	// Make sure next run is at least tomorrow
	if !nextRun.After(time.Now()) {
		nextRun = nextRun.AddDate(0, 1, 0)
	}
	nextRunStr := nextRun.Format("2006-01-02T15:04:05")
	return s.repo.UpdateLastRun(taskID, now, nextRunStr)
}

func calculateNextRun(dayOfMonth int) time.Time {
	now := time.Now()
	year, month, _ := now.Date()

	// Try this month first
	nextRun := time.Date(year, month, dayOfMonth, 8, 0, 0, 0, now.Location())

	// Clamp to last day of month if needed
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, now.Location()).Day()
	if dayOfMonth > lastDay {
		nextRun = time.Date(year, month, lastDay, 8, 0, 0, 0, now.Location())
	}

	// If already past, move to next month
	if !nextRun.After(now) {
		nextMonth := month + 1
		nextYear := year
		if nextMonth > 12 {
			nextMonth = 1
			nextYear++
		}
		lastDay = time.Date(nextYear, nextMonth+1, 0, 0, 0, 0, 0, now.Location()).Day()
		day := dayOfMonth
		if day > lastDay {
			day = lastDay
		}
		nextRun = time.Date(nextYear, nextMonth, day, 8, 0, 0, 0, now.Location())
	}

	fmt.Printf("[Scheduler] Next run calculated: %s (day %d)\n", nextRun.Format("2006-01-02"), dayOfMonth)
	return nextRun
}
