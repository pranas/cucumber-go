package cucumber

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/cucumber/cucumber-engine/src/runner"
	messages "github.com/cucumber/cucumber-messages-go/v3"
)

type OrderType uint8

const (
	OrderRandom OrderType = 0
	OrderDefinition
)

var (
	ErrPending = errors.New("implementation pending")
)

type stepHandlerFunc func(TestCase, ...string) error

type testCaseInitializerFunc func(TestCase) error

type stepDefinition struct {
	Pattern string
	Handler stepHandlerFunc
}

type suite struct {
	config              Config
	baseDirectory       string
	files               []string
	stepDefinitions     []stepDefinition
	testCases           sync.Map
	testCaseInitializer testCaseInitializerFunc
	incoming            chan *messages.Envelope
	outgoing            chan *messages.Envelope
}

func NewSuite(config Config, args ...string) (*suite, error) {
	if config.Language == "" {
		config.Language = "en"
	}

	if config.Seed == 0 {
		config.Seed = uint64(time.Now().Unix())
	}

	if config.Formatter == nil {
		config.Formatter = NewDotFormatter(os.Stdout)
	}

	if len(config.Paths) == 0 {
		config.Paths = []string{"features/"}
	}

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(ioutil.Discard)
	fs.Usage = func() {}
	fs.StringVar(&config.Language, "lang", config.Language, "")
	fs.Uint64Var(&config.Seed, "seed", config.Seed, "")
	fs.Uint64Var(&config.Concurrency, "concurrency", config.Concurrency, "")
	fs.Uint64Var(&config.Concurrency, "c", config.Concurrency, "")
	fs.BoolVar(&config.FailFast, "fast", config.FailFast, "")
	fs.BoolVar(&config.DryRun, "dry", config.DryRun, "")
	fs.BoolVar(&config.Strict, "strict", config.Strict, "")
	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	if len(fs.Args()) > 0 {
		config.Paths = fs.Args()
	}

	e := runner.NewRunner()
	incoming, outgoing := e.GetCommandChannels()

	baseDirectory, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var files []string

	for _, path := range config.Paths {
		filesForPath, err := findFeatures(path)
		if err != nil {
			return nil, fmt.Errorf("failed to find features in path: %s", path)
		}
		files = append(files, filesForPath...)
	}

	for i := range files {
		files[i], err = filepath.Abs(files[i])
		if err != nil {
			return nil, err
		}
	}

	suite := &suite{
		config:              config,
		baseDirectory:       baseDirectory,
		files:               files,
		testCaseInitializer: func(TestCase) error { return nil },
		incoming:            incoming,
		outgoing:            outgoing,
	}

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

	order := messages.SourcesOrderType_RANDOM
	if s.config.Order == OrderDefinition {
		order = messages.SourcesOrderType_ORDER_OF_DEFINITION
	}

	s.respond(&messages.Envelope{
		Message: &messages.Envelope_CommandStart{
			CommandStart: &messages.CommandStart{
				BaseDirectory: s.baseDirectory,
				RuntimeConfig: &messages.RuntimeConfig{
					IsFailFast:  s.config.FailFast,
					IsDryRun:    s.config.DryRun,
					IsStrict:    s.config.Strict,
					MaxParallel: s.config.Concurrency,
				},
				SupportCodeConfig: &supportCodeConfig,
				SourcesConfig: &messages.SourcesConfig{
					Language:      s.config.Language,
					AbsolutePaths: s.files,
					Filters: &messages.SourcesFilterConfig{
						TagExpression: s.config.TagExpression,
					},
					Order: &messages.SourcesOrder{
						Type: order,
						Seed: s.config.Seed,
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

	s.config.Formatter.DisplaySummary(result)

	return result
}

func (s *suite) listen(resultCh chan Summary) {
	summary := Summary{}

	for command := range s.outgoing {
		s.config.Formatter.ProcessMessage(command)

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

			switch x.TestCaseFinished.TestResult.Status {
			case messages.TestResult_PASSED:
				summary.TestCasesPassed += 1
			case messages.TestResult_FAILED:
				summary.TestCasesFailed += 1
			case messages.TestResult_PENDING:
				summary.TestCasesPending += 1
			case messages.TestResult_UNDEFINED:
				summary.TestCasesUndefined += 1
			}
		case *messages.Envelope_TestStepFinished:
			summary.StepsTotal += 1

			switch x.TestStepFinished.TestResult.Status {
			case messages.TestResult_PASSED:
				summary.StepsPassed += 1
			case messages.TestResult_FAILED:
				summary.StepsFailed += 1
			case messages.TestResult_PENDING:
				summary.StepsPending += 1
			case messages.TestResult_UNDEFINED:
				summary.StepsUndefined += 1
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
	if err == ErrPending {
		testResult.Status = messages.TestResult_PENDING
	} else if err != nil {
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
