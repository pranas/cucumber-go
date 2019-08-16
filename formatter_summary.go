package cucumber

import (
	"fmt"
	"io"
	"strings"
	"time"

	messages "github.com/cucumber/cucumber-messages-go/v3"
	"github.com/fatih/color"
)

type stepDescription struct {
	ScenarioName     string
	ScenarioLocation string
	StepName         string
	StepLocation     string
	Error            string
}

type summaryFormatter struct {
	out         io.Writer
	failedSteps []stepDescription
	pickleMap   map[string]*messages.Pickle
	start       time.Time
	duration    time.Duration

	Success            bool
	TestCasesTotal     int
	TestCasesPassed    int
	TestCasesFailed    int
	TestCasesPending   int
	TestCasesUndefined int
	StepsTotal         int
	StepsPassed        int
	StepsFailed        int
	StepsPending       int
	StepsUndefined     int
	StepsSkipped       int
}

func NewSummaryFormatter(stdout io.Writer) *summaryFormatter {
	return &summaryFormatter{
		out:       stdout,
		pickleMap: map[string]*messages.Pickle{},
	}
}

func (sf *summaryFormatter) ProcessMessage(msg *messages.Envelope) {
	switch m := msg.Message.(type) {
	case *messages.Envelope_TestRunStarted:
		sf.start = time.Now()
	case *messages.Envelope_TestRunFinished:
		sf.duration = time.Since(sf.start)
		sf.Success = m.TestRunFinished.Success
		sf.displaySummary()
	case *messages.Envelope_CommandInitializeTestCase:
		sf.TestCasesTotal += 1

		sf.pickleMap[m.CommandInitializeTestCase.Pickle.Id] = m.CommandInitializeTestCase.Pickle
	case *messages.Envelope_TestCaseFinished:
		switch m.TestCaseFinished.TestResult.Status {
		case messages.TestResult_PASSED:
			sf.TestCasesPassed += 1
		case messages.TestResult_FAILED:
			sf.TestCasesFailed += 1
		case messages.TestResult_PENDING:
			sf.TestCasesPending += 1
		case messages.TestResult_UNDEFINED:
			sf.TestCasesUndefined += 1
		}
	case *messages.Envelope_TestStepFinished:
		sf.StepsTotal += 1

		switch m.TestStepFinished.TestResult.Status {
		case messages.TestResult_PASSED:
			sf.StepsPassed += 1
		case messages.TestResult_FAILED:
			sf.StepsFailed += 1

			pickle := sf.pickleMap[m.TestStepFinished.PickleId]
			pickleLocation := pickle.Locations[len(pickle.Locations)-1].Line

			step := pickle.Steps[m.TestStepFinished.Index]
			stepLocation := step.Locations[len(step.Locations)-1].Line

			sf.failedSteps = append(sf.failedSteps, stepDescription{
				ScenarioName:     pickle.Name,
				ScenarioLocation: fmt.Sprintf("%s:%d", pickle.Uri, pickleLocation),
				StepName:         step.Text,
				StepLocation:     fmt.Sprintf("%s:%d", pickle.Uri, stepLocation),
				Error:            m.TestStepFinished.TestResult.Message,
			})
		case messages.TestResult_PENDING:
			sf.StepsPending += 1
		case messages.TestResult_UNDEFINED:
			sf.StepsUndefined += 1
		case messages.TestResult_SKIPPED:
			sf.StepsSkipped += 1
		}
	}
}

func (sf *summaryFormatter) displaySummary() {
	if len(sf.failedSteps) > 0 {
		color.New(failureColor).Fprint(sf.out, "\n\nFailed steps:\n")
		for _, fs := range sf.failedSteps {
			color.New(failureColor).Fprintf(sf.out, "\n  Scenario: %s", fs.ScenarioName)
			color.New(color.FgBlack).Fprintf(sf.out, " # %s\n", fs.ScenarioLocation)
			color.New(failureColor).Fprintf(sf.out, "    %s", fs.StepName)
			color.New(color.FgBlack).Fprintf(sf.out, " # %s\n", fs.StepLocation)
			color.New(failureColor).Fprint(sf.out, "      Error: ")
			color.New(color.FgHiRed).Fprintf(sf.out, "%s\n", fs.Error)
		}
	}

	fmt.Fprint(sf.out, "\n")
	scenarioStatusSummary := statusSummary(sf.TestCasesPassed, sf.TestCasesFailed, sf.TestCasesPending, sf.TestCasesUndefined, 0)
	fmt.Fprintf(sf.out, "%d scenarios (%s)\n", sf.TestCasesTotal, scenarioStatusSummary)

	stepStatusSummary := statusSummary(sf.StepsPassed, sf.StepsFailed, sf.StepsPending, sf.StepsUndefined, sf.StepsSkipped)
	fmt.Fprintf(sf.out, "%d steps (%s)\n", sf.StepsTotal, stepStatusSummary)
	fmt.Fprintln(sf.out, sf.duration)
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
