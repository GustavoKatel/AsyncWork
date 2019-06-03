package asyncwork

import (
	"time"

	"github.com/GustavoKatel/asyncwork/interfaces"
)

type jobImpl struct {
	jobFn interfaces.JobFn
	tag   string
	delay time.Duration
}
