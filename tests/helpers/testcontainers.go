// This file is a helper for running tests with testcontainers.
// It is used by the e2e tests in tests/e2e-js in a standalone executable and by other test files in the test helpers package.
// Expects environment variables to be loaded from .env files.
//

package helpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/localnerve/propsdb/data"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainers struct {
	Network                 *testcontainers.DockerNetwork
	DBContainer             testcontainers.Container
	AuthorizerContainer     testcontainers.Container
	PropsDBContainer        testcontainers.Container
	PropsDBBuilderContainer testcontainers.Container
}

func (tc *TestContainers) Terminate(t *testing.T) {
	ctx := context.Background()
	if tc.PropsDBContainer != nil {
		if err := tc.PropsDBContainer.Terminate(ctx); err != nil {
			logMessage(t, "Failed to terminate PropsDB: %v", err)
		}
	}
	if tc.PropsDBBuilderContainer != nil {
		if err := tc.PropsDBBuilderContainer.Terminate(ctx); err != nil {
			logMessage(t, "Failed to terminate PropsDB Builder: %v", err)
		}
	}
	if tc.AuthorizerContainer != nil {
		if err := tc.AuthorizerContainer.Terminate(ctx); err != nil {
			logMessage(t, "Failed to terminate Authorizer: %v", err)
		}
	}
	if tc.DBContainer != nil {
		if err := tc.DBContainer.Terminate(ctx); err != nil {
			logMessage(t, "Failed to terminate MariaDB: %v", err)
		}
	}
	if tc.Network != nil {
		if err := tc.Network.Remove(ctx); err != nil {
			logMessage(t, "Failed to remove network: %v", err)
		}
	}
}

