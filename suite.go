package cucumber

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/cucumber/cucumber-engine/src/runner"
	messages "github.com/cucumber/cucumber-messages-go/v3"
)

type orderType = messages.SourcesOrderType

const (
	OrderRandom     orderType = messages.SourcesOrderType_RANDOM
	OrderDefinition           = messages.SourcesOrderType_ORDER_OF_DEFINITION
)

type stepHandlerFunc func(TestCase, ...string) error

type testCaseInitializerFunc func(TestCase) error

type stepDefinition struct {
	Pattern string
	Handler stepHandlerFunc
}

type Summary struct {
	Success  bool
	ExitCode int
	Duration time.Duration

	TestCasesTotal  int
	TestCasesPassed int
	StepsTotal      int
	StepsPassed     int
	StepsSkipped    int
}

type suite struct {
	language      string
	baseDirectory string
	files         []string
	seed          uint64
	order         orderType

	formatter           Formatter
	stepDefinitions     []stepDefinition
	testCases           sync.Map
	testCaseInitializer testCaseInitializerFunc
	incoming            chan *messages.Envelope
	outgoing            chan *messages.Envelope
}

func NewSuite(args ...string) (*suite, error) {
	fs := flag.NewFlagSet("cucumber", flag.ContinueOnError)
	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	e := runner.NewRunner()
	incoming, outgoing := e.GetCommandChannels()

	baseDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	files, err := filepath.Glob("features/*")
	if err != nil {
		return nil, err
	}

	for i := range files {
		files[i], err = filepath.Abs(files[i])
		if err != nil {
			return nil, err
		}
	}

	suite := &suite{
		testCaseInitializer: func(TestCase) error { return nil },
		incoming:            incoming,
		outgoing:            outgoing,

		formatter:     &nopFormatter{},
		language:      "en",
		baseDirectory: baseDirectory,
		files:         files,
		seed:          uint64(time.Now().Unix()),
		order:         OrderRandom,
	}

	//suite.formatter = &debugFormatter{}

	return suite, nil
}

func (s *suite) DefineTestCaseInitializer(fn testCaseInitializerFunc) {
	s.testCaseInitializer = fn
}

func (s *suite) DefineStep(pattern string, fn stepHandlerFunc) {
	s.stepDefinitions = append(s.stepDefinitions, stepDefinition{
		Pattern: pattern,
		Handler: fn,
	})
}

func (s *suite) Run() Summary {
	resultCh := make(chan Summary)
	go s.listen(resultCh)

	var stepDefinitionConfig []*messages.StepDefinitionConfig

	for i, sd := range s.stepDefinitions {
		stepDefinitionConfig = append(stepDefinitionConfig, &messages.StepDefinitionConfig{
			Id: strconv.Itoa(i),
			Pattern: &messages.StepDefinitionPattern{
				Source: sd.Pattern,
				Type:   messages.StepDefinitionPatternType_REGULAR_EXPRESSION,
			},
		})
	}

	supportCodeConfig := messages.SupportCodeConfig{
		StepDefinitionConfigs: stepDefinitionConfig,
	}

	s.respond(&messages.Envelope{
		Message: &messages.Envelope_CommandStart{
			CommandStart: &messages.CommandStart{
				BaseDirectory: s.baseDirectory,
				RuntimeConfig: &messages.RuntimeConfig{
					IsFailFast:  false,
					IsDryRun:    false,
					IsStrict:    false,
					MaxParallel: 0, // unbound
				},
				SupportCodeConfig: &supportCodeConfig,
				SourcesConfig: &messages.SourcesConfig{
					Language:      s.language,
					AbsolutePaths: s.files,
					Filters:       &messages.SourcesFilterConfig{},
					Order: &messages.SourcesOrder{
						Type: messages.SourcesOrderType_RANDOM,
						Seed: s.seed,
					},
				},
			},
		},
	})

	started := time.Now()
	result := <-resultCh
	result.Duration = time.Since(started)

	if !result.Success {
		result.ExitCode = 1
	}

	s.formatter.DisplaySummary(result)

	return result
}

