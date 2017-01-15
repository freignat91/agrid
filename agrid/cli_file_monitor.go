package main

import (
	"fmt"
	"github.com/freignat91/agrid/agridapi"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
	"time"
)

type monitorDisplayer struct {
	writer *tabwriter.Writer
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
}

func (m *agridCLI) fileMonitor(cmd *cobra.Command, args []string) error {
	api := agridapi.New(m.server)
	m.setAPILogLevel(api)
	api.SetUser(cmd.Flag("user").Value.String())
	fileType := cmd.Flag("type").Value.String()

	displayer := &monitorDisplayer{
		writer: tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0),
	}
	fmt.Fprintf(displayer.writer, "Date\tTransfertId\tStatus\tType\tFile\tMetadata\n")
	displayer.writer.Flush()
	err := api.FileSetTransferEventCallback(fileType, displayer.displayTransferEvent)
	if err != nil {
		return err
	}
	for {
		time.Sleep(1 * time.Second)
	}
}

func (m *monitorDisplayer) displayTransferEvent(event *agridapi.TransferEvent) error {
	fmt.Fprintf(m.writer, "%s\t%s\t%s\t%s\t%v\n", event.EventDate, event.TransferId, event.State, event.FileName, event.Metadata)
	m.writer.Flush()
	return nil
}
