package scheduler

import (
	"backend/internal/clients"
	"backend/internal/dto"
	"backend/internal/models"
	"backend/internal/service"
	"backend/pkg/graph"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

type Scheduler struct {
	taskService   *service.TaskService
	jobService    *service.JobService
	logger        *log.Logger
	jobCreated    chan struct{}
	stop          chan struct{}
	db            *sql.DB
	workerClient  *clients.WorkerClient
	taskScheduled chan struct{}
}

func NewScheduler(
	taskService *service.TaskService,
	mtxService *service.JobService,
	logger *log.Logger,
	jobScheduled chan struct{},
	db *sql.DB,
	client *clients.WorkerClient,
) *Scheduler {
	s := &Scheduler{
		taskService:  taskService,
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
	go s.taskProcessor()
	go s.pollTaskStatus()
}

func (s *Scheduler) Stop() {
	close(s.jobCreated)
}

func (s *Scheduler) taskCreator() {
	for {
		select {
		case <-s.jobCreated:
			err := s.createTasks()
			if err != nil {
				s.logger.Println("occurred error during creating task", err)
			}
		case <-time.After(5 * time.Second):
			err := s.createTasks()
			if err != nil {
				s.logger.Println("occurred error during creating task", err)
			}
		case <-s.stop:
			break
		}
	}
}

func (s *Scheduler) createTasks() error {
	status := "created"
	notScheduled, err := s.jobService.ListJobs(&status)
	if err != nil {
		return err
	}
	if len(notScheduled) == 0 {
		return errors.New("no scheduled jobs found")
	}
	parts := strings.Split(notScheduled[0].Dimensions, "x")
	id := notScheduled[0].ID
	nodesCount, err := strconv.Atoi(parts[0])
	if err != nil {
		return err
	}
	g := graph.New(nodesCount)
	job, err := s.jobService.GetJob(id)
	if err != nil {
		return err
	}
	buff := bytes.NewBufferString(job.Content)
	for {
		line, err := buff.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if strings.Contains(line, "%") {
			continue
		}
		edge := strings.Split(line, " ")
		if len(edge) < 2 {
			return errors.New("invalid format")
		}
		edge[0] = strings.Trim(edge[0], "\n")
		edge[1] = strings.Trim(edge[1], "\n")
		from, err := strconv.Atoi(edge[0])
		if err != nil {
			return err
		}
		to, err := strconv.Atoi(edge[1])
		if err != nil {
			return err
		}
		if from <= 0 || to <= 0 {
			return errors.New("invalid format")
		}
		g.AddEdge(from, to)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	for {
		cmp := g.GetNextComponent()
		if cmp == nil {
			break
		}
		if len(cmp.Nodes) < 2 || len(cmp.Edges) < 1 {
			s.logger.Println("continue loop due to empty component list")
			continue
		}
		mapping := make(map[int]int)
		for i, node := range cmp.Nodes {
			mapping[node] = i
		}
		nodesToSave := make([]int, len(cmp.Nodes))
		edgesToSave := make([]int, 0, len(cmp.Edges))
		for _, node := range cmp.Nodes {
			nodesToSave[mapping[node]] = node
		}
		for _, node := range cmp.Edges {
			edgesToSave = append(edgesToSave, mapping[node]+1)
		}
		err = s.taskService.CreateTaskInTx(len(cmp.Nodes), nodesToSave, id, edgesToSave, tx)
		if err != nil {
			break
		}
		<-time.After(100 * time.Millisecond)
	}
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not create tasks: %w", err)
	}
	err = s.jobService.ScheduleJobInTx(id, tx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not schedule job: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not commit transaction: %w", err)
	}
	return nil
}

func (s *Scheduler) taskProcessor() {
	for {
		task, err := s.taskService.GetTaskByStatus("created")
		if err != nil {
			s.logger.Println("occurred error during getting task", err)
			// here will be great if we track number of attempts(not only here)
			// if attempts_count > max_number_of_attempts just save error and go to next job
			<-time.After(5 * time.Second)
			continue
		}
		if task == nil {
			s.logger.Println("no created task continue task processing")
			<-time.After(5 * time.Second)
			continue
		}
		err = s.handleCreatedTask(task)
		if err != nil {
			s.logger.Println("occurred error during handling task", err)
			<-time.After(5 * time.Second)
			continue
		}
		<-time.After(5 * time.Second)
	}
}

func (s *Scheduler) handleCreatedTask(task *models.Task) error {
	// идем в балансер до воркеров считаем что там есть балансер, rps limiter и тд
	s.logger.Println("handling created task", task.ID)
	req := dto.TaskRequest{
		ID:    strconv.Itoa(task.ID),
		Edges: task.EdgesArray,
	}
	_, err := s.workerClient.Ping(req)
	if err != nil {
		s.logger.Println("occurred error during ping task", err)
		return err
	}
	s.logger.Println("updating task status to executing task", task.ID)
	return s.taskService.UpdateTaskStatus(task.ID, "executing")
}

func (s *Scheduler) getCreatedTask() (*models.Task, error) {
	task, err := s.taskService.GetTaskByStatus("created")
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Scheduler) pollTaskStatus() {
	for {
		task, err := s.taskService.GetTaskByStatus("executing")
		if err != nil {
			s.logger.Println("occurred error during getting task", err)
			<-time.After(2 * time.Second)
			continue
		}
		if task == nil {
			s.logger.Println("no running task found")
			<-time.After(10 * time.Second)
			continue
		}
		req := dto.TaskRequest{
			ID:    strconv.Itoa(task.ID),
			Edges: nil,
		}
		resp, err := s.workerClient.Ping(req)
		if err != nil {
			s.logger.Println("occurred error during ping task", err)
			<-time.After(2 * time.Second)
			continue
		}
		if resp.Status == "completed" {
			s.logger.Println(fmt.Sprintf("task %v completed, saving results", task.ID))
			tx, err := s.db.Begin()
			if err != nil {
				s.logger.Println("occurred error during begin transaction", err)
				<-time.After(10 * time.Second)
				continue
			}
			err = s.taskService.CompleteTaskInTx(
				task.ID,
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
