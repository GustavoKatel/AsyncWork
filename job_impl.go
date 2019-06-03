package asyncwork

import (
	"time"

	"github.com/GustavoKatel/AsyncWork/interfaces"
)

type jobImpl struct {
	jobFn interfaces.JobFn
	tag   string
	delay time.Duration
}
