package queryplan

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
)

type CallStackEntry struct {
	FunctionName string `json:"function_name"`
	FileName     string `json:"file_name"`
	LineNumber   int    `json:"line_number"`
}

func parseCallStackLine(line string) (CallStackEntry, error) {
	pattern := regexp.MustCompile(`^(.*)\s+(.*):(\d+)$`)
	matches := pattern.FindStringSubmatch(line)

	if len(matches) < 4 {
		return CallStackEntry{}, fmt.Errorf("invalid call stack line: %s", line)
	}

	lineNumber, err := strconv.Atoi(matches[3])
	if err != nil {
		return CallStackEntry{}, fmt.Errorf("invalid line number in call stack: %s", matches[3])
	}

	return CallStackEntry{
		FunctionName: matches[1],
		FileName:     matches[2],
		LineNumber:   lineNumber,
	}, nil
}

func captureCallStack() []string {
	const size = 64 // Adjust size as needed
	var pcs [size]uintptr
	n := runtime.Callers(2, pcs[:]) // Skip first few callers to avoid runtime functions
	frames := runtime.CallersFrames(pcs[:n])

	var callStack []string
	for {
		frame, more := frames.Next()
		callStack = append(callStack, frame.Function+" "+frame.File+":"+itoa(frame.Line))
		if !more {
			break
		}
	}

	return callStack
}
