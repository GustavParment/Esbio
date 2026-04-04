package repository

import (
	"cmd/api/internal/domain"
	"database/sql"
	"fmt"
)

type ScheduledTaskRepository interface {
	CreateTask(task *domain.ScheduledTask) error
	GetTaskByID(taskID int) (*domain.ScheduledTask, error)
	GetTasksByUserID(userID int) ([]*domain.ScheduledTask, error)
	GetDueTasks() ([]*domain.ScheduledTask, error)
	UpdateTask(task *domain.ScheduledTask) error
	UpdateLastRun(taskID int, lastRunAt string, nextRunAt string) error
	DeleteTask(taskID int) error
}

type scheduledTaskRepository struct {
	db *sql.DB
}

func NewScheduledTaskRepository(db *sql.DB) ScheduledTaskRepository {
	return &scheduledTaskRepository{db: db}
}

func (r *scheduledTaskRepository) CreateTask(task *domain.ScheduledTask) error {
	query := `
		INSERT INTO scheduled_tasks (user_id, description, prompt, template_voucher_id, day_of_month, active, next_run_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING task_id, created_at, updated_at
	`
	err := r.db.QueryRow(query,
		task.UserID,
		task.Description,
		task.Prompt,
		task.TemplateVoucherID,
		task.DayOfMonth,
		task.Active,
		task.NextRunAt,
	).Scan(&task.TaskID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create scheduled task: %w", err)
	}
	return nil
}

func (r *scheduledTaskRepository) GetTaskByID(taskID int) (*domain.ScheduledTask, error) {
	query := `
		SELECT task_id, user_id, description, prompt, template_voucher_id, day_of_month,
		       active, last_run_at, next_run_at, created_at, updated_at
		FROM scheduled_tasks WHERE task_id = $1
	`
	task := &domain.ScheduledTask{}
	err := r.db.QueryRow(query, taskID).Scan(
		&task.TaskID, &task.UserID, &task.Description, &task.Prompt,
		&task.TemplateVoucherID, &task.DayOfMonth, &task.Active,
		&task.LastRunAt, &task.NextRunAt, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("scheduled task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled task: %w", err)
	}
	return task, nil
}

func (r *scheduledTaskRepository) GetTasksByUserID(userID int) ([]*domain.ScheduledTask, error) {
	query := `
		SELECT task_id, user_id, description, prompt, template_voucher_id, day_of_month,
		       active, last_run_at, next_run_at, created_at, updated_at
		FROM scheduled_tasks WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.ScheduledTask
	for rows.Next() {
		task := &domain.ScheduledTask{}
		err := rows.Scan(
			&task.TaskID, &task.UserID, &task.Description, &task.Prompt,
			&task.TemplateVoucherID, &task.DayOfMonth, &task.Active,
			&task.LastRunAt, &task.NextRunAt, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scheduled task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *scheduledTaskRepository) GetDueTasks() ([]*domain.ScheduledTask, error) {
	query := `
		SELECT task_id, user_id, description, prompt, template_voucher_id, day_of_month,
		       active, last_run_at, next_run_at, created_at, updated_at
		FROM scheduled_tasks
		WHERE active = true AND next_run_at <= NOW()
		ORDER BY next_run_at
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get due tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.ScheduledTask
	for rows.Next() {
		task := &domain.ScheduledTask{}
		err := rows.Scan(
			&task.TaskID, &task.UserID, &task.Description, &task.Prompt,
			&task.TemplateVoucherID, &task.DayOfMonth, &task.Active,
			&task.LastRunAt, &task.NextRunAt, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scheduled task: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *scheduledTaskRepository) UpdateTask(task *domain.ScheduledTask) error {
	query := `
		UPDATE scheduled_tasks
		SET description = $1, prompt = $2, template_voucher_id = $3,
		    day_of_month = $4, active = $5, updated_at = NOW()
		WHERE task_id = $6
	`
	_, err := r.db.Exec(query,
		task.Description, task.Prompt, task.TemplateVoucherID,
		task.DayOfMonth, task.Active, task.TaskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update scheduled task: %w", err)
	}
	return nil
}

func (r *scheduledTaskRepository) UpdateLastRun(taskID int, lastRunAt string, nextRunAt string) error {
	query := `
		UPDATE scheduled_tasks SET last_run_at = $1, next_run_at = $2, updated_at = NOW()
		WHERE task_id = $3
	`
	_, err := r.db.Exec(query, lastRunAt, nextRunAt, taskID)
	if err != nil {
		return fmt.Errorf("failed to update last run: %w", err)
	}
	return nil
}

func (r *scheduledTaskRepository) DeleteTask(taskID int) error {
	_, err := r.db.Exec("DELETE FROM scheduled_tasks WHERE task_id = $1", taskID)
	if err != nil {
		return fmt.Errorf("failed to delete scheduled task: %w", err)
	}
	return nil
}
