package image

import "context"

type ImageServiceHandler interface {
	Pull(ctx context.Context, ref string) (PullResult, error)
}
