package utils

import (
	"fmt"
	"os"
	"sdr/labo1/src/utils/colors"
	"strings"
	"text/tabwriter"
)

func PrintWelcome() {
	fmt.Println(colors.Purple, colors.Bold, "\n▄▀▀▀ █▀▀▄ █▀▀▄\n ▀▀▄ █  █ █▄▄▀\n▄▄▄▀ █▄▄▀ █  █", "v. 1.0", colors.Reset)
	fmt.Println(colors.Red, colors.Bold, "by Nicolas Crausaz & Maxime Scharwath", colors.Reset)
	fmt.Println(colors.BackgroundYellow, colors.Red, colors.Bold, "Welcome to the SDR-Labo1 client", colors.Reset)
	fmt.Println("This client allows you to create & manage events")
}

func PrintHelp() {
	fmt.Println("Please type the wished command")
	fmt.Println(colors.Underline, "List of commands:", colors.Reset)
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
