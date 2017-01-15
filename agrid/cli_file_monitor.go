package main

import (
	"fmt"
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"text/tabwriter"
	"time"
)

type monitorDisplayer struct {
	writer  *tabwriter.Writer
	lines   []*agridapi.TransferEvent
	maxLine int
}

// PlatformMonitor is the main command for attaching platform subcommands.
var FileMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "monitor file tranfers",
	Long:  `monitor file tranfers`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := agridCli.fileMonitor(cmd, args); err != nil {
			agridCli.Fatal("Error: %v\n", err)
		}
	},
}

func init() {
	FileCmd.AddCommand(FileMonitorCmd)
	FileMonitorCmd.Flags().String("user", "", `set user name`)
	FileMonitorCmd.Flags().String("type", "", `monitor only file having the given type`)
	FileMonitorCmd.Flags().String("line", "10", `Max number of transferts displayed at the same time`)
}

func (m *agridCLI) fileMonitor(cmd *cobra.Command, args []string) error {
	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	api.SetUser(cmd.Flag("user").Value.String())
	fileType := cmd.Flag("type").Value.String()
	max, err := strconv.Atoi(cmd.Flag("line").Value.String())
	if err != nil {
		m.Fatal("Error option --line is not a number: %s", cmd.Flag("line").Value.String())
	}

	displayer := &monitorDisplayer{
		writer:  tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0),
		lines:   []*agridapi.TransferEvent{},
		maxLine: max,
	}
	fmt.Println("\033[2J\033[0;0H")
	fmt.Fprintf(displayer.writer, "Date\tTransfertId\tStatus\tType\tFile\tMetadata\n")
	displayer.writer.Flush()
	if err := api.FileSetTransferEventCallback(fileType, displayer.displayTransferEvent); err != nil {
		return err
	}
	for {
		time.Sleep(1 * time.Second)
	}
}

func (m *monitorDisplayer) displayTransferEvent(event *agridapi.TransferEvent) error {
	found := false
	for i, evt := range m.lines {
		if evt.TransferId == event.TransferId {
			found = true
			m.lines[i] = event
		}
	}
	if !found {
		m.lines = append(m.lines, event)
	}
	if len(m.lines) > m.maxLine {
		m.lines = m.lines[1:]
	}
	fmt.Println("\033[0;0H")
	fmt.Fprintf(m.writer, "Date\tTransfertId\tStatus\tType\tFile\tMetadata\n")
	for _, evt := range m.lines {
		fmt.Fprintf(m.writer, "%s\t%s\t%s\t%s\t%s\t%v                            \n", evt.EventDate, evt.TransferId, evt.State, evt.FileType, evt.FileName, evt.Metadata)
	}
	m.writer.Flush()
	return nil
}
