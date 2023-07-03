package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/vision-cli/common/execute"
	"github.com/vision-cli/common/tmpl"
	"github.com/vision-cli/vision-plugin-gateway-v1/plugin"
)

func main() {
	input := ""
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input += scanner.Text()
	}
	if scanner.Err() != nil {
		panic(scanner.Err())
	}
	e := execute.NewOsExecutor()
	t := tmpl.NewOsTmpWriter()
	fmt.Fprint(os.Stdout, plugin.Handle(input, e, t))
}
