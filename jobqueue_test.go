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
func (aJob asyncJob) Run(jr JobRelated, args ...interface{}) (resp bool) {
	fmt.Println(aJob.name, len(args))
	time.Sleep(5 * time.Second) // for simulating some intense work
	return
}

func TestJobsList(t *testing.T) {
	var aJob asyncJob

	aJob.name = "Bhargav"

	jqueue := Init()
	defer jqueue.Kill()

	jqueue.Add(JobRelatedData{Description: "Test 1", Priority: 2, JobStatus: aJob, Params: []interface{}{"name", "anything else"}})

	josbList := jqueue.GetListOfJobs()

	if len(josbList) != 1 {
		t.Errorf("Expected to see just one job, but instead found %v jobs", len(josbList))
	}
}
