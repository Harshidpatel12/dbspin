package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"
)

type GuiConfig struct {
	Image       string
	DefaultPort int
	Env         func(dbPort int) []string
	UrlTemplate func(guiPort int) string
}

type DBConfig struct {
	DefaultPort int
	Image       string
	Env         []string
	Volume      string
	ConnString  func(port int) string
	GuiConfig   *GuiConfig
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
		GuiConfig: &GuiConfig{
			Image:       "adminer:latest",
			DefaultPort: 8080,
			Env:         func(dbPort int) []string { return []string{} },
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d/?server=host.docker.internal&username=postgres&db=postgres", guiPort)
			},
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
		GuiConfig: &GuiConfig{
			Image:       "rediscommander/redis-commander:latest",
			DefaultPort: 8081,
			Env: func(dbPort int) []string {
				return []string{
					fmt.Sprintf("REDIS_HOSTS=local:host.docker.internal:%d", dbPort),
				}
			},
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d", guiPort)
			},
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
		GuiConfig: &GuiConfig{
			Image:       "adminer:latest",
			DefaultPort: 8080,
			Env:         func(dbPort int) []string { return []string{} },
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d/?server=host.docker.internal&username=root&db=mysql", guiPort)
			},
		},
	},
	"mariadb": {
		DefaultPort: 3306,
		Image:       "mariadb:11",
		Env:         []string{"MARIADB_ROOT_PASSWORD=mariadb", "MARIADB_DATABASE=mariadb"},
		Volume:      "/var/lib/mysql",
		ConnString: func(port int) string {
			return fmt.Sprintf("mysql://root:mariadb@tcp(localhost:%d)/mariadb", port)
		},
		GuiConfig: &GuiConfig{
			Image:       "adminer:latest",
			DefaultPort: 8080,
			Env:         func(dbPort int) []string { return []string{} },
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d/?server=host.docker.internal&username=root&db=mariadb", guiPort)
			},
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
		GuiConfig: &GuiConfig{
			Image: "mongo-express:latest",
			DefaultPort: 8081,
			Env: func(dbPort int) []string {
				return []string{
					"ME_CONFIG_MONGODB_SERVER=host.docker.internal",
					fmt.Sprintf("ME_CONFIG_MONGODB_PORT=%d", dbPort),
					"ME_CONFIG_BASICAUTH_USERNAME=admin",
					"ME_CONFIG_BASICAUTH_PASSWORD=pass",
				}
			},
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d (Username: admin, Password: pass)", guiPort)
			},
		},
	},
	"elasticsearch": {
		DefaultPort: 9200,
		Image:       "elasticsearch:8.12.2",
		Env:         []string{"discovery.type=single-node", "xpack.security.enabled=false"},
		Volume:      "/usr/share/elasticsearch/data",
		ConnString: func(port int) string {
			return fmt.Sprintf("http://localhost:%d", port)
		},
		GuiConfig: &GuiConfig{
			Image:       "mherman/elasticvue:latest",
			DefaultPort: 8080,
			Env:         func(dbPort int) []string { return []string{} },
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d", guiPort)
			},
		},
	},
	"rabbitmq": {
		DefaultPort: 5672,
		Image:       "rabbitmq:3-management-alpine",
		Env:         []string{"RABBITMQ_DEFAULT_USER=guest", "RABBITMQ_DEFAULT_PASS=guest"},
		Volume:      "/var/lib/rabbitmq",
		ConnString: func(port int) string {
			return fmt.Sprintf("amqp://guest:guest@localhost:%d/", port)
		},
		GuiConfig: &GuiConfig{
			Image:       "",
			DefaultPort: 15672,
			Env:         func(dbPort int) []string { return []string{} },
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d (Username: guest, Password: guest)", guiPort)
			},
		},
	},
	"pgvector": {
		DefaultPort: 5432,
		Image:       "pgvector/pgvector:16",
		Env:         []string{"POSTGRES_PASSWORD=postgres", "POSTGRES_USER=postgres", "POSTGRES_DB=postgres"},
		Volume:      "/var/lib/postgresql/data",
		ConnString: func(port int) string {
			return fmt.Sprintf("postgresql://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)
		},
		GuiConfig: &GuiConfig{
			Image:       "adminer:latest",
			DefaultPort: 8080,
			Env:         func(dbPort int) []string { return []string{} },
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d/?server=host.docker.internal&username=postgres&db=postgres", guiPort)
			},
		},
	},
	"timescaledb": {
		DefaultPort: 5432,
		Image:       "timescale/timescaledb:latest-pg16",
		Env:         []string{"POSTGRES_PASSWORD=postgres", "POSTGRES_USER=postgres", "POSTGRES_DB=postgres"},
		Volume:      "/var/lib/postgresql/data",
		ConnString: func(port int) string {
			return fmt.Sprintf("postgresql://postgres:postgres@localhost:%d/postgres?sslmode=disable", port)
		},
		GuiConfig: &GuiConfig{
			Image:       "adminer:latest",
			DefaultPort: 8080,
			Env:         func(dbPort int) []string { return []string{} },
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d/?server=host.docker.internal&username=postgres&db=postgres", guiPort)
			},
		},
	},
	"kafka": {
		DefaultPort: 9092,
		Image:       "apache/kafka:3.7.0",
		Env: []string{
			"KAFKA_NODE_ID=1",
			"KAFKA_PROCESS_ROLES=broker,controller",
			"KAFKA_LISTENERS=PLAINTEXT://:9092,CONTROLLER://:9093",
			"KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092",
			"KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093",
			"KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
		},
		Volume: "/var/lib/kafka/data",
		ConnString: func(port int) string {
			return fmt.Sprintf("localhost:%d", port)
		},
		GuiConfig: &GuiConfig{
			Image:       "provectus/kafka-ui:latest",
			DefaultPort: 8080,
			Env: func(dbPort int) []string {
				return []string{
					"KAFKA_CLUSTERS_0_NAME=local",
					fmt.Sprintf("KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS=host.docker.internal:%d", dbPort),
				}
			},
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d", guiPort)
			},
		},
	},
	"meilisearch": {
		DefaultPort: 7700,
		Image:       "getmeili/meilisearch:v1.6",
		Env:         []string{"MEILI_NO_ANALYTICS=true", "MEILI_MASTER_KEY=masterKey"},
		Volume:      "/data.ms",
		ConnString: func(port int) string {
			return fmt.Sprintf("http://localhost:%d (Master Key: masterKey)", port)
		},
	},
	"localstack": {
		DefaultPort: 4566,
		Image:       "localstack/localstack:latest",
		Env:         []string{"SERVICES=sqs,sns,s3,dynamodb"},
		Volume:      "/var/lib/localstack",
		ConnString: func(port int) string {
			return fmt.Sprintf("http://localhost:%d", port)
		},
	},
	"minio": {
		DefaultPort: 9000,
		Image:       "minio/minio:latest",
		Env:         []string{"MINIO_ROOT_USER=minioadmin", "MINIO_ROOT_PASSWORD=minioadmin"},
		Volume:      "/data",
		ConnString: func(port int) string {
			return fmt.Sprintf("http://localhost:%d", port)
		},
		GuiConfig: &GuiConfig{
			Image:       "",
			DefaultPort: 9001,
			Env:         func(dbPort int) []string { return []string{} },
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d (Username: minioadmin, Password: minioadmin)", guiPort)
			},
		},
	},
	"dynamodb": {
		DefaultPort: 8000,
		Image:       "amazon/dynamodb-local:latest",
		Env:         []string{},
		Volume:      "/data",
		ConnString: func(port int) string {
			return fmt.Sprintf("http://localhost:%d", port)
		},
		GuiConfig: &GuiConfig{
			Image:       "aaronshaf/dynamodb-admin:latest",
			DefaultPort: 8001,
			Env: func(dbPort int) []string {
				return []string{
					fmt.Sprintf("DYNAMO_ENDPOINT=http://host.docker.internal:%d", dbPort),
				}
			},
			UrlTemplate: func(guiPort int) string {
				return fmt.Sprintf("http://localhost:%d", guiPort)
			},
		},
	},
}

