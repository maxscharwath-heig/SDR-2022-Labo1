package utils

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func PrintWelcome() {
	fmt.Println("\n   _____ ____  ____ \n  / ___// __ \\/ __ \\\n  \\__ \\/ / / / /_/ /\n ___/ / /_/ / _, _/ \n/____/_____/_/ |_|")
	fmt.Println("Welcome to the SDR-Labo1 client")
	fmt.Println("")
}

func PrintHelp() {
	fmt.Println("Please type the wished command")
	fmt.Println("List of commands:")
	fmt.Println("- create")
	fmt.Println("- close")
	fmt.Println("- register")
	fmt.Println("- show")
	fmt.Println("- show [number]")
	fmt.Println("- show [number] --resume")
	fmt.Println("_________________________")
}

func PrintTable(headers []string, data []string) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 3, '\t', tabwriter.AlignRight)

	formattedHeaders := strings.Join(headers[:], "\t")
	fmt.Fprintln(writer, formattedHeaders)

	for _, row := range data {
		fmt.Fprintln(writer, row)
	}

	writer.Flush()
}
