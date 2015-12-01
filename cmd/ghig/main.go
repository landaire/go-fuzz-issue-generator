package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
)

const (
	hangMessage      = "The following program hangs with the provided go-fuzz output:"
	panicMessage     = "The following program panics with the provided output:"
	standardTemplate = `{{ .Message }}

` + "```" + `
{{ .Program }}
` + "```" + `

-----

Standard output:

` + "```" + `
{{ .Output }}
` + "```" + `

Output ran through [panicparse](https://github.com/maruel/panicparse):


` + "```" + `
{{ .PanicParse }}
` + "```" + `


This {{ .Type }} was discovered with [go-fuzz](https://github.com/dvyukov/go-fuzz).
`
)

type CrashDescription struct {
	Type, Message, Program, Output, PanicParse string
}

type Application struct {
	QuotedData string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Invalid arguments given. Usage: ghig <applicationTemplate> <unquotedCrashFile>")

		return
	}

	appTemplateText, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	applicationTemplate := template.Must(template.New("application").Parse(string(appTemplateText)))
	descriptionTemplate := template.Must(template.New("application").Parse(standardTemplate))

	var (
		applicationOutput bytes.Buffer
		descriptionOutput bytes.Buffer
	)

	quotedData, err := ioutil.ReadFile(os.Args[2] + ".quoted")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Get the application sample
	if err := applicationTemplate.Execute(&applicationOutput, Application{
		string(quotedData),
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Read go-fuzz's output
	output, err := ioutil.ReadFile(os.Args[2] + ".output")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	message := panicMessage
	crashType := "panic"

	if strings.HasPrefix(string(output), "program hanged") {
		message = hangMessage
		crashType = "hang"
	} else {
		file, err := os.Create(path.Join(os.TempDir(), "ghig.go"))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}

		outputCopy := bytes.NewBuffer(applicationOutput.Bytes())
		io.Copy(file, outputCopy)
		file.Close()

		goRunCommand := exec.Command("go", "run", file.Name())

		output, _ = goRunCommand.CombinedOutput()
	}

	if err := descriptionTemplate.Execute(&descriptionOutput, NewCrashDescription(crashType, message, string(applicationOutput.Bytes()), string(output))); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	ioutil.WriteFile(os.Args[2]+"_description.md", descriptionOutput.Bytes(), os.ModePerm)
}

func NewCrashDescription(crashType, message, program, output string) CrashDescription {
	goRunCommand := exec.Command("panicparse")
	goRunCommand.Stdin = strings.NewReader(output)

	goRunStdout, err := goRunCommand.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	return CrashDescription{
		Type:       crashType,
		Message:    message,
		Program:    program,
		PanicParse: string(goRunStdout),
		Output:     output,
	}
}