func CreateAllTestContainers(t *testing.T) (*TestContainers, error) {
	ctx := context.Background()
	testContainers := &TestContainers{}

	debugContainer := os.Getenv("DEBUG_CONTAINER")

	// Create a network
	nw, err := network.New(ctx)
	if err != nil {
		exitWithError(t, err, "Failed to create network")
	}
	testContainers.Network = nw
	networkName := nw.Name

	// Create and start the Database container
	dbType := os.Getenv("DB_TYPE")
	dbNetworkName := os.Getenv("DB_HOST")
	tcpDbPort, err := nat.NewPort("tcp", os.Getenv("DB_PORT"))
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, "Failed to create DB port")
	}
	dbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        os.Getenv("DB_IMAGE"),
			ExposedPorts: []string{string(tcpDbPort)},

			Env:        getDBInitEnvMap(dbType),
			WaitingFor: wait.ForListeningPort(tcpDbPort).WithStartupTimeout(60 * time.Second),
			Networks:   []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: {dbNetworkName},
			},
		},
		Started: true,
	})
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, "Failed to start Database")
	}
	testContainers.DBContainer = dbContainer

	// Initialize the database(s)
	dbHost, _ := dbContainer.Host(ctx)
	dbPort, _ := dbContainer.MappedPort(ctx, tcpDbPort)
	switch dbType {
	case "postgres":
		if err := performPostgresDBInit(t, testContainers, dbHost, dbPort); err != nil {
			testContainers.Terminate(t)
			exitWithError(t, err, "Failed to initialize databases")
		}
	case "mysql", "mariadb":
		if err := performMySqlDBInit(t, testContainers, dbHost, dbPort); err != nil {
			testContainers.Terminate(t)
			exitWithError(t, err, "Failed to initialize databases")
		}
	}

	// Create and start the Authorizer container
	authzNetworkName := "authorizer"
	tcpAuthzPort, err := nat.NewPort("tcp", os.Getenv("AUTHZ_PORT"))
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, "Failed to create Authorizer port")
	}
	authzDbConnection := fmt.Sprintf("root:%s@tcp(%s:%s)/%s", os.Getenv("DB_ROOT_PASSWORD"), dbNetworkName, os.Getenv("DB_PORT"), os.Getenv("AUTHZ_DATABASE"))
	authzLogLevel := "info"
	if debugContainer == "true" {
		authzLogLevel = "debug"
	}
	authorizerContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        os.Getenv("AUTHZ_IMAGE"),
			ExposedPorts: []string{string(tcpAuthzPort)},
			Env: map[string]string{
				"ENV":           "production",
				"CLIENT_ID":     os.Getenv("AUTHZ_CLIENT_ID"),
				"PORT":          os.Getenv("AUTHZ_PORT"),
				"DATABASE_TYPE": dbType,
				"DATABASE_NAME": os.Getenv("AUTHZ_DATABASE"),
				"DATABASE_URL":  authzDbConnection,
				"ADMIN_SECRET":  os.Getenv("AUTHZ_ADMIN_SECRET"),
				"ROLES":         "admin,user",
				"DEFAULT_ROLES": "user",
				"LOG_LEVEL":     authzLogLevel,
			},
			WaitingFor: wait.ForLog("Authorizer running at PORT:").WithStartupTimeout(10 * time.Second),
			Networks:   []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: {authzNetworkName},
			},
		},
		Started: true,
	})
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, "Failed to start Authorizer")
	}
	testContainers.AuthorizerContainer = authorizerContainer

	// Log the localhost and mapped ports for Authorizer for test processes
	authzHost, _ := authorizerContainer.Host(ctx)
	authzPort, _ := authorizerContainer.MappedPort(ctx, tcpAuthzPort)
	logMessage(t, "AUTHZ_URL=%s:%s", authzHost, authzPort.Port())

	imageName := "propsdb-test:latest"

	// Check if image exists
	imageExists, err := imageExists(ctx, imageName)
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, "Failed to check if image exists")
	}

	propsdbPortNumber := os.Getenv("PORT")
	tcpPropsdbPort, err := nat.NewPort("tcp", propsdbPortNumber)
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, "Failed to create PropsDB port")
	}

	propsdbExposedPorts := []string{string(tcpPropsdbPort)}
	if debugContainer == "true" {
		propsdbExposedPorts = append(propsdbExposedPorts, "2345/tcp")
	}

	hostConfigModifier := func(hostConfig *container.HostConfig) {
		if debugContainer == "true" {
			hostConfig.PortBindings = nat.PortMap{
				"2345/tcp": []nat.PortBinding{
					{HostIP: "127.0.0.1", HostPort: "2345"}, // Force local 2345
				},
			}
			hostConfig.CapAdd = []string{"SYS_PTRACE"}
			hostConfig.SecurityOpt = []string{"apparmor:unconfined"}
		}
	}

	var waitStrategy wait.Strategy
	waitStrategy = wait.ForHTTP("/metrics").WithPort(tcpPropsdbPort).WithStartupTimeout(30 * time.Second)
	if debugContainer == "true" {
		waitStrategy = wait.ForLog("API server listening at: [::]:2345").WithStartupTimeout(5 * time.Minute)
	}

	// Create PropsDB container request (we add to it later)
	propsdbContainerRequest := testcontainers.ContainerRequest{
		ExposedPorts: propsdbExposedPorts,
		Env: map[string]string{
			"DB_TYPE":                 dbType,
			"DB_HOST":                 dbNetworkName,
			"DB_PORT":                 os.Getenv("DB_PORT"),
			"DB_APP_DATABASE":         os.Getenv("DB_APP_DATABASE"),
			"DB_APP_USER":             os.Getenv("DB_APP_USER"),
			"DB_APP_PASSWORD":         os.Getenv("DB_APP_PASSWORD"),
			"DB_USER":                 os.Getenv("DB_USER"),
			"DB_PASSWORD":             os.Getenv("DB_PASSWORD"),
			"DB_APP_CONNECTION_LIMIT": os.Getenv("DB_APP_CONNECTION_LIMIT"),
			"DB_CONNECTION_LIMIT":     os.Getenv("DB_CONNECTION_LIMIT"),
			"AUTHZ_URL":               fmt.Sprintf("http://%s:%s", authzNetworkName, os.Getenv("AUTHZ_PORT")),
			"AUTHZ_CLIENT_ID":         os.Getenv("AUTHZ_CLIENT_ID"),
			"PORT":                    propsdbPortNumber,
		},
		HostConfigModifier: hostConfigModifier,
		WaitingFor:         waitStrategy,
		Networks:           []string{networkName},
	}

	if debugContainer == "true" {
		propsdbContainerRequest.Entrypoint = []string{
			"/usr/local/bin/dlv",
			"--listen=:2345",
			"--headless=true",
			"--api-version=2",
			"--accept-multiclient",
			"exec",
			"./propsdb",
		}
	}

	if !imageExists {
		// Build PropsDB builder image and add fromDockerfile to PropsDB container request
		propsdbResourceReaperSessionID := uuid.New().String()

		propsdbBuildArgs := map[string]*string{
			"RESOURCE_REAPER_SESSION_ID": &propsdbResourceReaperSessionID,
		}
		if debugContainer == "true" {
			propsdbBuildArgs["DEBUG"] = &debugContainer
		}

		buildContext := os.Getenv("TESTCONTAINERS_BUILD_CONTEXT")
		if buildContext == "" {
			buildContext = "../.."
		}

		logMessage(t, "Image %s does not exist, building...", imageName)
		propsdbBuilderContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context:    buildContext,
					Dockerfile: "Dockerfile",
					Repo:       "propsdb-test-builder",
					Tag:        "latest",
					BuildArgs:  propsdbBuildArgs,
					BuildOptionsModifier: func(opts *build.ImageBuildOptions) {
						opts.Target = "builder" // Build specific stage
					},
					PrintBuildLog: true,
				},
			},
			Started: false,
		})
		if err != nil {
			testContainers.Terminate(t)
			exitWithError(t, err, "Failed to build propsdb-test-builder")
		}
		testContainers.PropsDBBuilderContainer = propsdbBuilderContainer

		imageNameParts := strings.Split(imageName, ":")
		fromDockerfile := testcontainers.FromDockerfile{
			Context:    buildContext,
			Dockerfile: "Dockerfile",
			Repo:       imageNameParts[0],
			Tag:        imageNameParts[1],
			KeepImage:  true, // Keep the image so we can reuse it
			BuildArgs:  propsdbBuildArgs,
			BuildOptionsModifier: func(opts *build.ImageBuildOptions) {
				opts.Target = "runtime"
			},
			PrintBuildLog: true,
		}

		propsdbContainerRequest.FromDockerfile = fromDockerfile
	} else {
		// Add Image to PropsDB container request to reuse the existing image
		logMessage(t, "Image %s exists, reusing...", imageName)
		propsdbContainerRequest.Image = imageName
	}

	// Create and start the PropsDB container
	propsdbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: propsdbContainerRequest,
		Started:          true,
	})
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, "Failed to start PropsDB")
	}
	testContainers.PropsDBContainer = propsdbContainer

	// Log the localhost and mapped ports for PropsDB
	propsdbHost, _ := propsdbContainer.Host(ctx)
	propsdbPort, _ := propsdbContainer.MappedPort(ctx, tcpPropsdbPort)
	logMessage(t, "BASE_URL=%s:%s", propsdbHost, propsdbPort.Port())

	logMessage(t, "PropsDB testcontainer started successfully")
	return testContainers, nil
}

