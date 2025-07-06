package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/les19/docker-log-web-dispatch/logs"
)

func ListenForContainerLogs(ctx context.Context, cli *client.Client, containerID string, containerName string, tailCount string, logSender logs.LogSender, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("Starting log stream for container: %s (%s)", containerName, containerID[:12])

	logOptions := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: true,
		Tail:       tailCount,
	}

	reader, err := cli.ContainerLogs(ctx, containerID, logOptions)
	if err != nil {
		log.Printf("Failed to get container logs for %s (%s): %v", containerName, containerID[:12], err)
		return
	}
	defer reader.Close()

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	var streamWg sync.WaitGroup
	streamWg.Add(3)

	go func() {
		defer streamWg.Done()
		defer stdoutWriter.Close()
		defer stderrWriter.Close()
		// stdcopy.StdCopy will write to stdoutWriter and stderrWriter
		if _, err := stdcopy.StdCopy(stdoutWriter, stderrWriter, reader); err != nil && err != io.EOF {
			// io.EOF is expected when the container stops
			log.Printf("Error demultiplexing Docker logs for %s (%s): %v", containerName, containerID[:12], err)
		}
	}()

	go logs.ProcessLogStream(ctx, stdoutReader, fmt.Sprintf("%s-STDOUT", containerName), logSender, &streamWg)

	go logs.ProcessLogStream(ctx, stderrReader, fmt.Sprintf("%s-STDERR", containerName), logSender, &streamWg)

	streamWg.Wait()

	log.Printf("Log streaming finished for container: %s (%s)", containerName, containerID[:12])
}

func MonitorContainers(ctx context.Context, cli *client.Client, logSender logs.LogSender, nameFilters []string, mainWg *sync.WaitGroup) {
	defer mainWg.Done()

	activeListeners := make(map[string]struct{})
	var mu sync.Mutex

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Container monitoring stopped.")
			return
		case <-ticker.C:
			containers, err := cli.ContainerList(ctx, container.ListOptions{})
			if err != nil {
				log.Printf("Error listing containers during monitoring: %v", err)
				continue
			}

			mu.Lock()
			for _, c := range containers {
				// Apply the name filter here
				if MatchesContainerNameFilter(c.Names, nameFilters) {
					if _, exists := activeListeners[c.ID]; !exists {
						log.Printf("Discovered new container matching filter: %s (%s)", c.Names[0], c.ID[:12])
						activeListeners[c.ID] = struct{}{} // Mark as active

						mainWg.Add(1)
						go ListenForContainerLogs(ctx, cli, c.ID, c.Names[0], "", logSender, mainWg)
					}
				}
			}
			mu.Unlock()
		}
	}
}
