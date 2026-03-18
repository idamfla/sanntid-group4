package coordinator

import (
	"context"
	"elevator_program/elevator"
	"elevator_program/elevio"
	"elevator_program/message"
	"elevator_program/types"
	"fmt"
	"sync"
	"time"
)

type TaskKey struct {
	Owner  string
	TaskID elevio.ButtonEvent
}

type TaskMonitor struct {
	timeout time.Duration
	tasks   map[TaskKey]context.CancelFunc
	mu      sync.Mutex
}

func NewTaskMonitor(timeout time.Duration) TaskMonitor {
	return TaskMonitor{
		timeout: timeout,
		tasks:   make(map[TaskKey]context.CancelFunc),
	}
}

func (tm *TaskMonitor) StartTask(taskKey TaskKey, e *elevator.Elevator) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Cancel existing timer if it exists
	if cancel, ok := tm.tasks[taskKey]; ok {
		cancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), tm.timeout)
	tm.tasks[taskKey] = cancel

	go func() {
		<-ctx.Done()

		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("Task %+v timed out! Trigger fault tolerance.\n", taskKey)
			// TODO send fault to fault tolerance

			msg := message.ElevatorMessage{
				MsgType:   types.MSG_T_TaskUpdate,
				Id:        taskKey.Owner,
				Task:      taskKey.TaskID,
				BtnStatus: types.Pending,
			}
			e.SendToProtocol <- msg
		}

		tm.mu.Lock()
		delete(tm.tasks, taskKey)
		tm.mu.Unlock()
	}()
}

func (tm *TaskMonitor) FinishTask(taskKey TaskKey) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if cancel, ok := tm.tasks[taskKey]; ok {
		cancel()
		delete(tm.tasks, taskKey)
		fmt.Printf("Task %+v completed, timer cleared.\n", taskKey)
	}
}