func (s *suite) listen(resultCh chan Summary) {
	summary := Summary{}

	for command := range s.outgoing {
		s.formatter.ProcessMessage(command)

		switch x := command.Message.(type) {
		case *messages.Envelope_TestRunFinished:
			summary.Success = x.TestRunFinished.Success
			break
		case *messages.Envelope_CommandRunBeforeTestRunHooks:
			s.respond(&messages.Envelope{
				Message: &messages.Envelope_CommandActionComplete{
					CommandActionComplete: &messages.CommandActionComplete{
						CompletedId: x.CommandRunBeforeTestRunHooks.ActionId,
						Result: &messages.CommandActionComplete_TestResult{
							TestResult: &messages.TestResult{
								Status: messages.TestResult_PASSED,
							},
						},
					},
				},
			})
		case *messages.Envelope_CommandRunAfterTestRunHooks:
			s.respond(&messages.Envelope{
				Message: &messages.Envelope_CommandActionComplete{
					CommandActionComplete: &messages.CommandActionComplete{
						CompletedId: x.CommandRunAfterTestRunHooks.ActionId,
						Result: &messages.CommandActionComplete_TestResult{
							TestResult: &messages.TestResult{
								Status: messages.TestResult_PASSED,
							},
						},
					},
				},
			})
		case *messages.Envelope_CommandGenerateSnippet:
			s.respond(&messages.Envelope{
				Message: &messages.Envelope_CommandActionComplete{
					CommandActionComplete: &messages.CommandActionComplete{
						CompletedId: x.CommandGenerateSnippet.ActionId,
						Result: &messages.CommandActionComplete_Snippet{
							Snippet: "",
						},
					},
				},
			})
		case *messages.Envelope_CommandInitializeTestCase:
			summary.TestCasesTotal += 1
			go s.initializeTestCase(x.CommandInitializeTestCase)
		case *messages.Envelope_TestCaseFinished:
			s.testCases.Delete(x.TestCaseFinished.PickleId)
			if x.TestCaseFinished.TestResult.Status == messages.TestResult_PASSED {
				summary.TestCasesPassed += 1
			}
		case *messages.Envelope_TestStepFinished:
			summary.StepsTotal += 1
			switch x.TestStepFinished.TestResult.Status {
			case messages.TestResult_PASSED:
				summary.StepsPassed += 1
			case messages.TestResult_SKIPPED:
				summary.StepsSkipped += 1
			}
		case *messages.Envelope_CommandRunTestStep:
			go s.runTestStep(x.CommandRunTestStep)
		}
	}
	resultCh <- summary
}

func (s *suite) respond(m *messages.Envelope) {
	s.incoming <- m
}

func (s *suite) initializeTestCase(command *messages.CommandInitializeTestCase) {
	testResult := messages.TestResult{
		Status: messages.TestResult_PASSED,
	}

	tc := &testCase{}
	s.testCases.Store(command.Pickle.Id, tc)

	err := s.testCaseInitializer(tc)
	if err != nil {
		testResult.Status = messages.TestResult_FAILED
	}

	s.respond(&messages.Envelope{
		Message: &messages.Envelope_CommandActionComplete{
			CommandActionComplete: &messages.CommandActionComplete{
				CompletedId: command.ActionId,
				Result: &messages.CommandActionComplete_TestResult{
					TestResult: &testResult,
				},
			},
		},
	})
}

func (s *suite) runTestStep(command *messages.CommandRunTestStep) {
	testResult := messages.TestResult{
		Status: messages.TestResult_PASSED,
	}

	err := s.callStepHandler(command)
	if err != nil {
		testResult.Status = messages.TestResult_FAILED
	}

	s.respond(&messages.Envelope{
		Message: &messages.Envelope_CommandActionComplete{
			CommandActionComplete: &messages.CommandActionComplete{
				CompletedId: command.ActionId,
				Result: &messages.CommandActionComplete_TestResult{
					TestResult: &testResult,
				},
			},
		},
	})
}

func (s *suite) callStepHandler(command *messages.CommandRunTestStep) error {
	i, err := strconv.Atoi(command.StepDefinitionId)
	if err != nil {
		return err
	}

	var captures []string

	for _, patternMatch := range command.PatternMatches {
		// TODO: when would we get multiple captures within a single pattern match?
		captures = append(captures, patternMatch.Captures...)
	}

	testCase, _ := s.testCases.Load(command.PickleId)
	return s.stepDefinitions[i].Handler(testCase.(TestCase), captures...)
}