func getDBInitEnvMap(dbType string) map[string]string {
	switch dbType {
	case "postgres":
		return map[string]string{
			"POSTGRES_PASSWORD": os.Getenv("DB_APP_PASSWORD"),
			"POSTGRES_USER":     os.Getenv("DB_APP_USER"),
			"POSTGRES_DB":       os.Getenv("DB_APP_DATABASE"),
		}
	default:
	case "mariadb", "mysql":
		return map[string]string{
			"MYSQL_ROOT_PASSWORD": os.Getenv("DB_ROOT_PASSWORD"),
			"MYSQL_DATABASE":      os.Getenv("DB_APP_DATABASE"),
			"MYSQL_USER":          os.Getenv("DB_APP_USER"),
			"MYSQL_PASSWORD":      os.Getenv("DB_APP_PASSWORD"),
		}
	}
	return nil
}

func performMySqlDBInit(t *testing.T, testContainers *TestContainers, dbHost string, dbPort nat.Port) error {
	db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(%s:%s)/", os.Getenv("DB_ROOT_PASSWORD"), dbHost, dbPort.Port()))
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, "Failed to connect to MariaDB for setup")
	}
	defer db.Close()

	// Wait for connection to be really ready
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, "MariaDB not ready after 30 seconds")
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", os.Getenv("DB_APP_DATABASE")))
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, fmt.Sprintf("Failed to create %s", os.Getenv("DB_APP_DATABASE")))
	}
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", os.Getenv("AUTHZ_DATABASE")))
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, fmt.Sprintf("Failed to create %s", os.Getenv("AUTHZ_DATABASE")))
	}
	_, err = db.Exec(fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")))
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, fmt.Sprintf("Failed to create user %s", os.Getenv("DB_USER")))
	}
	_, err = db.Exec(fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%%' IDENTIFIED BY '%s'", os.Getenv("DB_APP_USER"), os.Getenv("DB_APP_PASSWORD")))
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, fmt.Sprintf("Failed to create user %s", os.Getenv("DB_APP_USER")))
	}
	_, err = db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.authorizer_users (id CHAR(36) NOT NULL PRIMARY KEY)", os.Getenv("AUTHZ_DATABASE")))
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, fmt.Sprintf("Failed to create table authorizer_users in %s", os.Getenv("AUTHZ_DATABASE")))
	}
	_, err = db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON *.* TO 'root'@'%%' IDENTIFIED BY '%s' WITH GRANT OPTION", os.Getenv("DB_ROOT_PASSWORD")))
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, fmt.Sprintf("Failed to grant privileges for %s", os.Getenv("DB_APP_DATABASE")))
	}
	_, err = db.Exec("FLUSH PRIVILEGES")
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, fmt.Sprintf("Failed to flush privileges for %s", os.Getenv("DB_APP_DATABASE")))
	}
	err = executeSQL(db, data.InitdbMariaDBTables)
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, fmt.Sprintf("Failed to execute %s tables init sql", os.Getenv("DB_TYPE")))
	}
	err = executeSQL(db, data.InitdbMariaDBPrivileges)
	if err != nil {
		testContainers.Terminate(t)
		exitWithError(t, err, fmt.Sprintf("Failed to execute %s privileges init sql", os.Getenv("DB_TYPE")))
	}

	return nil
}

