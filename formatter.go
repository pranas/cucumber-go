package cucumber

import (
	"fmt"

	messages "github.com/cucumber/cucumber-messages-go/v3"
)

type Formatter interface {
	ProcessMessage(msg *messages.Envelope)
}

type debugFormatter struct{}

func (df *debugFormatter) ProcessMessage(msg *messages.Envelope) {
	fmt.Printf("cucumber-engine OUT: %+v\n", msg)
}

type nopFormatter struct{}

func (nf *nopFormatter) ProcessMessage(msg *messages.Envelope) {
}
