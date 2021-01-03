package jobqueue

import (
	"fmt"
	"testing"
	"time"
)

type asyncJob struct {
	name string
}

// Run => implementing the jobqueue interface method
func (aJob asyncJob) Run(jr JobRelated, args ...interface{}) {
	fmt.Println(aJob.name, args[1])
	jr.Update("Printed a message")
	time.Sleep(5 * time.Second) // for simulating some intense work
}

func TestJobsList(t *testing.T) {
	var aJob asyncJob

	aJob.name = "Bhargav"

	jqueue := Init()
	defer jqueue.Kill()

	jqueue.Add(JobRelatedData{Description: "Test 1", Priority: 3, JobStatus: aJob, Params: []interface{}{"name", "anything else 1"}})
	jqueue.Add(JobRelatedData{Description: "Test 2", Priority: 1, JobStatus: aJob, Params: []interface{}{"name", "anything else 2"}})
	jqueue.Add(JobRelatedData{Description: "Test 3", Priority: 2, JobStatus: aJob, Params: []interface{}{"name", "anything else 3"}})
	jqueue.Add(JobRelatedData{Description: "Test 4", Priority: 2, JobStatus: aJob, Params: []interface{}{"name", "anything else 4"}})

	josbList := jqueue.GetListOfJobs()

	time.Sleep(30 * time.Second)

	if len(josbList) != 4 {
		t.Errorf("Expected to see just one job, but instead found %v jobs", len(josbList))
	}
}
