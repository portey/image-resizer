package error

const (
	NotFound      ServiceError = "NotFound"
	Internal      ServiceError = "Internal"
	RaceCondition ServiceError = "RaceCondition"
)

type (
	ServiceError string
)

func (c ServiceError) Error() string {
	return string(c)
}
