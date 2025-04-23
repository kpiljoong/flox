package file

import "context"

type HandlerFunc func(event map[string]interface{})

func StartFile(ctx context.Context, path string, handle HandlerFunc, namespace string, trackOffset bool, startFrom string) {
	tailer := NewTailer(path, namespace, trackOffset, startFrom)
	tailer.Run(ctx, handle)
}
