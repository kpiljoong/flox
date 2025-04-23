package output

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type StdoutOutput struct {
	writer io.Writer
	ctx    context.Context
}

func NewStdoutOutput(ctx context.Context) *StdoutOutput {
	return &StdoutOutput{
		writer: os.Stdout,
		ctx:    ctx,
	}
}

func NewStdoutOutputWithWriter(ctx context.Context, w io.Writer) *StdoutOutput {
	return &StdoutOutput{
		writer: w,
		ctx:    ctx,
	}
}

func (o *StdoutOutput) Send(event map[string]interface{}) error {
	select {
	case <-o.ctx.Done():
		return o.ctx.Err()
	default:
		b, err := json.Marshal(event)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(o.writer, string(b))
		return err
	}
}
