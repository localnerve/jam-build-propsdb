// main.go
//
// A scalable, high performance drop-in replacement for the jam-build nodejs data service
// Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC
//
// This file is part of jam-build-propsdb.
// jam-build-propsdb is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later version.
// jam-build-propsdb is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
// You should have received a copy of the GNU Affero General Public License along with jam-build-propsdb.
// If not, see <https://www.gnu.org/licenses/>.
// Additional terms under GNU AGPL version 3 section 7:
// a) The reasonable legal notice of original copyright and author attribution must be preserved
//    by including the string: "Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC"
//    in this material, copies, or source code of derived works.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/localnerve/jam-build-propsdb/tests/helpers"
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
	var err error
	testContainers, err = helpers.CreateAllTestContainers(nil)
	if err != nil {
		log.Fatalf("Failed to create test containers: %v\n", err)
	}

	// Wait for signal or interactive termination
	if os.Getenv("HOST_DEBUG") == "true" {
		log.Printf(">>> If Debugging, PRESS ENTER TO TERMINATE AND COLLECT COVERAGE <<<\n")

		done := make(chan bool)
		go func() {
			buf := make([]byte, 1)
			_, _ = os.Stdin.Read(buf)
			done <- true
		}()

		select {
		case sig := <-sigs:
			log.Printf("\nReceived signal: %v, terminating test containers...\n", sig)
		case <-done:
			log.Printf("\nTermination requested via stdin, terminating test containers...\n")
		}
	} else {
		sig := <-sigs
		log.Printf("\nReceived signal: %v, terminating test containers...\n", sig)
	}

	if testContainers != nil {
		testContainers.Terminate(nil)
	}
}
