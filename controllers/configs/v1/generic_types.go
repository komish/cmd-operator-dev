package v1

import (
	"fmt"
	"time"

	"k8s.io/klog"
)

// loggingFlags appear to be coming from klog. These should be universal for
// all controllers.
type loggingFlags struct {
	AddDirectoryHeaders bool          `json:"add_dir_header"`
	AlsoLogToSTDERR     bool          `json:"alsologtostderr"`
	LogFlushFrequency   time.Duration `json:"log-flush-frequency"`
	LogBacktraceAt      TraceLocation `json:"log_backtrace_at"`
	LogDir              string        `json:"log_dir"`
	LogFile             string        `json:"log_file"`
	LogFileMaxSize      uint          `json:"log_file_max_size"`
	LogToSTDERR         bool          `json:"logtostderr"`
	SkipHeaders         bool          `json:"skip_headers"`
	SkipLogHeaders      bool          `json:"skip_log_headers"`
	STDERRThreshold     severity      `json:"stderrthreshold"`
	VerbosityLevel      klog.Level    `json:"v"`
	VModule             string        `json:"vmodule"` // accepting string because original type ModuleSpec is unexported from klog
}

// traceLocation is an emulation of the unexported struct traceLocation in klog https://github.com/kubernetes/klog/blob/master/klog.go#L325
type TraceLocation struct {
	file string
	line int
}

func (t *TraceLocation) String() string {
	return fmt.Sprintf("%s:%v", t.file, t.line)
}

func (t *TraceLocation) IsSet() bool {
	return t.line > 0
}

// severity is an emulation of the unexported struct severity in klog https://github.com/kubernetes/klog/blob/master/klog.go#L98
type severity int32
