package output

import (
	"context"
	"encoding/json"
	"os"
)

type FileOutput struct {
	file *os.File
	ctx  context.Context
}

func NewFileOutput(ctx context.Context, path string) (*FileOutput, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &FileOutput{file: file, ctx: ctx}, nil
}

func (o *FileOutput) Send(event map[string]interface{}) error {
	select {
	case <-o.ctx.Done():
		return o.ctx.Err()
	default:
		b, err := json.Marshal(event)
		if err != nil {
			return err
		}
		_, err = o.file.Write(append(b, '\n'))
		return err
	}
}
