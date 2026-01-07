package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/localnerve/propsdb/tests/helpers"
)

func main() {
	var showHelp bool
	flag.BoolVar(&showHelp, "h", false, "show help")
	var envFilename string
	flag.StringVar(&envFilename, "f", "", "path to the .env file")
	flag.Parse()

	usage := `
Run the propsdb testcontainers with the environment variables from the .env file.

Usage:

testcontainers [-h] [-f ENV_FILE_PATH]

ENV_FILE_PATH: path to the .env file

example
  testcontainers -f /path/to/something/.env
`
	// if -h flag print usage and return
	if showHelp {
		fmt.Println(usage)
		return
	}

	if envFilename != "" {
		log.Printf("Loading environment variables from %s\n", envFilename)
		if err := godotenv.Load(envFilename); err != nil {
			log.Fatalf("Failed to load environment variables: %v\n", err)
		}
	} else {
		log.Printf("No environment file specified, using current environment variables\n")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP, syscall.SIGQUIT)

	var testContainers *helpers.TestContainers
	go func() {
		var err error
		testContainers, err = helpers.CreateAllTestContainers(nil)
		if err != nil {
			log.Fatalf("Failed to create test containers: %v\n", err)
		}
	}()

	sig := <-sigs
	log.Printf("\nReceived signal: %v, terminating test containers...\n", sig)
	if testContainers != nil {
		testContainers.Terminate(nil)
	}
}