func setupAutocomplete() {
	if os.Getuid() == 0 || os.Getenv("SUDO_USER") != "" {
		fmt.Fprintln(os.Stderr, "Error: Cannot run autocomplete setup as root or via sudo to avoid permission conflicts.")
		os.Exit(1)
	}

	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		fmt.Fprintln(os.Stderr, "Error: Could not detect active shell (SHELL environment variable is empty).")
		os.Exit(1)
	}
	shell := filepath.Base(shellPath)

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
		os.Exit(1)
	}

	switch shell {
	case "bash":
		compFile := filepath.Join(home, ".dbspin_completion")
		if _, err := os.Stat(compFile); err == nil {
			fmt.Printf("* Autocomplete is already configured for bash (~/.dbspin_completion)\n")
			return
		}
		err = os.WriteFile(compFile, []byte(bashCompletionScript), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing completion file: %v\n", err)
			os.Exit(1)
		}
		bashrcPath := filepath.Join(home, ".bashrc")
		hasLine := false
		if data, err := os.ReadFile(bashrcPath); err == nil {
			if strings.Contains(string(data), ".dbspin_completion") {
				hasLine = true
			}
		}
		if !hasLine {
			f, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err == nil {
				_, _ = f.WriteString("\n# dbspin auto-completion\n[ -f ~/.dbspin_completion ] && source ~/.dbspin_completion\n")
				_ = f.Close()
			}
		}
		fmt.Printf("* Auto-configured autocomplete for bash (sourced in ~/.bashrc)\n")

	case "zsh":
		compDir := filepath.Join(home, ".zsh", "completion")
		compFile := filepath.Join(compDir, "_dbspin")
		if _, err := os.Stat(compFile); err == nil {
			fmt.Printf("* Autocomplete is already configured for zsh (%s)\n", compFile)
			return
		}
		_ = os.MkdirAll(compDir, 0755)
		err = os.WriteFile(compFile, []byte(zshCompletionScript), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing completion file: %v\n", err)
			os.Exit(1)
		}
		zshrcPath := filepath.Join(home, ".zshrc")
		hasLine := false
		if data, err := os.ReadFile(zshrcPath); err == nil {
			if strings.Contains(string(data), ".zsh/completion") {
				hasLine = true
			}
		}
		if !hasLine {
			f, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err == nil {
				_, _ = f.WriteString("\n# dbspin auto-completion\nfpath=(~/.zsh/completion $fpath)\nautoload -Uz compinit && compinit\n")
				_ = f.Close()
			}
		}
		fmt.Printf("* Auto-configured autocomplete for zsh (configured in ~/.zshrc)\n")

	case "fish":
		compDir := filepath.Join(home, ".config", "fish", "completions")
		compFile := filepath.Join(compDir, "dbspin.fish")
		if _, err := os.Stat(compFile); err == nil {
			fmt.Printf("* Autocomplete is already configured for fish (%s)\n", compFile)
			return
		}
		_ = os.MkdirAll(compDir, 0755)
		err = os.WriteFile(compFile, []byte(fishCompletionScript), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing completion file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("* Auto-configured autocomplete for fish\n")

	default:
		fmt.Fprintf(os.Stderr, "Error: Auto-setup not supported for shell '%s'. Please configure manually.\n", shell)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		handleInteractive()
		return
	}

	command := os.Args[1]

	// Verify Docker is installed and running
	checkDockerInstalled()

	switch command {
	case "up":
		if len(os.Args) < 3 {
			handleInteractive()
		} else {
			handleUp(os.Args[2:])
		}
	case "down":
		handleDown(os.Args[2:])
	case "list":
		handleList(os.Args[2:])
	case "logs":
		handleLogs(os.Args[2:])
	case "shell":
		handleShell(os.Args[2:])
	case "export":
		handleExport(os.Args[2:])
	case "import":
		handleImport(os.Args[2:])
	case "info":
		handleInfo(os.Args[2:])
	case "prune":
		handlePrune()
	case "compose":
		handleCompose(os.Args[2:])
	case "completion":
		handleCompletion(os.Args[2:])
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
	fmt.Println("  dbspin up <engine> [flags]   Spin up a database container (postgres, redis, mysql, mariadb, mongo, elasticsearch, rabbitmq, pgvector, timescaledb, kafka, meilisearch, localstack, minio, dynamodb)")
	fmt.Println("  dbspin down <container>       Stop and remove a running database or container")
	fmt.Println("  dbspin list [flags]          List all running databases managed by dbspin")
	fmt.Println("  dbspin logs <container>       Show logs for a database or companion container")
	fmt.Println("  dbspin shell <container>      Connect to database container via an interactive CLI client")
	fmt.Println("  dbspin export <container>     Export database contents (redirect using > output.sql or use -f)")
	fmt.Println("  dbspin import <container>     Import database contents (redirect using < input.sql or use -f)")
	fmt.Println("  dbspin info <container>       Print database connection parameters, credentials, and web URLs")
	fmt.Println("  dbspin prune                  Stop and remove ALL dbspin containers and volumes")
	fmt.Println("  dbspin compose [engines...]  Generate a docker-compose.yml configuration for specified engines")
	fmt.Println("  dbspin completion [shell]     Generate shell completions (bash, zsh, fish)")
	fmt.Println("\nFlags for 'up' command:")
	fmt.Println("  -port int                     Override default host port")
	fmt.Println("  -version string               Override default docker image tag/version")
	fmt.Println("  -name string                  Override default container and volume name suffix")
	fmt.Println("  -env string                   Additional environment variables comma-separated (e.g. KEY1=VAL1,KEY2=VAL2)")
	fmt.Println("  -seed string                  Mount a seed file into the container initialization folder")
	fmt.Println("  -gui                          Start a companion web GUI dashboard for the database")
	fmt.Println("  -wait                         Wait for the database to be fully ready before returning")
	fmt.Println("\nFlags for 'list' command:")
	fmt.Println("  -v, --verbose                 Show connection strings and active GUI dashboard URLs")
	fmt.Println("\nFlags for 'compose' command:")
	fmt.Println("  -gui                          Include companion GUI dashboard services in the compose file")
}

func checkDockerInstalled() {
	_, err := exec.LookPath("docker")
	if err != nil {
		fmt.Println("Error: 'docker' CLI is not installed or not found in PATH.")
		os.Exit(1)
	}

	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		fmt.Println("Error: Docker daemon is not running. Please start Docker and try again.")
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
		fmt.Println("Error: Missing database engine name. E.g. 'dbspin up postgres'")
		os.Exit(1)
	}

	engine := args[0]
	config, ok := dbConfigs[engine]
	if !ok {
		fmt.Printf("Error: Unknown database engine '%s'. Supported engines: postgres, redis, mysql, mariadb, mongo, elasticsearch, rabbitmq, pgvector, timescaledb, kafka, meilisearch, localstack, minio, dynamodb\n", engine)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("up", flag.ExitOnError)
	portOpt := fs.Int("port", 0, "Override default host port")
	versionOpt := fs.String("version", "", "Override default docker image tag/version")
	nameOpt := fs.String("name", "", "Override default container and volume name suffix")
	envOpt := fs.String("env", "", "Additional environment variables comma-separated (e.g. KEY1=VAL1,KEY2=VAL2)")
	seedOpt := fs.String("seed", "", "Mount a seed file into the container initialization folder")
	guiOpt := fs.Bool("gui", false, "Start a companion web GUI dashboard for the database")
	waitOpt := fs.Bool("wait", false, "Wait for the database to be fully initialized and ready to accept connections")
	_ = fs.Parse(args[1:])

	suffix := engine
	if *nameOpt != "" {
		suffix = fmt.Sprintf("%s-%s", engine, *nameOpt)
	}

	containerName := fmt.Sprintf("dbspin-%s", suffix)
	volumeName := fmt.Sprintf("dbspin-%s-data", suffix)

	// Check if container already exists
	exists, status := containerExists(containerName)
	if exists {
		if status == "running" {
			fmt.Printf("Warning: database '%s' is already running.\n", suffix)
			if *guiOpt && config.GuiConfig != nil {
				startGui(containerName, config.DefaultPort, config.GuiConfig, true)
			}
			return
		}
		fmt.Printf("* Database '%s' exists but is stopped. Starting it...\n", suffix)
		startCmd := exec.Command("docker", "start", containerName)
		if err := startCmd.Run(); err != nil {
			fmt.Printf("Error: Failed to start container '%s': %v\n", containerName, err)
			os.Exit(1)
		}
		fmt.Printf("* Started database '%s' successfully!\n", suffix)
		hostPort := getContainerMappedPort(containerName, config.DefaultPort)
		if *waitOpt {
			waitForContainerReady(containerName, engine, hostPort)
		}
		if *guiOpt && config.GuiConfig != nil {
			startGui(containerName, hostPort, config.GuiConfig, false)
		}
		return
	}

	// Port allocation
	hostPort := config.DefaultPort
	if *portOpt != 0 {
		hostPort = *portOpt
		if !isPortFree(hostPort) {
			fmt.Printf("Error: Port %d is already in use.\n", hostPort)
			os.Exit(1)
		}
	} else {
		hostPort = findFreePort(hostPort)
		if hostPort != config.DefaultPort {
			fmt.Printf("Warning: Default port %d is busy. Auto-bound to next free port: %d\n", config.DefaultPort, hostPort)
		}
	}

	// Resolve image version
	image := config.Image
	if *versionOpt != "" {
		parts := strings.Split(config.Image, ":")
		image = fmt.Sprintf("%s:%s", parts[0], *versionOpt)
	}

	var mgmtPort int
	if (engine == "rabbitmq" || engine == "minio") && *guiOpt && config.GuiConfig != nil {
		mgmtPort = findFreePort(config.GuiConfig.DefaultPort)
	}

	fmt.Printf("* Spinning up %s database...\n", engine)
	fmt.Printf("  --> Pulling Docker image '%s' (if not present)...\n", image)

	// Build Docker Run arguments
	runArgs := []string{
		"run", "-d",
		"--name", containerName,
		"-p", fmt.Sprintf("%d:%d", hostPort, config.DefaultPort),
		"-v", fmt.Sprintf("%s:%s", volumeName, config.Volume),
	}

	if mgmtPort != 0 {
		runArgs = append(runArgs, "-p", fmt.Sprintf("%d:%d", mgmtPort, config.GuiConfig.DefaultPort))
	}

	// Add Env variables from config (with dynamic substitutions for Kafka ports)
	for _, env := range config.Env {
		if engine == "kafka" && strings.HasPrefix(env, "KAFKA_ADVERTISED_LISTENERS=") {
			runArgs = append(runArgs, "-e", fmt.Sprintf("KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:%d", hostPort))
		} else {
			runArgs = append(runArgs, "-e", env)
		}
	}

	// Add user custom Env variables
	if *envOpt != "" {
		pairs := strings.Split(*envOpt, ",")
		for _, pair := range pairs {
			if strings.Contains(pair, "=") {
				runArgs = append(runArgs, "-e", pair)
			}
		}
	}

	// Handle seed mounting for SQL/JSON/JS files or directories
	if *seedOpt != "" {
		absPath, err := filepath.Abs(*seedOpt)
		if err != nil {
			absPath = *seedOpt
		}
		baseName := filepath.Base(absPath)
		switch engine {
		case "postgres", "pgvector", "timescaledb", "mysql", "mariadb", "mongo":
			fi, err := os.Stat(absPath)
			if err == nil && fi.IsDir() {
				runArgs = append(runArgs, "-v", fmt.Sprintf("%s:/docker-entrypoint-initdb.d:ro", absPath))
				fmt.Printf("  --> Mounting seed directory: %s\n", absPath)
			} else {
				runArgs = append(runArgs, "-v", fmt.Sprintf("%s:/docker-entrypoint-initdb.d/%s:ro", absPath, baseName))
				fmt.Printf("  --> Mounting seed file: %s\n", absPath)
			}
		default:
			fmt.Printf("Warning: Seeding is not natively supported for '%s' containers.\n", engine)
		}
	}

	runArgs = append(runArgs, image)

	// Add minio extra console arguments
	if engine == "minio" {
		runArgs = append(runArgs, "server", "/data", "--console-address", ":9001")
	}

	// Execute docker run
	runCmd := exec.Command("docker", runArgs...)
	if output, err := runCmd.CombinedOutput(); err != nil {
		fmt.Printf("Error running container: %v\n%s\n", err, string(output))
		os.Exit(1)
	}

	fmt.Printf("* Container '%s' created successfully!\n", containerName)
	if *waitOpt {
		waitForContainerReady(containerName, engine, hostPort)
	}
	fmt.Printf("\n%-12s %d (Host) -> %d (Container)\n", "Port Mapping:", hostPort, config.DefaultPort)
	fmt.Printf("%-12s %s\n", "Connection:", config.ConnString(hostPort))

	if *guiOpt && config.GuiConfig != nil {
		if config.GuiConfig.Image == "" {
			fmt.Printf("%-12s %s\n", "GUI Dashboard:", config.GuiConfig.UrlTemplate(mgmtPort))
		} else {
			startGui(containerName, hostPort, config.GuiConfig, false)
		}
	}
}

func waitForContainerReady(containerName, engine string, hostPort int) {
	fmt.Fprintf(os.Stderr, "* Waiting for database '%s' to be ready...\n", engine)

	maxAttempts := 30
	sleepInterval := time.Second

	for i := 1; i <= maxAttempts; i++ {
		ready := false
		var err error

		switch engine {
		case "postgres", "pgvector", "timescaledb":
			cmd := exec.Command("docker", "exec", containerName, "pg_isready", "-U", "postgres", "-h", "localhost")
			err = cmd.Run()
			ready = (err == nil)

		case "redis":
			cmd := exec.Command("docker", "exec", containerName, "redis-cli", "ping")
			output, err2 := cmd.Output()
			ready = (err2 == nil && strings.Contains(strings.ToLower(string(output)), "pong"))

		case "mysql", "mariadb":
			cmd := exec.Command("docker", "exec", containerName, "mysqladmin", "ping", "-u", "root", "-proot")
			err = cmd.Run()
			ready = (err == nil)

		case "mongo":
			cmd := exec.Command("docker", "exec", containerName, "mongosh", "--eval", "db.adminCommand('ping')")
			if err2 := cmd.Run(); err2 == nil {
				ready = true
			} else {
				cmdLegacy := exec.Command("docker", "exec", containerName, "mongo", "--eval", "db.adminCommand('ping')")
				ready = (cmdLegacy.Run() == nil)
			}

		case "rabbitmq":
			cmd := exec.Command("docker", "exec", containerName, "rabbitmq-diagnostics", "-q", "check_running")
			err = cmd.Run()
			ready = (err == nil)

		default:
			conn, err2 := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", hostPort), 500*time.Millisecond)
			if err2 == nil {
				_ = conn.Close()
				ready = true
			}
		}

		if ready {
			fmt.Fprintf(os.Stderr, "  --> Database '%s' is ready to accept connections!\n", engine)
			return
		}

		time.Sleep(sleepInterval)
	}

	fmt.Fprintf(os.Stderr, "Error: Timeout waiting for database '%s' to become ready.\n", engine)
	os.Exit(1)
}

func getContainerMappedPort(name string, containerPort int) int {
	cmd := exec.Command("docker", "inspect", "-f", fmt.Sprintf("{{(index (index .NetworkSettings.Ports \"%d/tcp\") 0).HostPort}}", containerPort), name)
	output, err := cmd.Output()
	if err != nil {
		return containerPort
	}
	var port int
	_, _ = fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &port)
	if port == 0 {
		return containerPort
	}
	return port
}

