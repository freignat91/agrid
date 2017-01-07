package agridapi

import (
	"fmt"
)

// NodeWriteStats makes node write node stats in node logs
func (api *AgridAPI) InfoWriteStats(node string) error {
	if node == "" {
		node = "*"
	}
	client, err := api.getClient()
	if err != nil {
		return err
	}
	if _, err := client.createSendMessage(node, false, "writeStatsInLog"); err != nil {
		return err
	}
	return nil
}

// InfoGetNodeName get the node name having index "index"
func (api *AgridAPI) InfoGetNodeName(index int) (string, error) {
	client, err := api.getClient()
	if err != nil {
		return "", err
	}
	ret, err := client.createSendMessage("", false, "getNodeName", fmt.Sprintf("%d", index))
	if err != nil {
		return "", err
	}
	return ret.Args[0], nil
}
