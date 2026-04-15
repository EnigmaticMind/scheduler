package appointments

type errorConst string

func (e errorConst) Error() string {
	return string(e)
}

const (
	ErrInvalidAptDur            = errorConst("appointments must be 30 minutes long")
	ErrInvalidAptStartEnd       = errorConst("appointsments must be every half hour")
	ErrInvalidAptBusinessHours  = errorConst("appointments must be during business hours M-F 8am-5pm Pacific Time")
	ErrInvalidAptOverlap        = errorConst("appointment time not available")
	ErrInvalidAptEndBeforeStart = errorConst("appointment time start before end")
)