func startGui(dbContainerName string, dbHostPort int, guiConfig *GuiConfig, checkRunningOnly bool) {
	guiContainerName := fmt.Sprintf("%s-gui", dbContainerName)
	exists, status := containerExists(guiContainerName)

	if exists {
		if status == "running" {
			if checkRunningOnly {
				fmt.Printf("Warning: Companion GUI '%s' is already running.\n", guiContainerName)
			}
			return
		}
		fmt.Printf("* Companion GUI '%s' exists but is stopped. Starting it...\n", guiContainerName)
		_ = exec.Command("docker", "start", guiContainerName).Run()
		return
	}

	guiPort := findFreePort(guiConfig.DefaultPort)
	guiRunArgs := []string{
		"run", "-d",
		"--name", guiContainerName,
		"-p", fmt.Sprintf("%d:%d", guiPort, guiConfig.DefaultPort),
		"--add-host=host.docker.internal:host-gateway",
	}

	for _, env := range guiConfig.Env(dbHostPort) {
		guiRunArgs = append(guiRunArgs, "-e", env)
	}

	guiRunArgs = append(guiRunArgs, guiConfig.Image)

	guiCmd := exec.Command("docker", guiRunArgs...)
	if output, err := guiCmd.CombinedOutput(); err != nil {
		fmt.Printf("Warning: Failed to start companion GUI: %v\n%s\n", err, string(output))
	} else {
		fmt.Printf("%-12s %s\n", "GUI Dashboard:", guiConfig.UrlTemplate(guiPort))
	}
}

