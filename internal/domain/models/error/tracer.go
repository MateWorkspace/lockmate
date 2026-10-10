package domainmodelserror

import "errors"

var (
	TraceNotFound      = errors.New("TRACE_NOT_FOUND")
	TraceEventIdFailed = errors.New("TRACE_EVENT_ID_FAILED")
)
