package scheduler

import (
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
	taskService       *service.TaskService
	jobService        *service.JobService
	logger            *log.Logger
	jobCreated        chan struct{}
	stop              chan struct{}
	taskProcessorDone chan struct{}
	db                *sql.DB
}

func NewScheduler(
	taskService *service.TaskService,
	mtxService *service.JobService,
	logger *log.Logger,
	jobScheduled chan struct{},
	db *sql.DB,
) *Scheduler {
	s := &Scheduler{
		taskService:       taskService,
		jobService:        mtxService,
		logger:            logger,
		jobCreated:        jobScheduled,
		db:                db,
		taskProcessorDone: make(chan struct{}),
		stop:              make(chan struct{}),
	}
	return s
}

func (s *Scheduler) Start() {
	go s.taskCreator()
}

func (s *Scheduler) Stop() {
	close(s.jobCreated)
	<-s.taskProcessorDone
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
	s.taskProcessorDone <- struct{}{}
}

func (s *Scheduler) createTasks() error {
	notScheduled, err := s.jobService.ListJobs(true)
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
		fmt.Println(edge[0], edge[1], from, to, "###")
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
		//if len(cmp.Nodes) < 2 || len(cmp.Edges) < 1 {
		//	s.logger.Println("continue loop due to empty component list")
		//	continue
		//}
		fmt.Println("component nodes is", cmp.Nodes, nodesCount)
		fmt.Println("component edges is", cmp.Edges)
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
		err = s.taskService.CreateTaskInTx(nodesCount, nodesToSave, id, edgesToSave, tx)
		if err != nil {
			break
		}
		<-time.After(100 * time.Millisecond)
	}
	fmt.Println("leave component processing")
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