func handleDown(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Missing database engine or container name. E.g. 'dbspin down postgres'")
		os.Exit(1)
	}

	engineOrName := args[0]
	var containerName string
	if strings.HasPrefix(engineOrName, "dbspin-") {
		containerName = engineOrName
	} else {
		containerName = fmt.Sprintf("dbspin-%s", engineOrName)
	}

	exists, _ := containerExists(containerName)
	if !exists {
		fmt.Printf("Error: Database container '%s' does not exist.\n", containerName)
		return
	}

	// Clean up GUI container if exists
	guiContainerName := fmt.Sprintf("%s-gui", containerName)
	guiExists, _ := containerExists(guiContainerName)
	if guiExists {
		fmt.Printf("* Stopping companion GUI '%s'...\n", guiContainerName)
		_ = exec.Command("docker", "stop", guiContainerName).Run()
		fmt.Printf("  --> Removing container '%s'...\n", guiContainerName)
		_ = exec.Command("docker", "rm", guiContainerName).Run()
	}

	fmt.Printf("* Stopping database container '%s'...\n", containerName)
	stopCmd := exec.Command("docker", "stop", containerName)
	_ = stopCmd.Run()

	fmt.Printf("  --> Removing container '%s'...\n", containerName)
	rmCmd := exec.Command("docker", "rm", containerName)
	if err := rmCmd.Run(); err != nil {
		fmt.Printf("Error removing container: %v\n", err)
		return
	}

	fmt.Printf("* Database '%s' removed successfully!\n", strings.TrimPrefix(containerName, "dbspin-"))
}

