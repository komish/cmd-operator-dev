package types

import (
	"time"
)

// LoggingFlags appear to be coming from klog. These should be universal for
// all controllers.
type loggingFlags struct {
	AddDirectoryHeaders bool          `json:"add_dir_header"`
	AlsoLogToSTDERR     bool          `json:"alsologtostderr"`
	LogFlushFrequency   time.Duration `json:"log-flush-frequency"`
	LogBacktraceAt      string        `json:"log_backtrace_at"`
	LogDir              string        `json:"log_dir"`
	LogFile             string        `json:"log_file"`
	LogFileMaxSize      uint          `json:"log_file_max_size"`
	LogToSTDERR         bool          `json:"logtostderr"`
	SkipHeaders         bool          `json:"skip_headers"`
	SkipLogHeaders      bool          `json:"skip_log_headers"`
	STDERRThreshold     int32         `json:"stderrthreshold"` // similar type to klog unexported severity type
	VerbosityLevel      int32         `json:"v"`               // similar type to klog exported Level type
	VModule             string        `json:"vmodule"`         // accepting string because original type ModuleSpec is unexported from klog
}
