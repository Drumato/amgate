package action

import (
	"context"

	"github.com/Drumato/amgate/pkg/dispatcher"
)

type Action interface {
	Name() string
	Run(
		ctx context.Context,
		result dispatcher.DispatchResult,
	) error
}