func handleList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	verboseOpt := fs.Bool("v", false, "Verbose output showing connection strings and GUI URLs")
	verboseOptLong := fs.Bool("verbose", false, "Verbose output showing connection strings and GUI URLs")
	_ = fs.Parse(args)

	verbose := *verboseOpt || *verboseOptLong

	cmd := exec.Command("docker", "ps", "-a", "--filter", "name=dbspin-", "--format", "{{.Names}}\t{{.Image}}\t{{.Ports}}\t{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error listing databases: %v\n", err)
		os.Exit(1)
	}

	outputStr := string(output)
	if len(strings.TrimSpace(outputStr)) == 0 {
		fmt.Println("No active databases found. Run 'dbspin up <engine>' to spin one up!")
		return
	}

	lines := strings.Split(strings.TrimSpace(outputStr), "\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	if verbose {
		_, _ = fmt.Fprintln(w, "NAMES\tIMAGE\tPORTS\tSTATUS\tCONNECTION\tGUI URL")
	} else {
		_, _ = fmt.Fprintln(w, "NAMES\tIMAGE\tPORTS\tSTATUS")
	}

	for _, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) < 4 {
			continue
		}
		name := fields[0]
		image := fields[1]
		ports := fields[2]
		status := fields[3]

		if ports == "" {
			ports = "-"
		}

		if verbose {
			var engine string
			isGui := strings.HasSuffix(name, "-gui")
			cleanName := strings.TrimPrefix(name, "dbspin-")
			if isGui {
				cleanName = strings.TrimSuffix(cleanName, "-gui")
			}

			for k := range dbConfigs {
				if strings.HasPrefix(cleanName, k) {
					engine = k
					break
				}
			}

			connectionStr := "-"
			guiUrl := "-"

			if engine != "" {
				config := dbConfigs[engine]
				hostPort := getContainerMappedPort(name, config.DefaultPort)

				if isGui {
					connectionStr = "N/A"
					if config.GuiConfig != nil {
						guiUrl = config.GuiConfig.UrlTemplate(hostPort)
					}
				} else {
					connectionStr = config.ConnString(hostPort)
					if config.GuiConfig != nil {
						if config.GuiConfig.Image == "" {
							guiPort := getContainerMappedPort(name, config.GuiConfig.DefaultPort)
							guiUrl = config.GuiConfig.UrlTemplate(guiPort)
						} else {
							guiContainerName := fmt.Sprintf("%s-gui", name)
							exists, status := containerExists(guiContainerName)
							if exists && status == "running" {
								guiPort := getContainerMappedPort(guiContainerName, config.GuiConfig.DefaultPort)
								guiUrl = config.GuiConfig.UrlTemplate(guiPort)
							}
						}
					}
				}
			}

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", name, image, ports, status, connectionStr, guiUrl)
		} else {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, image, ports, status)
		}
	}

	_ = w.Flush()
}

func handleLogs(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Missing database engine or container name. E.g. 'dbspin logs postgres'")
		os.Exit(1)
	}

	engineOrName := args[0]
	var containerName string
	if strings.HasPrefix(engineOrName, "dbspin-") {
		containerName = engineOrName
	} else {
		containerName = fmt.Sprintf("dbspin-%s", engineOrName)
	}

	exists, _ := containerExists(containerName)
	if !exists {
		fmt.Printf("Error: Database container '%s' does not exist.\n", containerName)
		return
	}

	fmt.Printf("Streaming logs for '%s' (Ctrl+C to exit)...\n\n", containerName)
	cmd := exec.Command("docker", "logs", "-f", containerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}

func handleShell(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Missing database engine or container name. E.g. 'dbspin shell postgres'")
		os.Exit(1)
	}

	engineOrName := args[0]
	var containerName string
	if strings.HasPrefix(engineOrName, "dbspin-") {
		containerName = engineOrName
	} else {
		containerName = fmt.Sprintf("dbspin-%s", engineOrName)
	}

	exists, _ := containerExists(containerName)
	if !exists {
		fmt.Printf("Error: Database container '%s' does not exist.\n", containerName)
		os.Exit(1)
	}

	var engine string
	for k := range dbConfigs {
		if strings.Contains(containerName, k) {
			engine = k
			break
		}
	}

	var shellArgs []string
	switch engine {
	case "postgres", "pgvector", "timescaledb":
		shellArgs = []string{"exec", "-it", containerName, "psql", "-U", "postgres"}
	case "redis":
		shellArgs = []string{"exec", "-it", containerName, "redis-cli"}
	case "mysql":
		shellArgs = []string{"exec", "-it", containerName, "mysql", "-u", "root", "-pmysql"}
	case "mariadb":
		shellArgs = []string{"exec", "-it", containerName, "mysql", "-u", "root", "-pmariadb"}
	case "mongo":
		shellArgs = []string{"exec", "-it", containerName, "mongosh", "-u", "mongo", "-p", "mongo", "--authenticationDatabase", "admin"}
	default:
		shellArgs = []string{"exec", "-it", containerName, "sh"}
	}

	cmd := exec.Command("docker", shellArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running shell session: %v\n", err)
		os.Exit(1)
	}
}

