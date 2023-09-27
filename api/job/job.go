package job

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// A Job is a task that has a current state and will finish (with complete or not) someday
type Job struct {
	// Job's task, the action the task does, like "Insert data into table"
	Task string
	// The state of the job can be:
	// Not started
	// Starting
	// Executing
	// Completed
	// Failed
	State string
	// A description representing the actual job state, like "Getting data", "Page created with success", or an error that occuried "couldn't get the data".
	StateDescription string
	// A optional value to return to the user anytime.
	// This value can represent the current data the job is processing
	Value string
	// When the job was created. Should have format "2006-01-02 15:04:05"
	CreatedAt string
	// When the job was completed with success or failed.
	// It will be set by the SetCompletedState or the SetFailedState functions.
	// Should have format "2006-01-02 15:04:05"
	Completed_Failed_At string
	id                  uuid.UUID
	mutex               sync.Mutex
}

// The Set*State functions set the current state of a Job
func (job *Job) SetNotStartedState() {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.State = "Not started"
}

func (job *Job) SetStartingStateWithValue(stateMessage, value string) {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.State = "Starting"
	job.StateDescription = stateMessage
	job.Value = value
}

func (job *Job) SetStartingState(stateMessage string) {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.State = "Starting"
	job.StateDescription = stateMessage
}

// Set a staus of executing to the job
//
// Parameters:
// args - The arguments that will be used by the task while executing before it finishes or fail
func (job *Job) SetExecutingStateWithValue(stateMessage, value string) {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.State = "Executing"
	job.StateDescription = stateMessage
	job.Value = value
}

func (job *Job) SetExecutingState(stateMessage string) {
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.State = "Executing"
	job.StateDescription = stateMessage
}

// Set a state of finished with complete to the job
//
// Parameters:
// value - The returning value of the process task
func (job *Job) SetCompletedStateWithValue(stateMessage, value string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.Completed_Failed_At = now
	job.State = "Completed"
	job.StateDescription = stateMessage
	job.Value = value
}

func (job *Job) SetCompletedState(stateMessage string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.Completed_Failed_At = now
	job.StateDescription = stateMessage
	job.State = "Completed"
}

// Set a state of failed to the job
//
// Parameters:
// err - The returning error of the process task. This value will be used to set the job StateDescription
func (job *Job) SetFailedState(err error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	job.mutex.Lock()
	defer job.mutex.Unlock()

	job.Completed_Failed_At = now
	job.State = "Failed"
	job.StateDescription = err.Error()
}

func NewJobsList() *Jobs {
	jobs := []*Job{}
	return &Jobs{
		Jobs:  jobs,
		mutex: sync.Mutex{},
	}
}

// A collection of jobs (active or not)
type Jobs struct {
	Jobs  []*Job
	mutex sync.Mutex
}

func (jobs *Jobs) AddJob(job *Job) {
	jobs.mutex.Lock()
	defer jobs.mutex.Unlock()

	jobUUID := uuid.New()
	job.id = jobUUID
	jobs.Jobs = append(jobs.Jobs, job)
}

func (jobs *Jobs) GetJobs() []*Job {
	jobs.mutex.Lock()
	defer jobs.mutex.Unlock()

	return jobs.Jobs
}

func (jobs *Jobs) DeleteAllJobs() {
	jobs.mutex.Lock()
	defer jobs.mutex.Unlock()

	jobs.Jobs = []*Job{}
}
