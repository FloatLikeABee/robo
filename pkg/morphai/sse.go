package morphai

import (
	"bufio"
	"io"
	"strings"
)

type sseEvent struct {
	Event string
	Data  string
}

func readSSE(r io.Reader, fn func(sseEvent) error) error {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	sc.Buffer(buf, 2*1024*1024)
	var event, data strings.Builder
	flush := func() error {
		if event.Len() == 0 && data.Len() == 0 {
			return nil
		}
		err := fn(sseEvent{Event: event.String(), Data: data.String()})
		event.Reset()
		data.Reset()
		return err
	}
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if line == "" {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if strings.HasPrefix(line, "event:") {
			event.Reset()
			event.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "event:")))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return flush()
}
