package cucumber

import (
	"fmt"
	"io"
	"strings"

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
}

func NewSummaryFormatter(stdout io.Writer) *summaryFormatter {
	return &summaryFormatter{
		out:       stdout,
		pickleMap: map[string]*messages.Pickle{},
	}
}

func (sf *summaryFormatter) ProcessMessage(msg *messages.Envelope) {
	switch m := msg.Message.(type) {
	case *messages.Envelope_CommandInitializeTestCase:
		sf.pickleMap[m.CommandInitializeTestCase.Pickle.Id] = m.CommandInitializeTestCase.Pickle
	case *messages.Envelope_TestStepFinished:
		switch m.TestStepFinished.TestResult.Status {
		case messages.TestResult_FAILED:
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
		}
	}
}

func (sf *summaryFormatter) DisplaySummary(s Summary) {
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
	scenarioStatusSummary := statusSummary(s.TestCasesPassed, s.TestCasesFailed, s.TestCasesPending, s.TestCasesUndefined, 0)
	fmt.Fprintf(sf.out, "%d scenarios (%s)\n", s.TestCasesTotal, scenarioStatusSummary)

	stepStatusSummary := statusSummary(s.StepsPassed, s.StepsFailed, s.StepsPending, s.StepsUndefined, s.StepsSkipped)
	fmt.Fprintf(sf.out, "%d steps (%s)\n", s.StepsTotal, stepStatusSummary)
	fmt.Fprintln(sf.out, s.Duration)
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
