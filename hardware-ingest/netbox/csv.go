package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
)

// readMachineBytes Function reads a byte array and converts it back to Machine Array.
func readMachineBytes(machines []byte, n *Netbox) ([]*Machine, error) {
	var hardwareMachines []*Machine
	err := json.Unmarshal(machines, &hardwareMachines)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling the input byte stream: %v", err)
	}

	n.logger.V(1).Info("Deserealizing input stream successful", "num_machines", len(hardwareMachines))

	return hardwareMachines, nil
}

// WriteToCSV Helper Function creates Error channel and context to cancel file writing on keyboard interrupt.
func writeToCSVHelper(ctx context.Context, machines []*Machine, n *Netbox) error {
	errChan := make(chan error)
	go writeToCSV(errChan, machines, n)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// writeToCSV Function reads from slice of machines and writes to a hardware.csv file.
func writeToCSV(errChan chan error, machines []*Machine, n *Netbox) {
	// Create a csv file usign OS operations
	file, err := os.Create("hardware.csv")
	if err != nil {
		errChan <- fmt.Errorf("error creating file: %v", err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	headers := [11]string{"hostname", "bmc_ip", "bmc_username", "bmc_password", "mac", "ip_address", "netmask", "gateway", "nameservers", "labels", "disk"}
	err = writer.Write(headers[:])
	if err != nil {
		errChan <- fmt.Errorf("error Writing Column names into file: %v", err)
		return
	}
	var machinesString [][]string
	for _, machine := range machines {
		nsCombined := extractNameServers(machine.Nameservers)
		row := []string{machine.Hostname, machine.BMCIPAddress, machine.BMCUsername, machine.BMCPassword, machine.MACAddress, machine.IPAddress, machine.Netmask, machine.Gateway, nsCombined, "type=" + machine.Labels["type"], machine.Disk}
		machinesString = append(machinesString, row)
	}
	err = writer.WriteAll(machinesString)
	if err != nil {
		errChan <- fmt.Errorf("error writing to file: %v", err)
		return
	}
	mydir, _ := os.Getwd()
	n.logger.V(1).Info("Write to csv successful", "path_to_file", mydir+"/hardware.csv")
	errChan <- nil
}

// extractNameServers Function reads a slice of string and combines them with '|' separator.
func extractNameServers(nameservers []string) string {
	nsCombined := ""
	for idx, ns := range nameservers {
		if idx == 0 {
			nsCombined += ns
		} else {
			nsCombined = nsCombined + "|" + ns
		}
	}
	return nsCombined
}