func handleExport(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: Missing database engine or container name. E.g. 'dbspin export postgres -f dump.sql'")
		os.Exit(1)
	}

	engineOrName := args[0]
	
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	fileOpt := fs.String("f", "", "Output file path (default: write to stdout)")
	fileOptLong := fs.String("file", "", "Output file path (default: write to stdout)")
	_ = fs.Parse(args[1:])

	filePath := *fileOpt
	if *fileOptLong != "" {
		filePath = *fileOptLong
	}

	var containerName string
	if strings.HasPrefix(engineOrName, "dbspin-") {
		containerName = engineOrName
	} else {
		containerName = fmt.Sprintf("dbspin-%s", engineOrName)
	}

	exists, _ := containerExists(containerName)
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: Database container '%s' does not exist.\n", containerName)
		os.Exit(1)
	}

	var engine string
	for k := range dbConfigs {
		if strings.Contains(containerName, k) {
			engine = k
			break
		}
	}

	var exportArgs []string
	switch engine {
	case "postgres", "pgvector", "timescaledb":
		exportArgs = []string{"exec", containerName, "pg_dump", "-U", "postgres"}
	case "mysql":
		exportArgs = []string{"exec", containerName, "mysqldump", "-u", "root", "-pmysql", "mysql"}
	case "mariadb":
		exportArgs = []string{"exec", containerName, "mysqldump", "-u", "root", "-pmariadb", "mariadb"}
	case "redis":
		exportArgs = []string{"exec", containerName, "redis-cli", "--rdb", "-"}
	case "mongo":
		exportArgs = []string{"exec", containerName, "mongodump", "--archive"}
	default:
		fmt.Fprintf(os.Stderr, "Error: Database engine '%s' does not support automated exporting.\n", engine)
		os.Exit(1)
	}

	cmd := exec.Command("docker", exportArgs...)
	cmd.Stderr = os.Stderr

	if filePath != "" {
		f, err := os.Create(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
		cmd.Stdout = f
	} else {
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running export command: %v\n", err)
		os.Exit(1)
	}
}

func handleImport(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: Missing database engine or container name. E.g. 'dbspin import postgres -f dump.sql'")
		os.Exit(1)
	}

	engineOrName := args[0]

	fs := flag.NewFlagSet("import", flag.ExitOnError)
	fileOpt := fs.String("f", "", "Input file path (default: read from stdin)")
	fileOptLong := fs.String("file", "", "Input file path (default: read from stdin)")
	_ = fs.Parse(args[1:])

	filePath := *fileOpt
	if *fileOptLong != "" {
		filePath = *fileOptLong
	}

	var containerName string
	if strings.HasPrefix(engineOrName, "dbspin-") {
		containerName = engineOrName
	} else {
		containerName = fmt.Sprintf("dbspin-%s", engineOrName)
	}

	exists, _ := containerExists(containerName)
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: Database container '%s' does not exist.\n", containerName)
		os.Exit(1)
	}

	var engine string
	for k := range dbConfigs {
		if strings.Contains(containerName, k) {
			engine = k
			break
		}
	}

	var importArgs []string
	switch engine {
	case "postgres", "pgvector", "timescaledb":
		importArgs = []string{"exec", "-i", containerName, "psql", "-U", "postgres"}
	case "mysql":
		importArgs = []string{"exec", "-i", containerName, "mysql", "-u", "root", "-pmysql", "mysql"}
	case "mariadb":
		importArgs = []string{"exec", "-i", containerName, "mysql", "-u", "root", "-pmariadb", "mariadb"}
	case "redis":
		importArgs = []string{"exec", "-i", containerName, "redis-cli", "--pipe"}
	case "mongo":
		importArgs = []string{"exec", "-i", containerName, "mongorestore", "--archive"}
	default:
		fmt.Fprintf(os.Stderr, "Error: Database engine '%s' does not support automated importing.\n", engine)
		os.Exit(1)
	}

	cmd := exec.Command("docker", importArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if filePath != "" {
		f, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
		cmd.Stdin = f
	} else {
		cmd.Stdin = os.Stdin
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running import command: %v\n", err)
		os.Exit(1)
	}
}

func handlePrune() {
	fmt.Println("Warning: This will stop and remove ALL containers and volumes prefixed with 'dbspin-'.")
	fmt.Print("Are you sure you want to proceed? (y/N): ")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil || (strings.ToLower(response) != "y" && strings.ToLower(response) != "yes") {
		fmt.Println("Prune cancelled.")
		return
	}

	// 1. Get all containers starting with dbspin-
	fmt.Println("* Finding dbspin containers...")
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name=dbspin-", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		containers := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, container := range containers {
			if container == "" {
				continue
			}
			fmt.Printf("* Stopping container '%s'...\n", container)
			_ = exec.Command("docker", "stop", container).Run()
			fmt.Printf("  --> Removing container '%s'...\n", container)
			_ = exec.Command("docker", "rm", container).Run()
		}
	} else {
		fmt.Println("No dbspin containers found.")
	}

	// 2. Get all volumes starting with dbspin-
	fmt.Println("* Finding dbspin volumes...")
	cmdVol := exec.Command("docker", "volume", "ls", "--filter", "name=dbspin-", "--format", "{{.Name}}")
	outputVol, err := cmdVol.Output()
	if err == nil && len(strings.TrimSpace(string(outputVol))) > 0 {
		volumes := strings.Split(strings.TrimSpace(string(outputVol)), "\n")
		for _, volume := range volumes {
			if volume == "" {
				continue
			}
			fmt.Printf("  --> Removing volume '%s'...\n", volume)
			_ = exec.Command("docker", "volume", "rm", volume).Run()
		}
	} else {
		fmt.Println("No dbspin volumes found.")
	}

	fmt.Println("* Prune completed successfully!")
}

