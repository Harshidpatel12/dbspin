package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

type DBConfig struct {
	DefaultPort int
	Image       string
	Env         []string
	Volume      string
	ConnString  func(port int) string
}

var dbConfigs = map[string]DBConfig{
	"postgres": {
		DefaultPort: 5432,
		Image:       "postgres:16-alpine",
		Env:         []string{"POSTGRES_PASSWORD=postgres", "POSTGRES_USER=postgres", "POSTGRES_DB=postgres"},
		Volume:      "/var/lib/postgresql/data",
		ConnString: func(port int) string {
			return fmt.Sprintf("postgresql://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)
		},
	},
	"redis": {
		DefaultPort: 6379,
		Image:       "redis:7-alpine",
		Env:         []string{},
		Volume:      "/data",
		ConnString: func(port int) string {
			return fmt.Sprintf("redis://localhost:%d", port)
		},
	},
	"mysql": {
		DefaultPort: 3306,
		Image:       "mysql:8.0-oracle",
		Env:         []string{"MYSQL_ROOT_PASSWORD=mysql", "MYSQL_DATABASE=mysql"},
		Volume:      "/var/lib/mysql",
		ConnString: func(port int) string {
			return fmt.Sprintf("mysql://root:mysql@tcp(localhost:%d)/mysql", port)
		},
	},
	"mongo": {
		DefaultPort: 27017,
		Image:       "mongo:7.0",
		Env:         []string{"MONGO_INITDB_ROOT_USERNAME=mongo", "MONGO_INITDB_ROOT_PASSWORD=mongo"},
		Volume:      "/data/db",
		ConnString: func(port int) string {
			return fmt.Sprintf("mongodb://mongo:mongo@localhost:%d/?authSource=admin", port)
		},
	},
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Verify Docker is installed and running
	checkDockerInstalled()

	switch command {
	case "up":
		handleUp(os.Args[2:])
	case "down":
		handleDown(os.Args[2:])
	case "list":
		handleList()
	case "logs":
		handleLogs(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("dbspin — Instant Database Spinner")
	fmt.Println("\nUsage:")
	fmt.Println("  dbspin up <engine> [flags]   Spin up a database container (postgres, redis, mysql, mongo)")
	fmt.Println("  dbspin down <engine>         Stop and remove a running database container")
	fmt.Println("  dbspin list                  List all running databases managed by dbspin")
	fmt.Println("  dbspin logs <engine>         Show logs for a database container")
	fmt.Println("\nFlags for 'up' command:")
	fmt.Println("  -port int                     Override default host port")
	fmt.Println("  -version string               Override default docker image tag/version")
}

func checkDockerInstalled() {
	_, err := exec.LookPath("docker")
	if err != nil {
		fmt.Println("❌ Error: 'docker' CLI is not installed or not found in PATH.")
		os.Exit(1)
	}

	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Error: Docker daemon is not running. Please start Docker and try again.")
		os.Exit(1)
	}
}

func isPortFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

func findFreePort(startPort int) int {
	port := startPort
	for {
		if isPortFree(port) {
			return port
		}
		port++
	}
}

func containerExists(name string) (bool, string) {
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", name)
	outputBytes, err := cmd.Output()
	if err != nil {
		return false, ""
	}
	return true, strings.TrimSpace(string(outputBytes))
}

func handleUp(args []string) {
	if len(args) < 1 {
		fmt.Println("❌ Error: Missing database engine name. E.g. 'dbspin up postgres'")
		os.Exit(1)
	}

	engine := args[0]
	config, ok := dbConfigs[engine]
	if !ok {
		fmt.Printf("❌ Error: Unknown database engine '%s'. Supported engines: postgres, redis, mysql, mongo\n", engine)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("up", flag.ExitOnError)
	portOpt := fs.Int("port", 0, "Override default host port")
	versionOpt := fs.String("version", "", "Override default docker image tag/version")
	_ = fs.Parse(args[1:])

	containerName := fmt.Sprintf("dbspin-%s", engine)
	volumeName := fmt.Sprintf("dbspin-%s-data", engine)

	// Check if container already exists
	exists, status := containerExists(containerName)
	if exists {
		if status == "running" {
			fmt.Printf("⚠️  Database '%s' is already running.\n", engine)
			return
		}
		fmt.Printf("🔄 Database '%s' exists but is stopped. Starting it...\n", engine)
		startCmd := exec.Command("docker", "start", containerName)
		if err := startCmd.Run(); err != nil {
			fmt.Printf("❌ Error: Failed to start container '%s': %v\n", containerName, err)
			os.Exit(1)
		}
		fmt.Printf("✅ Started database '%s' successfully!\n", engine)
		return
	}

	// Port allocation
	hostPort := config.DefaultPort
	if *portOpt != 0 {
		hostPort = *portOpt
		if !isPortFree(hostPort) {
			fmt.Printf("❌ Error: Port %d is already in use.\n", hostPort)
			os.Exit(1)
		}
	} else {
		hostPort = findFreePort(hostPort)
		if hostPort != config.DefaultPort {
			fmt.Printf("⚠️  Default port %d is busy. Auto-bound to next free port: %d\n", config.DefaultPort, hostPort)
		}
	}

	// Resolve image version
	image := config.Image
	if *versionOpt != "" {
		parts := strings.Split(config.Image, ":")
		image = fmt.Sprintf("%s:%s", parts[0], *versionOpt)
	}

	fmt.Printf("🚀 Spinning up %s database...\n", engine)
	fmt.Printf("➡️  Pulling Docker image '%s' (if not present)...\n", image)

	// Build Docker Run arguments
	runArgs := []string{
		"run", "-d",
		"--name", containerName,
		"-p", fmt.Sprintf("%d:%d", hostPort, config.DefaultPort),
		"-v", fmt.Sprintf("%s:%s", volumeName, config.Volume),
	}

	// Add Env variables
	for _, env := range config.Env {
		runArgs = append(runArgs, "-e", env)
	}

	runArgs = append(runArgs, image)

	// Execute docker run
	runCmd := exec.Command("docker", runArgs...)
	if output, err := runCmd.CombinedOutput(); err != nil {
		fmt.Printf("❌ Error running container: %v\n%s\n", err, string(output))
		os.Exit(1)
	}

	fmt.Printf("✅ Container '%s' created successfully!\n", containerName)
	fmt.Printf("\n%-12s %d (Host) -> %d (Container)\n", "Port Mapping:", hostPort, config.DefaultPort)
	fmt.Printf("%-12s %s\n", "Connection:", config.ConnString(hostPort))
}

func handleDown(args []string) {
	if len(args) < 1 {
		fmt.Println("❌ Error: Missing database engine name. E.g. 'dbspin down postgres'")
		os.Exit(1)
	}

	engine := args[0]
	containerName := fmt.Sprintf("dbspin-%s", engine)

	exists, _ := containerExists(containerName)
	if !exists {
		fmt.Printf("❌ Error: Database '%s' does not exist.\n", engine)
		return
	}

	fmt.Printf("🛑 Stopping database '%s'...\n", engine)
	stopCmd := exec.Command("docker", "stop", containerName)
	_ = stopCmd.Run()

	fmt.Printf("🗑️  Removing container '%s'...\n", containerName)
	rmCmd := exec.Command("docker", "rm", containerName)
	if err := rmCmd.Run(); err != nil {
		fmt.Printf("❌ Error removing container: %v\n", err)
		return
	}

	fmt.Printf("✅ Database '%s' removed successfully!\n", engine)
}

func handleList() {
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name=dbspin-", "--format", "table {{.Names}}\t{{.Image}}\t{{.Ports}}\t{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Error listing databases: %v\n", err)
		os.Exit(1)
	}

	outputStr := string(output)
	if len(strings.TrimSpace(outputStr)) <= 10 { // empty table header
		fmt.Println("No active databases found. Run 'dbspin up <engine>' to spin one up!")
		return
	}

	fmt.Println(outputStr)
}

func handleLogs(args []string) {
	if len(args) < 1 {
		fmt.Println("❌ Error: Missing database engine name. E.g. 'dbspin logs postgres'")
		os.Exit(1)
	}

	engine := args[0]
	containerName := fmt.Sprintf("dbspin-%s", engine)

	exists, _ := containerExists(containerName)
	if !exists {
		fmt.Printf("❌ Error: Database '%s' does not exist.\n", engine)
		return
	}

	fmt.Printf("📋 Streaming logs for '%s' (Ctrl+C to exit)...\n\n", containerName)
	cmd := exec.Command("docker", "logs", "-f", containerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
