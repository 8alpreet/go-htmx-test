package contact

import (
	"fmt"
	"time"
)

var (
	archiveStatus   = "Waiting"
	archiveProgress = float64(0)
)

type Archiver struct {
	// thread        *thread
}

// type thread struct {
//     target func()
// }

// func (t *thread) start() {
//     go t.target()
// }

func NewArchiver() *Archiver {
	return &Archiver{}
}

func (a *Archiver) Status() string {
	return archiveStatus
}

func (a *Archiver) Progress() float64 {
	return archiveProgress
}

func (a *Archiver) Run() {
	if archiveStatus == "Waiting" {
		archiveStatus = "Running"
		archiveProgress = 0
		go a.runImpl()
		// a.thread = &thread{
		//     target: a.runImpl,
		// }
		// a.thread.start()
	}
}

func (a *Archiver) runImpl() {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		// check in case reset was called?
		if archiveStatus != "Running" {
			return
		}
		archiveProgress = float64(i+1) / 10
		fmt.Println("Here...", archiveProgress)
	}
	time.Sleep(time.Second)
	// check in case reset was called?
	if archiveStatus != "Running" {
		return
	}
	archiveStatus = "Complete"
}

func (a *Archiver) ArchiveFile() string {
	return "contacts.json"
}

func (a *Archiver) Reset() {
	archiveStatus = "Waiting"
}
