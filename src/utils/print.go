// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

// The file contains various printing utilities used in the application

package utils

import (
	"fmt"
	"os"
	"sdr/labo1/src/utils/colors"
	"strings"
	"text/tabwriter"
)

func printLogo() {
	version := "v2.0.0"
	fmt.Println(colors.Purple+colors.Bold+"\n▄▀▀▀ █▀▀▄ █▀▀▄\n ▀▀▄ █  █ █▄▄▀\n▄▄▄▀ █▄▄▀ █  █", version, colors.Reset)
	fmt.Println(colors.Red+colors.Bold+"by Nicolas Crausaz & Maxime Scharwath", colors.Reset)
}

func PrintClientWelcome() {
	printLogo()
	fmt.Println(colors.BackgroundYellow, colors.Red, colors.Bold, "Welcome to the SDR-Labo1 client", colors.Reset)
	fmt.Println("This client allows you to create & manage events")
}

func PrintServerWelcome() {
	printLogo()
	fmt.Println(colors.BackgroundYellow + colors.Red + colors.Bold + "Welcome to the SDR-Labo1 server" + colors.Reset)
	fmt.Println(colors.Underline + "Write [quit] to quit server" + colors.Reset)
}

func PrintHelp() {
	fmt.Println("Please type the wished command")
	fmt.Println(colors.Underline + "List of commands:" + colors.Reset)
	fmt.Println("- create")
	fmt.Println("- close")
	fmt.Println("- register")
	fmt.Println("- show")
	fmt.Println("- show [number]")
	fmt.Println("- show [number] --resume")
	fmt.Println("- quit")
	fmt.Println("_________________________")
}

func PrintTable(headers []string, data []string) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 3, '\t', tabwriter.AlignRight)

	formattedHeaders := strings.Join(headers[:], "\t")
	_, _ = fmt.Fprintln(writer, formattedHeaders)

	for _, row := range data {
		_, _ = fmt.Fprintln(writer, row)
	}

	_ = writer.Flush()
}

func PrintSuccess(message string) {
	fmt.Println("✅ " + colors.Green + message + colors.Reset)
}

func PrintError(message string) {
	fmt.Println("❌ " + colors.Red + message + colors.Reset)
}
