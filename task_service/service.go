package taskservice

type Service interface {
	Start(string) error
	Stop(string) error
}

type ServiceType int

const (
	Systemctl ServiceType = iota
	NNSM
	TaskSched
)

func ServiceTypeFrom(t string) ServiceType {
	switch t {
	case "nssm":
		return NNSM
	case "tasksched":
		return TaskSched
	default:
		return TaskSched
	}
}
