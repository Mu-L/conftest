package runner

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/policy"
	"github.com/open-policy-agent/opa/v1/tester"
	"github.com/open-policy-agent/opa/v1/topdown"
)

// VerifyRunner is the runner for the Verify command, executing
// Rego policy unit-tests.
type VerifyRunner struct {
	Capabilities      string
	RegoVersion       string `mapstructure:"rego-version"`
	Policy            []string
	Data              []string
	Output            string
	NoColor           bool `mapstructure:"no-color"`
	Trace             bool
	Strict            bool
	Report            string
	Quiet             bool
	ShowBuiltinErrors bool `mapstructure:"show-builtin-errors"`
}

const (
	ReportFull  = "full"
	ReportNotes = "notes"
	ReportFails = "fails"
)

// Run executes the Rego tests for the given policies.
func (r *VerifyRunner) Run(ctx context.Context) (output.CheckResults, []*tester.Result, error) {
	capabilities, err := policy.LoadCapabilities(r.Capabilities)
	if err != nil {
		return nil, nil, fmt.Errorf("load capabilities: %w", err)
	}
	opts := policy.CompilerOptions{
		Strict:       r.Strict,
		RegoVersion:  r.RegoVersion,
		Capabilities: capabilities,
	}
	engine, err := policy.LoadWithData(r.Policy, r.Data, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("load: %w", err)
	}

	// Traces should be enabled when Trace or Report options are on
	enableTracing := r.Trace || r.IsReportOptionOn()
	if enableTracing {
		engine.EnableTracing()
	}

	runner := tester.NewRunner().
		SetCompiler(engine.Compiler()).
		SetStore(engine.Store()).
		SetModules(engine.Modules()).
		EnableTracing(enableTracing).
		SetRuntime(engine.Runtime()).
		RaiseBuiltinErrors(r.ShowBuiltinErrors)
	ch, err := runner.RunTests(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("running tests: %w", err)
	}

	var results output.CheckResults
	var rawResults []*tester.Result
	for result := range ch {
		if result.Error != nil {
			return nil, nil, fmt.Errorf("run test: %w", result.Error)
		}

		rawResults = append(rawResults, result)
		buf := new(bytes.Buffer)
		topdown.PrettyTrace(buf, result.Trace)
		var traces []string
		for _, line := range strings.Split(buf.String(), "\n") {
			if len(line) > 0 {
				traces = append(traces, line)
			}
		}

		var outputResult output.Result
		if result.Fail || result.Skip {
			outputResult.Message = result.Package + "." + result.Name
		}

		queryResult := output.QueryResult{
			Query:   result.Name,
			Results: []output.Result{outputResult},
			Traces:  traces,
		}

		checkResult := output.CheckResult{
			FileName: result.Location.File,
			Queries:  []output.QueryResult{queryResult},
		}
		if result.Fail {
			checkResult.Failures = []output.Result{outputResult}
		} else if result.Skip {
			checkResult.Skipped = []output.Result{outputResult}
		} else {
			checkResult.Successes++
		}

		results = append(results, checkResult)
	}

	return results, rawResults, nil
}

// IsReportOptionOn returns true if the reporting option is turned on, otherwise false.
func (r *VerifyRunner) IsReportOptionOn() bool {
	return r.Report == ReportFull ||
		r.Report == ReportNotes ||
		r.Report == ReportFails
}
