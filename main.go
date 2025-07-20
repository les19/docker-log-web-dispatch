// main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/les19/docker-log-web-dispatch/config"
	dockerclient "github.com/les19/docker-log-web-dispatch/docker"
	"github.com/les19/docker-log-web-dispatch/logs"

	"github.com/docker/docker/api/types/container" // Only for container.ListOptions
)

func main() {
	cfg := config.LoadConfig()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("Starting Go Docker Log Listener...")
	log.Printf("Logger Service URL: %s", cfg.LoggerServiceURL)
	log.Printf("HTTP Client Timeout: %s", cfg.HTTPClientTimeout)
	log.Printf("Log Tail Count: %s", cfg.LogTailCount)
	log.Printf("Listen to All Containers: %t", cfg.ListenAll)
	log.Printf("Container Name Filters: %v", cfg.ContainerNameFilters)
	log.Println("--------------------------------------------------------------------")
	log.Println("IMPORTANT: This container must be run with '-v /var/run/docker.sock:/var/run/docker.sock'")
	log.Println("Press Ctrl+C to stop the program.")
	log.Println("--------------------------------------------------------------------")

	cli, err := dockerclient.NewDockerClient()
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer cli.Close()

	logSender := logs.NewHTTPLogSender(cfg.LoggerServiceURL, cfg.LoggerAuthHeaderName, cfg.LoggerAuthHeaderValue, cfg.HTTPClientTimeout)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
		cancel()
	}()

	allContainers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list containers: %v", err)
	}

	var initialMatchingContainers []container.Summary
	for _, c := range allContainers {
		if dockerclient.MatchesContainerNameFilter(c.Names, cfg.ContainerNameFilters) {
			initialMatchingContainers = append(initialMatchingContainers, c)
		}
	}

	if len(initialMatchingContainers) == 0 {
		log.Println("No running containers matching the name filters found initially.")
		if !cfg.ListenAll {
			log.Println("LISTEN_ALL_CONTAINERS is false. Exiting in 5 seconds as no initial matching containers were found.")
			time.Sleep(5 * time.Second)

			return
		}
		log.Println("Will monitor for new containers if LISTEN_ALL_CONTAINERS is true.")
	} else {
		log.Printf("Found %d running container(s) matching filters:", len(initialMatchingContainers))
		for _, c := range initialMatchingContainers {
			log.Printf("  ID: %s, Name(s): %v, Image: %s", c.ID[:12], c.Names, c.Image)
		}
	}

	var wg sync.WaitGroup

	if cfg.ListenAll {

		wg.Add(1)
		go dockerclient.MonitorContainers(ctx, cli, logSender, cfg.ContainerNameFilters, &wg)

		for _, c := range initialMatchingContainers {
			wg.Add(1)
			go dockerclient.ListenForContainerLogs(ctx, cli, c.ID, c.Names[0], cfg.LogTailCount, logSender, &wg)
		}
	} else {
		if len(initialMatchingContainers) > 0 {
			targetContainer := initialMatchingContainers[0]
			wg.Add(1)
			go dockerclient.ListenForContainerLogs(ctx, cli, targetContainer.ID, targetContainer.Names[0], cfg.LogTailCount, logSender, &wg)
		} else {
			log.Println("No matching containers to listen to and LISTEN_ALL_CONTAINERS is false. Exiting.")

			return
		}
	}

	wg.Wait()

	log.Println("Program finished.")
}