func handleInfo(args []string) {
	if len(args) < 1 {
		fmt.Println("Error: Missing database engine or container name. E.g. 'dbspin info postgres'")
		os.Exit(1)
	}

	engineOrName := args[0]
	var containerName string
	if strings.HasPrefix(engineOrName, "dbspin-") {
		containerName = engineOrName
	} else {
		containerName = fmt.Sprintf("dbspin-%s", engineOrName)
	}

	exists, _ := containerExists(containerName)
	if !exists {
		fmt.Printf("Error: Database container '%s' does not exist.\n", containerName)
		os.Exit(1)
	}

	// Parse out engine name
	var engine string
	for k := range dbConfigs {
		if strings.Contains(containerName, k) {
			engine = k
			break
		}
	}

	config, ok := dbConfigs[engine]
	if !ok {
		fmt.Printf("Error: Unknown database engine for container '%s'.\n", containerName)
		os.Exit(1)
	}

	// Fetch status and inspect details
	cmdStatus := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", containerName)
	status, _ := cmdStatus.Output()

	cmdIP := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", containerName)
	ip, _ := cmdIP.Output()

	mappedPort := getContainerMappedPort(containerName, config.DefaultPort)

	fmt.Println("==================================================")
	fmt.Printf("  Container Info: %s\n", containerName)
	fmt.Println("==================================================")
	fmt.Printf("Engine:       %s\n", engine)
	fmt.Printf("Status:       %s", status)
	fmt.Printf("Container IP: %s", ip)
	fmt.Printf("Host Port:    %d\n", mappedPort)
	fmt.Printf("Container Pr: %d\n", config.DefaultPort)
	fmt.Printf("Connection:   %s\n", config.ConnString(mappedPort))

	if len(config.Env) > 0 {
		fmt.Println("\nDefault Credentials:")
		for _, env := range config.Env {
			fmt.Printf("  - %s\n", env)
		}
	}

	// Check if GUI is running
	guiContainerName := fmt.Sprintf("%s-gui", containerName)
	guiExists, guiStatus := containerExists(guiContainerName)
	if guiExists {
		fmt.Printf("\nCompanion GUI Dashboard: %s-gui\n", containerName)
		fmt.Printf("  Status:     %s\n", guiStatus)
		if guiStatus == "running" && config.GuiConfig != nil {
			guiPort := getContainerMappedPort(guiContainerName, config.GuiConfig.DefaultPort)
			fmt.Printf("  URL:        %s\n", config.GuiConfig.UrlTemplate(guiPort))
		}
	} else if engine == "rabbitmq" {
		guiPort := getContainerMappedPort(containerName, 15672)
		fmt.Printf("\nBuilt-in Management UI:\n")
		fmt.Printf("  URL:        http://localhost:%d\n", guiPort)
	} else if engine == "minio" {
		guiPort := getContainerMappedPort(containerName, 9001)
		fmt.Printf("\nBuilt-in Console UI:\n")
		fmt.Printf("  URL:        http://localhost:%d\n", guiPort)
	}
	fmt.Println("==================================================")
}

func handleCompose(args []string) {
	fs := flag.NewFlagSet("compose", flag.ExitOnError)
	guiOpt := fs.Bool("gui", false, "Include companion GUI dashboards as services")
	_ = fs.Parse(args)

	var engines []string
	if len(fs.Args()) > 0 {
		for _, eng := range fs.Args() {
			if _, ok := dbConfigs[eng]; !ok {
				fmt.Printf("Error: Unknown database engine '%s'. Supported engines: postgres, redis, mysql, mariadb, mongo, elasticsearch, rabbitmq, pgvector, timescaledb, kafka, meilisearch, localstack, minio, dynamodb\n", eng)
				os.Exit(1)
			}
			engines = append(engines, eng)
		}
	} else {
		// Include all default engines in order
		engines = []string{
			"postgres", "redis", "mysql", "mariadb", "mongo",
			"elasticsearch", "rabbitmq", "pgvector", "timescaledb",
			"kafka", "meilisearch", "localstack", "minio", "dynamodb",
		}
	}

	fmt.Println("version: '3.8'")
	fmt.Println()
	fmt.Println("services:")

	var volumesUsed []string

	for _, eng := range engines {
		config := dbConfigs[eng]
		fmt.Printf("  %s:\n", eng)
		fmt.Printf("    image: %s\n", config.Image)
		fmt.Printf("    container_name: dbspin-%s\n", eng)
		
		// Port mapping
		if eng == "minio" && *guiOpt {
			fmt.Printf("    ports:\n")
			fmt.Printf("      - \"%d:%d\"\n", config.DefaultPort, config.DefaultPort)
			fmt.Printf("      - \"%d:%d\"\n", config.GuiConfig.DefaultPort, config.GuiConfig.DefaultPort)
		} else if eng == "rabbitmq" && *guiOpt {
			fmt.Printf("    ports:\n")
			fmt.Printf("      - \"%d:%d\"\n", config.DefaultPort, config.DefaultPort)
			fmt.Printf("      - \"%d:%d\"\n", config.GuiConfig.DefaultPort, config.GuiConfig.DefaultPort)
		} else {
			fmt.Printf("    ports:\n")
			fmt.Printf("      - \"%d:%d\"\n", config.DefaultPort, config.DefaultPort)
		}

		// Environment variables
		if len(config.Env) > 0 {
			fmt.Printf("    environment:\n")
			for _, env := range config.Env {
				// For Kafka inside docker-compose network, we advertise its service name and port 9092
				if eng == "kafka" && strings.HasPrefix(env, "KAFKA_ADVERTISED_LISTENERS=") {
					fmt.Printf("      - KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://kafka:9092\n")
				} else {
					fmt.Printf("      - %s\n", env)
				}
			}
		}

		// Volume mapping
		volName := fmt.Sprintf("dbspin-%s-data", eng)
		fmt.Printf("    volumes:\n")
		fmt.Printf("      - %s:%s\n", volName, config.Volume)
		volumesUsed = append(volumesUsed, volName)

		// Extra args for minio
		if eng == "minio" {
			fmt.Printf("    command: server /data --console-address :9001\n")
		}

		fmt.Printf("    restart: unless-stopped\n")
		fmt.Println()

		// Generate GUI companion if requested
		if *guiOpt && config.GuiConfig != nil && config.GuiConfig.Image != "" {
			guiConfig := config.GuiConfig
			guiName := fmt.Sprintf("%s-gui", eng)
			fmt.Printf("  %s:\n", guiName)
			fmt.Printf("    image: %s\n", guiConfig.Image)
			fmt.Printf("    container_name: dbspin-%s\n", guiName)
			fmt.Printf("    ports:\n")
			fmt.Printf("      - \"%d:%d\"\n", guiConfig.DefaultPort, guiConfig.DefaultPort)

			envList := guiConfig.Env(config.DefaultPort)
			if len(envList) > 0 {
				fmt.Printf("    environment:\n")
				for _, env := range envList {
					// Replace host.docker.internal with database service name
					replacedEnv := strings.ReplaceAll(env, "host.docker.internal", eng)
					fmt.Printf("      - %s\n", replacedEnv)
				}
			}
			fmt.Printf("    depends_on:\n")
			fmt.Printf("      - %s\n", eng)
			fmt.Printf("    restart: unless-stopped\n")
			fmt.Println()
		}
	}

	if len(volumesUsed) > 0 {
		fmt.Println("volumes:")
		for _, vol := range volumesUsed {
			fmt.Printf("  %s:\n", vol)
		}
	}
}

func handleCompletion(args []string) {
	shell := "bash"
	if len(args) > 0 {
		shell = args[0]
	}

	switch shell {
	case "setup":
		setupAutocomplete()
	case "bash":
		fmt.Print(bashCompletionScript)
	case "zsh":
		fmt.Print(zshCompletionScript)
	case "fish":
		fmt.Print(fishCompletionScript)
	default:
		fmt.Printf("Error: Unsupported shell '%s'. Supported shells: setup, bash, zsh, fish\n", shell)
		os.Exit(1)
	}
}

