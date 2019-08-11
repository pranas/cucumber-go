package cucumber

import (
	"fmt"

	messages "github.com/cucumber/cucumber-messages-go/v3"
)

type Formatter interface {
	ProcessMessage(msg *messages.Envelope)
	DisplaySummary(summary Summary)
}

type debugFormatter struct{}

func (df *debugFormatter) ProcessMessage(msg *messages.Envelope) {
	fmt.Printf("cucumber-engine OUT: %+v\n", msg)
}

func (df *debugFormatter) DisplaySummary(summary Summary) {
	fmt.Printf("summary: %+v\n", summary)
}

type nopFormatter struct{}

func (nf *nopFormatter) ProcessMessage(msg *messages.Envelope) {
}
func (nf *nopFormatter) DisplaySummary(summary Summary) {
}
