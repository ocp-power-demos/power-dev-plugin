package main

import (
	"fmt"
	"os"

	"github.com/jaypipes/ghw"
)

// Launch the scanner
func main() {
	devices, err := ScanRootForDevices()
	if err != nil {
		fmt.Printf("Could not scan devices, aborting %s", err)
		os.Exit(2)
	}
	for idx, device := range devices {
		fmt.Printf("%d - %s", idx, device)
	}
}

// scans the local disk using ghw to find the blockdevices
func ScanRootForDevices() ([]string, error) {
	// relies on GHW_CHROOT=/host/dev
	// lsblk -f --json --paths -s | jq -r '.blockdevices[] | select(.fstype != "xfs")' | grep mpath | grep -v fstype | sort -u | wc -l
	// This may be the best way to get the devices.
	block, err := ghw.Block()
	if err != nil {
		fmt.Printf("Error getting block storage info: %v", err)
		return nil, err
	}
	devices := []string{}
	fmt.Printf("DEVICE: %v\n", block)
	for _, disk := range block.Disks {
		fmt.Printf("    - DISK: %v\n", disk.Name)
		for _, part := range disk.Partitions {
			fmt.Printf("        - PART: %v\n", part.Disk.Name)
			devices = append(devices, part.Name)
		}
	}
	return devices, nil
}
