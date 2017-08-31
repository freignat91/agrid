package agridapi

import (
	"fmt"
	"log"
	"strings"
)

const (
	LOG_ERROR = 0
	LOG_WARN  = 1
	LOG_INFO  = 2
	LOG_DEBUG = 3
)

type AgridAPI struct {
	serverList  []string
	serverIndex int
	logLevel    int
	userName    string
	userToken   string
}

// New create an Agrid api instance
func New(servers string) *AgridAPI {
	serverList := strings.Split(servers, ",")
	for i, serv := range serverList {
		serverList[i] = strings.Trim(serv, " ")
	}
	api := &AgridAPI{
		serverList: serverList,
		logLevel:   LOG_WARN,
	}
	api.userName = "common"
	return api
}

func (api *AgridAPI) getNextServerAddr() string {
	addr := api.serverList[api.serverIndex]
	api.serverIndex++
	if api.serverIndex >= len(api.serverList) {
		api.serverIndex = 0
	}
	return addr
}

func (api *AgridAPI) getClient() (*gnodeClient, error) {
	client := gnodeClient{}
	err := client.init(api)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (api *AgridAPI) SetLogLevel(level string) {
	if strings.ToLower(level) == "error" {
		api.logLevel = LOG_ERROR
	} else if strings.ToLower(level) == "warn" {
		api.logLevel = LOG_WARN
	} else if strings.ToLower(level) == "info" {
		api.logLevel = LOG_INFO
	} else if strings.ToLower(level) == "debug" {
		api.logLevel = LOG_DEBUG
	}
}

func (api *AgridAPI) LogLevelString() string {
	switch api.logLevel {
	case LOG_ERROR:
		return "error"
	case LOG_WARN:
		return "warn"
	case LOG_INFO:
		return "info"
	case LOG_DEBUG:
		return "debug"
	default:
		return "?"
	}
}

func (api *AgridAPI) error(format string, args ...interface{}) {
	if api.logLevel >= LOG_ERROR {
		log.Printf(format, args...)
	}
}

func (api *AgridAPI) warn(format string, args ...interface{}) {
	if api.logLevel >= LOG_WARN {
		log.Printf(format, args...)
	}
}

func (api *AgridAPI) info(format string, args ...interface{}) {
	if api.logLevel >= LOG_INFO {
		log.Printf(format, args...)
	}
}

func (api *AgridAPI) debug(format string, args ...interface{}) {
	if api.logLevel >= LOG_DEBUG {
		log.Printf(format, args...)
	}
}

func (api *AgridAPI) isDebug() bool {
	if api.logLevel >= LOG_DEBUG {
		return true
	}
	return false
}

func (api *AgridAPI) printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (api *AgridAPI) formatKey(key string) string {
	if key != "" {
		for len(key) < 32 {
			key = fmt.Sprintf("%s%s", key, key)
		}
		key = key[0:32]
	}
	return key
}

// SetUser define the current user
func (api *AgridAPI) SetUser(user string) {
	api.userName = "common"
	api.userToken = ""
	if user != "" {
		list := strings.Split(user, ":")
		if len(list) == 2 {
			api.userName = list[0]
			api.userToken = list[1]
		} else {
			api.userName = list[0]
		}
	}
}

// UserCreate create an user and return a token
func (api *AgridAPI) UserCreate(name string, token string) (string, error) {
	if err := api.verifyUserName(name); err != nil {
		return "", fmt.Errorf("Invalide user name: %v", err)
	}
	client, err := api.getClient()
	if err != nil {
		return "", err
	}
	ret, errs := client.createSendMessage("", true, "createUser", name, token)
	if errs != nil {
		return "", errs
	}
	return ret.Args[0], nil
}

func (api *AgridAPI) verifyUserName(name string) error {
	if name == "" || name == "common" {
		return fmt.Errorf("Invalid user name")
	}
	if strings.IndexAny(name, " /\\") >= 0 {
		return fmt.Errorf("Invalid character")
	}
	return nil
}

// UserRemove create an user
func (api *AgridAPI) UserRemove(name string, force bool) error {
	api.SetUser(name)
	defer func() {
		api.userName = ""
		api.userToken = ""
	}()
	client, err := api.getClient()
	if err != nil {
		return err
	}
	_, errs := client.createSendMessage("*", true, "removeUser", api.userName, api.userToken, fmt.Sprintf("%t", force))
	if errs != nil {
		return errs
	}
	return nil
}
