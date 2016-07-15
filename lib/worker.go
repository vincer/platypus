package lib

import (
	"fmt"
)


// http://nesv.github.io/golang/2014/02/25/worker-queues-in-go.html

type UpdateRequest struct {
	Id     string
	Height int
}

var UpdateQueue = make(chan UpdateRequest, 100)

func NewWorker(id int, workerQueue chan chan UpdateRequest) Worker {
	worker := Worker{
		ID:          id,
		Work:        make(chan UpdateRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool)}

	return worker
}

type Worker struct {
	ID          int
	Work        chan UpdateRequest
	WorkerQueue chan chan UpdateRequest
	QuitChan    chan bool
}

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *Worker) Start() {
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
			// Receive a work request.
				id := work.Id
				shade, err := FindShade(id)
				if (err != nil) {
					Log.Warning("Shade {} not found.", id);
					//c.JSON(http.StatusNotFound, lib.Response{Code: http.StatusNotFound, Message: "Not found"})
				} else {
					shade.SetHeight(work.Height)
				}

			case <-w.QuitChan:
			// We have been asked to stop.
				fmt.Printf("worker%d stopping\n", w.ID)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
//
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}

var WorkerQueue chan chan UpdateRequest

func StartDispatcher(nworkers int) {
  // First, initialize the channel we are going to but the workers' work channels into.
  WorkerQueue = make(chan chan UpdateRequest, nworkers)

  // Now, create all of our workers.
  for i := 0; i<nworkers; i++ {
    fmt.Println("Starting worker", i+1)
    worker := NewWorker(i+1, WorkerQueue)
    worker.Start()
  }

  go func() {
    for {
      select {
      case work := <-UpdateQueue:
        fmt.Println("Received work requeust")
        go func() {
          worker := <-WorkerQueue

          fmt.Println("Dispatching work request")
          worker <- work
        }()
      }
    }
  }()
}
