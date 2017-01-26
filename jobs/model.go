package jobs

import (
	"fmt"
	"strconv"
	"strings"
)

// Job representation of job
type Job struct {
	Title         string   `json:"title,omitempty"`
	Description   string   `json:"description,omitempty"`
	Salary        float64  `json:"salario,omitempty"`
	City          []string `json:"cidade,omitempty"`
	CityFormatted []string `json:"cidadeFormated,omitempty"`
}

// ID from job using title, salary and city
func (j Job) ID() string {
	return createID(j.Title, strconv.FormatFloat(j.Salary, 'f', 2, 64), strings.Join(j.City, " "))
}

const (
	JOB0000 string = "JOB0000" //unknown
	JOB1001 string = "JOB1001" //invalid
	JOB1002 string = "JOB1002" //item not found
)

type JobError struct {
	ErrCode string    `json:"error,omitempty"`
	Message string    `json:"message,omitempty"`
	ErrType ErrorType `json:"-"`
}

//ErrorType error types
type ErrorType int

//ErrorType error types values
const (
	ERROR_UNKNOWN ErrorType = iota
	ERROR_HTTP
	ERROR_INVALID
)

//NewJobErrorr JobError constructor
func newJobError(code, message string, errType ErrorType) *JobError {
	return &JobError{ErrCode: code, Message: message, ErrType: errType}
}

func (e *JobError) Error() string {
	return e.Message
}

func (e *JobError) Type() ErrorType {
	return e.ErrType
}

//NewUnknownError constructor unknown error
func NewUnknownError(msg string) *JobError {
	return newJobError(JOB0000, msg, ERROR_UNKNOWN)
}

//NewHTTPError constructor http echo error
func NewHTTPError(code int, msg string) *JobError {
	return newJobError(fmt.Sprintf("HTTP%d", code), msg, ERROR_HTTP)
}

//NewInvalidRequestError constructor invalid request
func NewInvalidRequestError(msg string) *JobError {
	return newJobError(JOB1001, msg, ERROR_INVALID)
}

//NewNotFoundError constructor not found request
func NewNotFoundError(msg string) *JobError {
	return newJobError(JOB1002, msg, ERROR_INVALID)
}
