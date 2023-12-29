package ogcore

import "context"

type Action func(ctx context.Context, state State) error
