package service

import (
	"fmt"
	"log"
	"time"
)

type SchedulerService struct {
	taskService  *ScheduledTaskService
	agentService *AgentService
	stopChan     chan struct{}
}

func NewSchedulerService(taskService *ScheduledTaskService, agentService *AgentService) *SchedulerService {
	return &SchedulerService{
		taskService:  taskService,
		agentService: agentService,
		stopChan:     make(chan struct{}),
	}
}

func (s *SchedulerService) Start() {
	go func() {
		log.Println("[Scheduler] Started - checking for due tasks every 60 seconds")
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		// Check immediately on start
		s.checkAndRunTasks()

		for {
			select {
			case <-ticker.C:
				s.checkAndRunTasks()
			case <-s.stopChan:
				log.Println("[Scheduler] Stopped")
				return
			}
		}
	}()
}

func (s *SchedulerService) Stop() {
	close(s.stopChan)
}

func (s *SchedulerService) checkAndRunTasks() {
	tasks, err := s.taskService.GetDueTasks()
	if err != nil {
		log.Printf("[Scheduler] Error fetching due tasks: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	log.Printf("[Scheduler] Found %d due task(s)", len(tasks))

	for _, task := range tasks {
		log.Printf("[Scheduler] Running task #%d: %s", task.TaskID, task.Description)

		// Build the prompt - include template voucher info if available
		prompt := task.Prompt
		now := time.Now()
		currentPeriod := fmt.Sprintf("%d-%02d", now.Year(), now.Month())
		currentDate := now.Format("2006-01-02")

		prompt = fmt.Sprintf("%s\n\nAnvänd datum: %s, period: %s, user_id: %d",
			prompt, currentDate, currentPeriod, task.UserID)

		// Execute through the agent
		conversationID := fmt.Sprintf("scheduler-%d-%s", task.TaskID, currentDate)
		response, err := s.agentService.Chat(task.UserID, conversationID, prompt, nil)
		if err != nil {
			log.Printf("[Scheduler] Error running task #%d: %v", task.TaskID, err)
			continue
		}

		log.Printf("[Scheduler] Task #%d completed: %s", task.TaskID, truncate(response, 100))

		// Update last run time and calculate next run
		if err := s.taskService.MarkTaskRun(task.TaskID); err != nil {
			log.Printf("[Scheduler] Error updating task #%d run time: %v", task.TaskID, err)
		}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
