package domainmodelslogger

type Level string

const (
	LevelNone  Level = "NONE"
	LevelError Level = "ERROR"
	LevelWarn  Level = "WARN"
	LevelInfo  Level = "INFO"
	LevelDebug Level = "DEBUG"
)

func (l Level) String() string {
	return string(l)
}

func (l Level) IsValid() bool {
	switch l {
	case LevelNone, LevelError, LevelWarn, LevelInfo, LevelDebug:
		return true
	default:
		return false
	}
}

func (l Level) Order() int {
	switch l {
	case LevelNone:
		return 0
	case LevelError:
		return 1
	case LevelWarn:
		return 2
	case LevelInfo:
		return 3
	case LevelDebug:
		return 4
	default:
		return -1
	}
}
