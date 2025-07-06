package logs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"sync"
)

func ProcessLogStream(ctx context.Context, reader io.Reader, streamType string, logSender LogSender, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled for %s stream processing, stopping.", streamType)
			return
		default:
			line := scanner.Bytes()
			if len(line) == 0 {
				continue // Skip empty lines
			}

			var jsonContent interface{}
			err := json.Unmarshal(line, &jsonContent) // First attempt: unmarshal the line as-is

			var lineToSend []byte
			if err != nil {
				// If initial unmarshal fails, try to strip a potential leading timestamp.
				// Docker timestamps (with Timestamps: true) usually look like "YYYY-MM-DDTHH:MM:SS.nnnnnnnnnZ "
				// We'll find the first space and try to parse the rest.
				firstSpace := bytes.IndexByte(line, ' ')
				if firstSpace != -1 && firstSpace+1 < len(line) {
					trimmedLine := line[firstSpace+1:]
					err = json.Unmarshal(trimmedLine, &jsonContent)
					if err == nil {
						lineToSend = trimmedLine
					}
				}
			} else {
				lineToSend = line
			}

			if lineToSend != nil {
				if sendErr := logSender.SendLog(lineToSend); sendErr != nil {
					log.Printf("Error sending %s log to logger service: %v", streamType, sendErr)
				}
			} else {
				// If both attempts failed (original line and trimmed line), it's truly non-JSON.
				// In production, consider sending these to a separate non-json log sink or just discarding.
				// Uncomment the next line to log non-JSON lines.
				// log.Printf("[%s][WARN][NON-JSON]: %s", streamType, string(line))
			}
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Printf("Error reading from %s stream: %v", streamType, err)
	}
}