func performPostgresDBInit(_ *testing.T, _ *TestContainers, _ string, _ nat.Port) error {
	return fmt.Errorf("Postgres not fully supported yet")
}

func executeSQL(db *sql.DB, sql string) error {
	lines := strings.Split(sql, "\n")

	var ncls []string
	for _, l := range lines {
		ncl := excludeComment(l)
		ncls = append(ncls, ncl)
	}

	l := strings.Join(ncls, "")
	queries := strings.Split(l, ";")
	queries = queries[:len(queries)-1]

	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			return fmt.Errorf("%s : when executing > %s", err.Error(), q)
		}
	}
	return nil
}

func excludeComment(line string) string {
	d := "\""
	s := "'"
	c := "--"

	var nc string
	ck := line
	mx := len(line) + 1

	for {
		if len(ck) == 0 {
			return nc
		}

		di := strings.Index(ck, d)
		si := strings.Index(ck, s)
		ci := strings.Index(ck, c)

		if di < 0 {
			di = mx
		}
		if si < 0 {
			si = mx
		}
		if ci < 0 {
			ci = mx
		}

		var ei int

		if di < si && di < ci {
			nc += ck[:di+1]
			ck = ck[di+1:]
			ei = strings.Index(ck, d)
		} else if si < di && si < ci {
			nc += ck[:si+1]
			ck = ck[si+1:]
			ei = strings.Index(ck, s)
		} else if ci < di && ci < si {
			return nc + ck[:ci]
		} else {
			return nc + ck
		}

		nc += ck[:ei+1]
		ck = ck[ei+1:]
	}
}

func imageExists(ctx context.Context, imageName string) (bool, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}
	defer cli.Close()

	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}

	return false, nil
}

func exitWithError(t *testing.T, err error, msg string) {
	if t != nil {
		t.Fatalf(msg+": %v", err)
	} else {
		fmt.Printf(msg+": %v\n", err)
		os.Exit(1)
	}
}

func logMessage(t *testing.T, format string, args ...any) {
	if t != nil {
		t.Logf(format, args...)
	} else {
		fmt.Printf(format+"\n", args...)
	}
}
