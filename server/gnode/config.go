package gnode

import (
	"fmt"
	"os"
	"strconv"
)

//AgentConfig configuration parameters
type GNodeConfig struct {
	nbNode           int
	nbLineConnect    int
	nbCrossConnect   int
	nbDuplicate      int
	nbDuplicateAck   int
	restPort         string
	grpcPort         string
	udpPort          string
	bufferSize       int
	parallelSender   int
	parallelReceiver int
	rootDataPath     string
	tracePersistence int
}

//update conf instance with default value and environment variables
func (cfg *GNodeConfig) init(version string, build string) {
	cfg.setDefault()
	cfg.loadConfigUsingEnvVariable()
	cfg.displayConfig(version, build)
}

//Set default value of configuration
func (cfg *GNodeConfig) setDefault() {
	cfg.nbLineConnect = 0 //0=auto-computed
	cfg.nbCrossConnect = 0
	cfg.restPort = "3000"
	cfg.grpcPort = "30103"
	cfg.udpPort = "3010"
	cfg.bufferSize = 10
	cfg.parallelReceiver = 100
	cfg.parallelSender = 100
	cfg.nbDuplicate = 3
	cfg.nbDuplicateAck = 1
	cfg.rootDataPath = "/data"
	cfg.tracePersistence = 200
}

//Update config with env variables
func (cfg *GNodeConfig) loadConfigUsingEnvVariable() {
	cfg.nbLineConnect = cfg.getIntParameter("NB_LINE_CONNECT", cfg.nbLineConnect)
	cfg.nbCrossConnect = cfg.getIntParameter("NB_CROSS_CONNECT", cfg.nbCrossConnect)
	cfg.restPort = cfg.getStringParameter("REST_PORT", cfg.restPort)
	cfg.grpcPort = cfg.getStringParameter("GRPC_PORT", cfg.grpcPort)
	cfg.udpPort = cfg.getStringParameter("UDP_PORT", cfg.udpPort)
	cfg.bufferSize = cfg.getIntParameter("BUFFER_SIZE", cfg.bufferSize)
	cfg.parallelSender = cfg.getIntParameter("PARALLEL_SENDER", cfg.parallelSender)
	cfg.parallelReceiver = cfg.getIntParameter("PARALLEL_RECEIVER", cfg.parallelReceiver)
	cfg.nbDuplicate = cfg.getIntParameter("NB_DUPLICATE", cfg.nbDuplicate)
	cfg.nbDuplicateAck = cfg.getIntParameter("NB_DUPLICATE_ACK", cfg.nbDuplicateAck)
	cfg.rootDataPath = cfg.getStringParameter("DATA_PATH", cfg.rootDataPath)
}

//display amp-pilot configuration
func (cfg *GNodeConfig) displayConfig(version string, build string) {
	fmt.Printf("agrid version: %v build: %s\n", version, build)
	fmt.Println("----------------------------------------------------------------------------")
	fmt.Println("Configuration:")
	if cfg.nbLineConnect == 0 {
		fmt.Printf("nb connect: auto\n")
	} else {
		fmt.Printf("nb line connect: %d\n", cfg.nbLineConnect)
		fmt.Printf("nb cross connect: %d\n", cfg.nbCrossConnect)
	}
	fmt.Printf("rest_port: %s\n", cfg.restPort)
	fmt.Printf("grpc_port: %s\n", cfg.grpcPort)
	fmt.Printf("udp_port: %s\n", cfg.udpPort)
	fmt.Printf("buffer_size: %d\n", cfg.bufferSize)
	fmt.Printf("parallel_sender: %d\n", cfg.parallelSender)
	fmt.Printf("parallel_receiver: %d\n", cfg.parallelReceiver)
	fmt.Println("----------------------------------------------------------------------------")
}

//return env variable value, if empty return default value
func (cfg *GNodeConfig) getStringParameter(envVariableName string, def string) string {
	value := os.Getenv(envVariableName)
	if value == "" {
		return def
	}
	return value
}

//return env variable value convert to int, if empty return default value
func (cfg *GNodeConfig) getIntParameter(envVariableName string, def int) int {
	value := os.Getenv(envVariableName)
	if value != "" {
		ivalue, err := strconv.Atoi(value)
		if err != nil {
			return def
		}
		return ivalue
	}
	return def
}
