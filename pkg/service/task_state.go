package service

type TaskState int

const (
	TaskInWIP    TaskState = iota // 0
	TaskComplete                  // 1
)
