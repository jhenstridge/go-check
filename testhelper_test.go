package check_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/check.v1"
)

var (
	helperRunFlag  = flag.String("helper.run", "", "Run helper suite")
	helperFailFlag = flag.String("helper.fail", "", "")
)

func TestHelperSuite(t *testing.T) {
	if helperRunFlag == nil || *helperRunFlag == "" {
		t.SkipNow()
	}
	switch *helperRunFlag {
	case "FailHelper":
		check.Run(t, &FailHelper{})
	case "SuccessHelper":
		check.Run(t, &SuccessHelper{})
	case "FixtureHelper":
		suite := &FixtureHelper{}
		if helperFailFlag != nil {
			suite.failOn = *helperFailFlag
		}
		check.Run(t, suite)
	default:
		t.Skip()
	}
}

// testEvent represents a json formatted event
type testEvent struct {
	Action  string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

type helperResult []testEvent

func (result helperResult) Status(test string) string {
	for _, event := range result {
		if event.Test != "TestHelperSuite/"+test {
			continue
		}
		switch event.Action {
		case "pass", "fail", "skip":
			return event.Action
		}
	}
	return ""
}

var isStatusLine = regexp.MustCompile(`^\s*(?:===|---) `).MatchString

func (result helperResult) Logs(test string) string {
	var lines []string
	for _, event := range result {
		if event.Test != "TestHelperSuite/"+test {
			continue
		}
		if event.Action == "output" && !isStatusLine(event.Output) {
			lines = append(lines, event.Output)
		}
	}
	return strings.Join(lines, "")
}

func runHelperSuite(c *check.C, name string, args ...string) (code int, output helperResult) {
	c.Helper()
	args = append([]string{"tool", "test2json", os.Args[0], "-test.v", "-test.run", "TestHelperSuite", "-helper.run", name}, args...)
	cmd := exec.Command("go", args...)
	data, err := cmd.Output()
	if execErr, ok := err.(*exec.ExitError); ok {
		code = execErr.ExitCode()
		err = nil
	}
	if err != nil {
		c.Fatal(err)
	}
	lines := bytes.Split(data, []byte("\n"))
	output = make(helperResult, 0, len(lines))
	for _, l := range lines {
		if len(l) == 0 {
			continue
		}
		var event testEvent
		if err := json.Unmarshal(l, &event); err != nil {
			c.Fatal(err)
		}
		output = append(output, event)
	}
	return code, output
}
