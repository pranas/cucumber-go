package cucumber

// Based on https://github.com/cucumber/dots-formatter-go/blob/master/dots.go

import (
	"fmt"
	"io"
	"strings"

	messages "github.com/cucumber/cucumber-messages-go/v3"
	"github.com/fatih/color"
)

const (
	successColor   = color.FgGreen
	failureColor   = color.FgRed
	skippedColor   = color.FgCyan
	undefinedColor = color.FgYellow
	pendingColor   = color.FgYellow
	ambiguousColor = color.FgMagenta
)

type dotFormatter struct {
	out io.Writer
}

func NewDotFormatter(stdout io.Writer) *dotFormatter {
	return &dotFormatter{
		out: stdout,
	}
}

func (df *dotFormatter) ProcessMessage(msg *messages.Envelope) {
	switch m := msg.Message.(type) {
	case *messages.Envelope_TestRunFinished:
		fmt.Fprint(df.out, "\n")
	case *messages.Envelope_TestHookFinished:
		switch m.TestHookFinished.TestResult.Status {
		case messages.TestResult_FAILED:
			color.New(failureColor).Fprint(df.out, "H")
		}
	case *messages.Envelope_TestStepFinished:
		switch m.TestStepFinished.TestResult.Status {
		case messages.TestResult_AMBIGUOUS:
			color.New(ambiguousColor).Fprint(df.out, "A")
		case messages.TestResult_FAILED:
			color.New(failureColor).Fprint(df.out, "F")
		case messages.TestResult_PASSED:
			color.New(successColor).Fprint(df.out, ".")
		case messages.TestResult_PENDING:
			color.New(pendingColor).Fprint(df.out, "P")
		case messages.TestResult_SKIPPED:
			color.New(skippedColor).Fprint(df.out, "-")
		case messages.TestResult_UNDEFINED:
			color.New(undefinedColor).Fprint(df.out, "U")
		}
	}
}

func (df *dotFormatter) DisplaySummary(s Summary) {
	fmt.Fprint(df.out, "\n")
	scenarioStatusSummary := statusSummary(s.TestCasesPassed, s.TestCasesFailed, s.TestCasesPending, s.TestCasesUndefined, 0)
	fmt.Fprintf(df.out, "%d scenarios (%s)\n", s.TestCasesTotal, scenarioStatusSummary)

	stepStatusSummary := statusSummary(s.StepsPassed, s.StepsFailed, s.StepsPending, s.StepsUndefined, s.StepsSkipped)
	fmt.Fprintf(df.out, "%d steps (%s)\n", s.StepsTotal, stepStatusSummary)
	fmt.Fprintln(df.out, s.Duration)
}

func statusSummary(passed, failed, pending, undefined, skipped int) string {
	var acc []string

	if passed > 0 {
		acc = append(acc, color.New(successColor).Sprintf("%d passed", passed))
	}

	if failed > 0 {
		acc = append(acc, color.New(failureColor).Sprintf("%d failed", failed))
	}

	if pending > 0 {
		acc = append(acc, color.New(pendingColor).Sprintf("%d pending", pending))
	}

	if undefined > 0 {
		acc = append(acc, color.New(undefinedColor).Sprintf("%d undefined", undefined))
	}

	if skipped > 0 {
		acc = append(acc, color.New(skippedColor).Sprintf("%d skipped", skipped))
	}

	return strings.Join(acc, ", ")
}
