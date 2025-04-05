package scheduler

import (
	"backend/internal/clients"
	"backend/internal/dto"
	"backend/internal/service"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

type Scheduler struct {
	jobService    *service.JobService
	logger        *log.Logger
	jobCreated    chan struct{}
	stop          chan struct{}
	db            *sql.DB
	workerClient  *clients.WorkerClient
	taskScheduled chan struct{}
}

func NewScheduler(
	mtxService *service.JobService,
	logger *log.Logger,
	jobScheduled chan struct{},
	db *sql.DB,
	client *clients.WorkerClient,
) *Scheduler {
	s := &Scheduler{
		jobService:   mtxService,
		logger:       logger,
		jobCreated:   jobScheduled,
		db:           db,
		stop:         make(chan struct{}),
		workerClient: client,
	}
	return s
}

func (s *Scheduler) Start() {
	go s.taskCreator()
	go s.pollJobStatus()
}

func (s *Scheduler) Stop() {
	close(s.jobCreated)
}

func (s *Scheduler) taskCreator() {
	for {
		select {
		case <-s.jobCreated:
			err := s.runTasks()
			if err != nil {
				s.logger.Println("occurred error during creating task", err)
			}
		case <-time.After(5 * time.Second):
			err := s.runTasks()
			if err != nil {
				s.logger.Println("occurred error during creating task", err)
			}
		case <-s.stop:
			break
		}
	}
}

func (s *Scheduler) runTasks() error {
	status := "created"
	notScheduled, err := s.jobService.ListJobs(&status)
	if err != nil {
		return err
	}
	if len(notScheduled) == 0 {
		return errors.New("no scheduled jobs found")
	}
	id := notScheduled[0].ID
	job, err := s.jobService.GetJob(id)
	if err != nil {
		return err
	}
	s.logger.Println("handling created job", job.ID)
	req := dto.JobRequest{
		ID:      strconv.Itoa(job.ID),
		Content: &job.Content,
	}
	_, err = s.workerClient.Ping(req)
	if err != nil {
		s.logger.Println("occurred error during ping task", err)
		return err
	}
	s.logger.Println("updating task status to executing task", job.ID)
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Println("occurred error during begin transaction", err)
		return err
	}
	err = s.jobService.SetStatus(job.ID, "executing", tx)
	if err != nil {
		s.logger.Println("occurred error during set job status", err)
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *Scheduler) pollJobStatus() {
	status := "executing"
	for {
		jobs, err := s.jobService.ListJobs(&status)
		if err != nil {
			s.logger.Println("occurred error during getting running jobs", err)
			<-time.After(2 * time.Second)
			continue
		}
		if len(jobs) == 0 {
			s.logger.Println("no running jobs found")
			<-time.After(10 * time.Second)
			continue
		}
		job := jobs[0]
		req := dto.JobRequest{
			ID:      strconv.Itoa(job.ID),
			Content: nil,
		}
		resp, err := s.workerClient.Ping(req)
		if err != nil {
			s.logger.Println("occurred error during ping task", err)
			<-time.After(2 * time.Second)
			continue
		}
		if resp.Status == "completed" {
			s.logger.Println(fmt.Sprintf("task %v completed, saving results", job.ID))
			tx, err := s.db.Begin()
			if err != nil {
				s.logger.Println("occurred error during begin transaction", err)
				<-time.After(10 * time.Second)
				continue
			}
			err = s.jobService.CompleteTaskInTx(
				job.ID,
				resp.Status,
				resp.Error,
				resp.Result,
				tx,
			)
			if err != nil {
				s.logger.Println("occurred error during complete task", err)
				_ = tx.Rollback()
				<-time.After(2 * time.Second)
				continue
			}
			_ = tx.Commit()
		}
		<-time.After(2 * time.Second)
	}
}
