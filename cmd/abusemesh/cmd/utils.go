package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/grpc/status"
)

//Prints the given error and exits
func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
	os.Exit(1)
}

//Prints the given grpc error and exits
func exitWithGrpcError(err error) {
	status, ok := status.FromError(err)
	if !ok {
		exitWithError(err)
	}

	fmt.Fprintf(os.Stderr, "ERROR: %s\n", status.Message())
	os.Exit(1)
}

//ConsoleOutputFormatter determines how output data will be formatted
//A nil value will print human readable output, all other values will print computer output
var ConsoleOutputFormatter OutputFormatter

//OutputFormatter declares the function signature for formatting output
type OutputFormatter func(object interface{}) string

//JSONFormatter formats all values as json
var JSONFormatter OutputFormatter = func(object interface{}) string {
	buf := bytes.Buffer{}

	jsonEncoder := json.NewEncoder(&buf)
	jsonEncoder.SetIndent("", "    ")
	jsonEncoder.Encode(object)

	return buf.String()
}

//printToStdout will print human or machine formatted output depending on the value of ConsoleOutputFormatter
func printToStdout(object interface{}, humanPrint OutputFormatter) {
	if ConsoleOutputFormatter != nil {
		fmt.Println(ConsoleOutputFormatter(object))
	} else {
		fmt.Println(humanPrint(object))
	}
}
