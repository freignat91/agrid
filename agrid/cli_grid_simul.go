package main

import (
	"github.com/freignat91/agrid/server/gnode"
	"github.com/spf13/cobra"
	"strconv"
)

// PlatformMonitor is the main command for attaching platform subcommands.
var GridSimulCmd = &cobra.Command{
	Use:   "simul",
	Short: "grid simulation",
	Long:  `grid connections simulation and stats`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientManager.gridSimul(cmd, args); err != nil {
			clientManager.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	GridCmd.AddCommand(GridSimulCmd)
	GridSimulCmd.Flags().Int("line", 0, `number of line connections`)
	GridSimulCmd.Flags().Int("cross", 0, `number of cross connections`)
}

func (m *ClientManager) gridSimul(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		m.Fatal("First argument should be a node number\n")
	}
	nbNode := 0
	if nb, err := strconv.Atoi(args[0]); err != nil {
		m.Fatal("First argument (node number) should be a number: %v\n", err)
	} else {
		nbNode = nb
	}
	nbLineConnect := 0
	if nb, err := strconv.Atoi(cmd.Flag("line").Value.String()); err != nil {
		m.Fatal("Parameter --line should be a number: %v\n", err)
	} else {
		nbLineConnect = nb
	}
	nbCrossConnect := 0
	if nb, err := strconv.Atoi(cmd.Flag("cross").Value.String()); err != nil {
		m.Fatal("Parameter --cross should be a number: %v\n", err)
	} else {
		nbCrossConnect = nb
	}
	gnode.CreateGrid(nbNode, nbLineConnect, nbCrossConnect, true)
	return nil
}