func handleInteractive() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("dbspin — Interactive Database Spinner")
	fmt.Println("==================================================")
	fmt.Println("Select a database engine to spin up:")

	engines := []string{
		"postgres", "redis", "mysql", "mariadb", "mongo",
		"elasticsearch", "rabbitmq", "pgvector", "timescaledb",
		"kafka", "meilisearch", "localstack", "minio", "dynamodb",
	}

	for i, eng := range engines {
		fmt.Printf("  %2d) %s\n", i+1, eng)
	}

	fmt.Print("\nEnter choice (1-14, or q to quit): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "q" || input == "quit" || input == "" {
		fmt.Println("Cancelled.")
		return
	}

	var choice int
	_, err := fmt.Sscanf(input, "%d", &choice)
	if err != nil || choice < 1 || choice > len(engines) {
		fmt.Println("Invalid choice.")
		return
	}

	engine := engines[choice-1]
	config := dbConfigs[engine]

	fmt.Printf("\nEnter host port to map (press Enter for default %d): ", config.DefaultPort)
	portInput, _ := reader.ReadString('\n')
	portInput = strings.TrimSpace(portInput)

	portVal := 0
	if portInput != "" {
		_, _ = fmt.Sscanf(portInput, "%d", &portVal)
	}

	fmt.Printf("Enter container name suffix (press Enter for none): ")
	nameInput, _ := reader.ReadString('\n')
	nameInput = strings.TrimSpace(nameInput)

	fmt.Printf("Enter custom seed file path (press Enter for none): ")
	seedInput, _ := reader.ReadString('\n')
	seedInput = strings.TrimSpace(seedInput)

	defaultTag := ""
	parts := strings.Split(config.Image, ":")
	if len(parts) > 1 {
		defaultTag = parts[1]
	}
	fmt.Printf("Enter custom image version/tag (press Enter for default '%s'): ", defaultTag)
	versionInput, _ := reader.ReadString('\n')
	versionInput = strings.TrimSpace(versionInput)

	var startGuiVal bool
	if config.GuiConfig != nil {
		fmt.Printf("Start companion web GUI dashboard? (y/N): ")
		guiInput, _ := reader.ReadString('\n')
		guiInput = strings.TrimSpace(strings.ToLower(guiInput))
		if guiInput == "y" || guiInput == "yes" {
			startGuiVal = true
		}
	}

	fmt.Printf("Wait for database to be fully ready before returning? (y/N): ")
	waitInput, _ := reader.ReadString('\n')
	waitInput = strings.TrimSpace(strings.ToLower(waitInput))
	var startWaitVal bool
	if waitInput == "y" || waitInput == "yes" {
		startWaitVal = true
	}

	// Build arguments to call handleUp
	upArgs := []string{engine}
	if portVal != 0 {
		upArgs = append(upArgs, "-port", fmt.Sprintf("%d", portVal))
	}
	if versionInput != "" {
		upArgs = append(upArgs, "-version", versionInput)
	}
	if nameInput != "" {
		upArgs = append(upArgs, "-name", nameInput)
	}
	if seedInput != "" {
		upArgs = append(upArgs, "-seed", seedInput)
	}
	if startGuiVal {
		upArgs = append(upArgs, "-gui")
	}
	if startWaitVal {
		upArgs = append(upArgs, "-wait")
	}

	fmt.Println("\n==================================================")
	fmt.Printf("Running command: dbspin up %s\n", strings.Join(upArgs, " "))
	fmt.Println("==================================================")
	handleUp(upArgs)
}

const bashCompletionScript = `_dbspin_completions() {
    local cur prev opts engines
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="up down list logs shell export import info prune compose completion help"
    engines="postgres redis mysql mariadb mongo elasticsearch rabbitmq pgvector timescaledb kafka meilisearch localstack minio dynamodb"

    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    case "${prev}" in
        up|compose)
            COMPREPLY=( $(compgen -W "${engines}" -- "${cur}") )
            return 0
            ;;
        down|logs|shell|export|import|info)
            local running_containers=$(docker ps -a --filter name=dbspin- --format "{{.Names}}" 2>/dev/null | sed 's/dbspin-//g')
            COMPREPLY=( $(compgen -W "${running_containers} ${engines}" -- "${cur}") )
            return 0
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "${cur}") )
            return 0
            ;;
    esac
}
complete -F _dbspin_completions dbspin
`

const zshCompletionScript = `#compdef dbspin
 
_dbspin() {
    local line
    local -a opts engines
    opts=(
        'up:Spin up a database container'
        'down:Stop and remove a running container'
        'list:List all running databases managed by dbspin'
        'logs:Show logs for a container'
        'shell:Connect to container via interactive CLI client'
        'export:Export database contents'
        'import:Import database contents'
        'info:Print connection parameters and credentials'
        'prune:Stop and remove ALL containers and volumes'
        'compose:Generate a docker-compose.yml configuration'
        'completion:Generate shell auto-completion script'
        'help:Print help instructions'
    )
    engines=(postgres redis mysql mariadb mongo elasticsearch rabbitmq pgvector timescaledb kafka meilisearch localstack minio dynamodb)
 
    _arguments -C \
        '1:Cmd:->cmds' \
        '*::Arg:->args'
 
    case $state in
        cmds)
            _describe -t commands 'dbspin commands' opts
            ;;
        args)
            case $line[1] in
                up|compose)
                    _values 'engines' $engines
                    ;;
                down|logs|shell|export|import|info)
                    local -a running_containers
                    running_containers=($(docker ps -a --filter name=dbspin- --format "{{.Names}}" 2>/dev/null | sed 's/dbspin-//g'))
                    _values 'containers' $running_containers $engines
                    ;;
                completion)
                    _values 'shells' bash zsh fish
                    ;;
            esac
            ;;
    esac
}
 
_dbspin "$@"
`
 
const fishCompletionScript = `# dbspin fish completion
complete -c dbspin -f
complete -c dbspin -n "not __fish_seen_subcommand_from up down list logs shell export import info prune compose completion help" -a "up down list logs shell export import info prune compose completion help"
complete -c dbspin -n "__fish_seen_subcommand_from up compose" -a "postgres redis mysql mariadb mongo elasticsearch rabbitmq pgvector timescaledb kafka meilisearch localstack minio dynamodb"
complete -c dbspin -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"
complete -c dbspin -n "__fish_seen_subcommand_from down logs shell export import info" -a "(docker ps -a --filter name=dbspin- --format '{{.Names}}' 2>/dev/null | sed 's/dbspin-//g')"
`
