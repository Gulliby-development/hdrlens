package message

import "strconv"

// Problem is a positioned syntax problem in a header field.
type Problem struct {
	Code, Text string
	Line, Col  int
}

// Error renders the problem with its code and source position.
func (problem *Problem) Error() string {
	return problem.Code + " " + problem.Text + " at " +
		strconv.Itoa(problem.Line) + ":" + strconv.Itoa(problem.Col)
}
