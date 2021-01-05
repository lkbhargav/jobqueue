package jobqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

type asyncJob struct {
	name    string
	timeout int
}

// Run => implementing the jobqueue interface method
func (aJob asyncJob) Run(ctx context.Context, jr JobRelated, args ...interface{}) {
	defer ctx.Done()
	fmt.Println(args[0], args[1])
	jr.Update(args[0].(string))
	time.Sleep(time.Duration(aJob.timeout) * time.Second) // for simulating some intense work
}

func printJobList(jobsList []JobInfo) {
	js, _ := json.Marshal(jobsList)
	fmt.Println(string(js))
}

func TestJobsList(t *testing.T) {
	var aJob asyncJob

	aJob.name = "Test job"
	aJob.timeout = 5

	jqueue := Init()
	defer jqueue.EndJobQueue()

	jqueue.Add(JobRelatedData{Description: "Test 1", Priority: Low, JobStatus: aJob, Params: []interface{}{"Yudhistira", "anything else 1"}})
	jqueue.Add(JobRelatedData{Description: "Test 2", Priority: High, JobStatus: aJob, Params: []interface{}{"Arjuna", "anything else 2"}})
	jqueue.Add(JobRelatedData{Description: "Test 3", Priority: Medium, JobStatus: aJob, Params: []interface{}{"Bhima", "anything else 3"}})
	jqueue.Add(JobRelatedData{Description: "Test 4", Priority: Medium, JobStatus: aJob, Params: []interface{}{"Sahadeva", "anything else 4"}})
	jqueue.Add(JobRelatedData{Description: "Test 5", Priority: High, JobStatus: aJob, Params: []interface{}{"Nakula", "anything else 5"}})

	jobsList := jqueue.GetListOfJobs()

	var jobID string

	for _, job := range jobsList {
		if job.Description == "Test 1" {
			jobID = job.JobID
		}
	}

	resp := jqueue.Kill(jobID)

	if !resp {
		t.Errorf("Expected the value to be true instead found it to %v", resp)
	}

	time.Sleep(35 * time.Second)

	jobs := jqueue.GetListOfJobs()

	if len(jobsList) != 5 {
		t.Errorf("Expected to see just one job, but instead found %v jobs", len(jobsList))
	}

	for _, j := range jobs {
		if j.JobID == jobID {
			if j.Status != Killed {
				t.Errorf("Expected the job to be killed but instead found it to be %v", j.Status)
			}
		}
	}
}

func TestForKillingARunningJob(t *testing.T) {
	var aJob asyncJob

	aJob.name = "Test job 2"
	aJob.timeout = 20

	jqueue := Init()
	defer jqueue.EndJobQueue()

	jqueue.Add(JobRelatedData{Description: "Test 1", Priority: Low, JobStatus: aJob, Params: []interface{}{"Yudhistira", "anything else 1"}})

	time.Sleep(2 * time.Second)

	resp := jqueue.Kill(jqueue.GetListOfJobs()[0].JobID)

	if !resp {
		t.Errorf("Expected the value to be true instead found it to %v", resp)
	}
}

func TestForKillingAInvalidJob(t *testing.T) {
	var aJob asyncJob

	aJob.name = "Test job 3"
	aJob.timeout = 20

	jqueue := Init()
	defer jqueue.EndJobQueue()

	jqueue.Add(JobRelatedData{Description: "Test 1", Priority: Low, JobStatus: aJob, Params: []interface{}{"Yudhistira", "anything else 1"}})

	time.Sleep(2 * time.Second)

	resp := jqueue.Kill("abcd")

	if resp {
		t.Errorf("Expected the value to be false instead found it to %v", resp)
	}
}

func TestForAfterKillingJobQueue(t *testing.T) {
	var aJob asyncJob

	aJob.name = "Test job 3"
	aJob.timeout = 20

	jqueue := Init()

	jqueue.Add(JobRelatedData{Description: "Test 1", Priority: Low, JobStatus: aJob, Params: []interface{}{"Yudhistira", "anything else 1"}})
	jqueue.Add(JobRelatedData{Description: "Test 2", Priority: High, JobStatus: aJob, Params: []interface{}{"Arjuna", "anything else 2"}})
	jqueue.Add(JobRelatedData{Description: "Test 3", Priority: Medium, JobStatus: aJob, Params: []interface{}{"Bhima", "anything else 3"}})
	jqueue.Add(JobRelatedData{Description: "Test 4", Priority: Medium, JobStatus: aJob, Params: []interface{}{"Sahadeva", "anything else 4"}})
	jqueue.Add(JobRelatedData{Description: "Test 5", Priority: High, JobStatus: aJob, Params: []interface{}{"Nakula", "anything else 5"}})

	jqueue.EndJobQueue()

	for _, job := range jobs {
		if job.status != Killed || !job.completed {
			t.Errorf("Expected the job to be killed, but found otherwise (status: %v, completed: %v)", job.status, job.completed)
		}
	}
}
