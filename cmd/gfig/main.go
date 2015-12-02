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
	hangMessage      = "hangs with the provided go-fuzz output:"
	panicMessage     = "panics with the provided output:"
	standardTemplate = `The following program at revision ` + "`" + `{{ .Revision }}` + "`" + ` {{ .Message }}

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
	Type, Message, Program, Output, PanicParse, Revision string
}

type Application struct {
	QuotedData string
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Invalid arguments given. Usage: gfig <gitRepo> <applicationTemplate> <unquotedCrashFile>")

		return
	}

	gitRepo := os.Args[1]
	appTemplatePath := os.Args[2]
	crashFile := os.Args[3]

	appTemplateText, err := ioutil.ReadFile(appTemplatePath)
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

	quotedData, err := ioutil.ReadFile(crashFile + ".quoted")
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
	output, err := ioutil.ReadFile(crashFile + ".output")
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
		file, err := os.Create(path.Join(os.TempDir(), "gfig_sample.go"))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}

		outputCopy := bytes.NewBuffer(applicationOutput.Bytes())
		io.Copy(file, outputCopy)
		file.Close()

		goRunCommand := exec.Command("go", "run", file.Name())

		output, _ = goRunCommand.CombinedOutput()

		if len(output) == 0 {
			fmt.Fprintln(os.Stderr, crashFile, "is not a valid crash")
			os.Exit(0)
		}
	}

	crashDescription := NewCrashDescription(gitRepo, crashType, message, string(applicationOutput.Bytes()), string(output))

	if err := descriptionTemplate.Execute(&descriptionOutput, crashDescription); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	ioutil.WriteFile(crashFile+"_description.md", descriptionOutput.Bytes(), os.ModePerm)
}

func NewCrashDescription(gitRepo, crashType, message, program, output string) CrashDescription {
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
		Revision:   GitDescribeRepo(gitRepo),
	}
}

func GitDescribeRepo(repoPath string) string {
	gitDescribeCommand := exec.Command("git", "describe", "--all", "--long")
	gitDescribeCommand.Dir = repoPath

	gitStdout, err := gitDescribeCommand.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	return string(gitStdout)
}
