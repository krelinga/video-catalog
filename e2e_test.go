package videocatalog

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestVideoCatalogEnd2End(t *testing.T) {
	ctx := context.Background()

	setup(t, ctx)
}

// Starts the server container and all dependencies, and returns a URL string that can be used in client connections.
func setup(t *testing.T, ctx context.Context) string {
	// Create docker network.
	net, err := network.New(ctx, network.WithCheckDuplicate())
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	networkName := net.Name

	// Database configuration
	const (
		dbHost = "postgres"
		dbPort = "5432"
		dbName = "videocatalog"
		dbUser = "videocataloguser"
		dbPass = "videocatalogpass"
	)

	// Start Postgres container
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPass,
		},
		Networks:       []string{networkName},
		NetworkAliases: map[string][]string{networkName: {dbHost}},
		WaitingFor:     wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		dumpContainerLogs(t, ctx, postgresContainer, dbHost)
	})

	// Build and start the server container
	serverReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    ".",
			Dockerfile: "Dockerfile",
			BuildArgs:  map[string]*string{},
		},
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"VC_SERVER_PORT": "8080",
			"VC_DB_HOST":     dbHost,
			"VC_DB_PORT":     dbPort,
			"VC_DB_NAME":     dbName,
			"VC_DB_USER":     dbUser,
			"VC_DB_PASSWORD": dbPass,
		},
		Networks:       []string{networkName},
		NetworkAliases: map[string][]string{networkName: {"server"}},
		WaitingFor:     wait.ForLog("Starting HTTP server on port 8080"),
	}
	serverContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: serverReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start server container: %v", err)
	}
	t.Cleanup(func() {
		dumpContainerLogs(t, ctx, serverContainer, "server")
	})

	// Get server mapped port
	mappedPort, err := serverContainer.MappedPort(ctx, "8080")
	if err != nil {
		t.Fatalf("failed to get server mapped port: %v", err)
	}
	serverHost, err := serverContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get server host: %v", err)
	}

	return fmt.Sprintf("http://%s:%s", serverHost, mappedPort.Port())
}

// dumpContainerLogs reads and logs all output from a container
func dumpContainerLogs(t *testing.T, ctx context.Context, container testcontainers.Container, name string) {
	logs, err := container.Logs(ctx)
	if err != nil {
		t.Logf("failed to get %s container logs: %v", name, err)
		return
	}
	defer logs.Close()

	logBytes, err := io.ReadAll(logs)
	if err != nil {
		t.Logf("failed to read %s container logs: %v", name, err)
		return
	}

	t.Logf("=== %s container logs ===\n%s", name, string(logBytes))
}
