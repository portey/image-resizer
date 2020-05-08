package errors

import "strings"

const (
	NotFound      ServiceError = "NotFound"
	Internal      ServiceError = "Internal"
	RaceCondition ServiceError = "RaceCondition"
)

type (
	ServiceError string

	InvalidParam struct {
		Param   string
		Message string
	}
	InvalidParams []InvalidParam
)

func (c ServiceError) Error() string {
	return string(c)
}

func (c InvalidParams) Error() string {
	messages := make([]string, len(c))
	for i, param := range c {
		messages[i] = param.Param + ":" + param.Message
	}

	return strings.Join(messages, ", ")
}
