package jobqueue

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// Priority => holds the priority types
type Priority int

const (
	// High => absolutely important
	High Priority = 1
	// Medium => Kind of important
	Medium Priority = 2
	// Low => Not that important
	Low Priority = 3
)

// Status => holds the status types
type Status string

const (
	// Running => signifies a job that is running
	Running Status = "Running"
	// Completed => signifies job completeion
	Completed Status = "Completed!"
	// Killed => signifies job is killed
	Killed Status = "Killed!"
	// InQueue => signifies job is in queue
	InQueue Status = "In queue"
)

// Master => master instance that gets returned on Job queue initialization
type Master struct { // helps bridging the gap between the outer calling functions and this inner world of managing job queues
	jobs map[string]JobRelatedData
	done chan bool
}

// JobRelated => used with exposed "Update" function for easy access to update the status
type JobRelated struct {
	jobID string // randomly generated ID for each job
}

// JobRelatedData => contract that the caller has to provide details about the job | also dubs as status manager / for job log recording
type JobRelatedData struct {
	Description      string        // for our own reference
	Priority         Priority      // three levels available | 1 => High; 2 => Medium; 3 => Low;
	JobStatus        JobStatus     // struct implementing the interface that is core
	Params           []interface{} // optional set of parameters
	createdTime      time.Time
	message          string
	status           Status
	statusChangeTime time.Time
	channel          chan bool
	completed        bool
}

// JobStatus => provides an interface for apps to implement and make it compatible with jobqueue
type JobStatus interface {
	Run(ctx context.Context, jr JobRelated, args ...interface{})
}

// JobInfo => used to expose data to the user
type JobInfo struct {
	JobID       string
	Description string
	Message     string
	CreatedTime time.Time
	Status      Status
	UpdatedTime time.Time
	Priority    Priority
}

const characters string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var jobs map[string]JobRelatedData
var jobInProgress bool
var jobsMutex sync.Mutex

// Init => helps initialize
func Init() Master {
	jobs = make(map[string]JobRelatedData)

	ticker := time.NewTicker(1 * time.Second) // ticker that runs every couple seconds triggering the loop in the goroutine
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if !jobInProgress { // gate keeper that makes sure only one job runs at any time
				outerLoop:
					for priority := 1; priority <= 3; priority++ { // helps prioritize between tasks in hand based on priority
						for jobID, v := range jobs { // loops through jobs until a fit is found as in the condition below
							if v.Priority == Priority(priority) && !v.completed {
								go func() {
									v.status = Running
									v.statusChangeTime = time.Now()
									v.channel = make(chan bool, 1)

									jobsMutex.Lock()
									jobs[jobID] = v // updating the jobs map with updated values
									jobsMutex.Unlock()

									compute(v.JobStatus, JobRelated{jobID: jobID}, v) // running the job in question as requested by the user

									v.status = Completed
									v.statusChangeTime = time.Now()
									v.completed = true

									jobsMutex.Lock()
									jobs[jobID] = v // updating the jobs map with updated values
									jobsMutex.Unlock()

									go func() { // we will remove the job from the list only after 24 hours
										time.Sleep(24 * time.Hour)
										delete(jobs, jobID) // remove the job that just got completed executing
									}()

									jobInProgress = false // helps open gates for new jobs
								}()

								v.createdTime = time.Now()
								jobInProgress = true

								jobsMutex.Lock()
								jobs[jobID] = v
								jobsMutex.Unlock()

								break outerLoop // breaks outerloop once it finds a job to assign to
							}
						}
					}
				}
			}
		}
	}()

	return Master{jobs: jobs, done: done}
}

func compute(js JobStatus, jr JobRelated, job JobRelatedData) { // The method that implements our interfaces | HERO
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		js.Run(ctx, jr, job.Params...)
		job.channel <- true
	}()

	select {
	case val := <-job.channel:
		if val {
			cancel()
			return
		}
	}

	cancel() // to avoid context leaks
}

// Add => helps add a job
func (m Master) Add(job JobRelatedData) JobRelated {
	jobID := randomPrefix(10)

	job.status = InQueue
	job.createdTime = time.Now()
	job.statusChangeTime = time.Now()

	jobsMutex.Lock()
	m.jobs[jobID] = job
	jobsMutex.Unlock()

	return JobRelated{jobID: jobID}
}

// Block => enables to wait for indefinite time
func (m Master) Block() {
	var ch chan bool

	<-ch
}

// GetListOfJobs => gets us a loop of jobs with relevant info for quick reference
func (m Master) GetListOfJobs() (response []JobInfo) {
	for jobID, info := range m.jobs {
		response = append(response, JobInfo{
			JobID:       jobID,
			Description: info.Description,
			Message:     info.message,
			CreatedTime: info.createdTime,
			Status:      info.status,
			UpdatedTime: info.statusChangeTime,
			Priority:    info.Priority,
		})
	}

	return
}

// EndJobQueue => acts like a kill switch to put down the jobs queue service (mostly the go routine that runs in the background)
func (m Master) EndJobQueue() {
	m.done <- true

	for jobID, job := range m.jobs {
		if job.status == Running {
			job.channel <- true
		}
		job.completed = true
		job.status = Killed

		jobsMutex.Lock()
		jobs[jobID] = job
		jobsMutex.Unlock()
	}

	go func() {
		time.Sleep(10 * time.Second)
		m.jobs = map[string]JobRelatedData{}
	}()
}

// Update => lets us update the status of the job periodically
func (jr JobRelated) Update(message string) {
	job := jobs[jr.jobID]
	job.message = message

	jobsMutex.Lock()
	jobs[jr.jobID] = job
	jobsMutex.Unlock()
}

// Kill => kills a job
func (m Master) Kill(jobID string) (resp bool) {
	job := m.jobs[jobID]

	if job.Priority == 0 { // filtering invalid jobID's
		return
	}

	if !job.completed && job.status == Running {
		job.channel <- true
		job.status = Killed
		job.statusChangeTime = time.Now()
		resp = true
	} else if !job.completed && job.status != Running {
		job.completed = true
		job.status = Killed
		job.statusChangeTime = time.Now()
		resp = true
	}

	jobsMutex.Lock()
	m.jobs[jobID] = job
	jobsMutex.Unlock()

	return
}

func randomPrefix(n int) string { // generates random job ids on the fly
	b := make([]byte, n)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}
