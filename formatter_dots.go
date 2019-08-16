package cucumber

// Based on https://github.com/cucumber/dots-formatter-go/blob/master/dots.go

import (
	"fmt"
	"io"

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
	out     io.Writer
	summary *summaryFormatter
}

func NewDotFormatter(stdout io.Writer) *dotFormatter {
	return &dotFormatter{
		out:     stdout,
		summary: NewSummaryFormatter(stdout),
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

	df.summary.ProcessMessage(msg)
}

func (df *dotFormatter) DisplaySummary(summary Summary) {
	df.summary.DisplaySummary(summary)
}
