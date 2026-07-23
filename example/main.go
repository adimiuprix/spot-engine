package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "mgmt":
			runManagementExample()
		case "iceberg":
			runIcebergExample()
		case "amend":
			runAmendExample()
		case "snapshot":
			runSnapshotExample()
		default:
			runTradingExample()
		}
	} else {
		runTradingExample()
	}
}

func runTradingExample() {
	fmt.Println("=== Trading Example ===\n")
	tradingExample()
}
