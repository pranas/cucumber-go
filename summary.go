package cucumber

import (
	"time"
)

type Summary struct {
	Success  bool
	ExitCode int
	Duration time.Duration

	TestCasesTotal     int
	TestCasesPassed    int
	TestCasesFailed    int
	TestCasesPending   int
	TestCasesUndefined int

	StepsTotal     int
	StepsPassed    int
	StepsFailed    int
	StepsPending   int
	StepsUndefined int
	StepsSkipped   int
}
