package jobqueue

import (
	"math/rand"
	"time"
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
	Priority         int           // three levels available | 1 => High; 2 => Medium; 3 => Low;
	JobStatus        JobStatus     // struct implementing the interface that is core
	Params           []interface{} // optional set of parameters
	createdTime      time.Time
	message          string
	status           string
	statusChangeTime time.Time
}

// JobStatus => provides an interface for apps to implement and make it compatible with jobqueue
type JobStatus interface {
	Run(jr JobRelated, args ...interface{})
}

// JobInfo => used to expose data to the user
type JobInfo struct {
	JobID       string
	Description string
	Message     string
	CreatedTime time.Time
	Status      string
	UpdatedTime time.Time
	Priority    int
}

const characters string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var jobs map[string]JobRelatedData
var jobInProgress bool

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
							if v.Priority == priority {
								go func() {
									v.status = "Running"
									v.statusChangeTime = time.Now()
									jobs[jobID] = v // updating the jobs map with updated values

									compute(v.JobStatus, JobRelated{jobID: jobID}, v.Params...) // running the job in question as requested by the user

									delete(jobs, jobID) // remove the job that just got completed executing

									jobInProgress = false // helps open gates for new jobs
								}()
								v.createdTime = time.Now()
								jobInProgress = true
								jobs[jobID] = v
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

func compute(js JobStatus, jr JobRelated, args ...interface{}) { // The method that implements our interfaces | HERO
	js.Run(jr, args...)
}

// Add => helps add a job
func (m Master) Add(job JobRelatedData) JobRelated {
	jobID := randomPrefix(10)

	job.status = "In queue"
	job.createdTime = time.Now()
	job.statusChangeTime = time.Now()

	m.jobs[jobID] = job

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

// Kill => acts like a kill switch to put down the jobs queue service (mostly the go routine that runs in the background)
func (m Master) Kill() {
	m.done <- true
}

// Update => lets us update the status of the job periodically
func (jr JobRelated) Update(message string) {
	job := jobs[jr.jobID]
	job.message = message
	jobs[jr.jobID] = job
}

func randomPrefix(n int) string { // generates random job ids on the fly
	b := make([]byte, n)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}
