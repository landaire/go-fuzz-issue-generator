# Go-Fuzz Issue Generator (gfig)

gfig is a utility for creating descriptions related to [Go-Fuzz](https://github.com/dvyukov/go-fuzz) crash files

Usage: `gfig <gitRepo> <applicationTemplate> <crashFile>`

The git repo path should be the full path for target package's git repo. This is used for generating revision information. The application template is what will be used for generating the sample application.
An example of a template is:

```
package main

import (
    "github.com/foo/bar"
)

var (
    data = {{ .QuotedData }}
)

func main() {
    config, err := bar.Load(data)

    if config != nil && err != nil {
        panic("non-nil config with err")
    }
}
```

When `gfig` looks at the crash file it will determine the crash type (panic or hang),
apply the respective message, and generate a description of the crash with the filename
`<crashFileName>_description.md`.

The `{{ .QuotedData }}` in the application template will be replaced with the contents
of `<crashFile>.quoted`.

## Examples

### Panic

> The following program at revision `tags/v0.3.1-0-g6d743bb` panics with the provided output:
> 
> ```
> package main
> 
> import (
>     "github.com/pelletier/go-toml"
> )
> 
> var (
>     data =  "[i.][\".\"][[.]][[i]][" +
>     "[.]][[6125147516=134" +
>     "68168619859891416e4R" +
>     "=4646778106689453125" +
>     "5e6]][[[[i&\n]][[i]][" +
>     "[.][.][.]][[.]][]`[." +
>     "]`[[.]][[.]][0xx][]"
> 
> )
> 
> func main() {
>     config, err := toml.Load(data)
> 
>     if config != nil && err != nil {
>         panic("non-nil config with err")
>     }
> }
> 
> ```
> 
> -----
> 
> Standard output:
> 
> ```
> panic: interface conversion: interface is *toml.TomlTree, not []*toml.TomlTree [recovered]
>     panic: interface conversion: interface is *toml.TomlTree, not []*toml.TomlTree
> 
> goroutine 1 [running]:
> github.com/pelletier/go-toml.Load.func1(0xc820085f08)
>     /Users/lander/go/src/github.com/pelletier/go-toml/toml.go:368 +0x8e
> github.com/pelletier/go-toml.(*tomlParser).parseGroupArray(0xc8200182a0, 0xc82000a770)
>     /Users/lander/go/src/github.com/pelletier/go-toml/parser.go:110 +0xe2a
> github.com/pelletier/go-toml.(*tomlParser).(github.com/pelletier/go-toml.parseGroupArray)-fm(0xc82000a770)
>     /Users/lander/go/src/github.com/pelletier/go-toml/parser.go:80 +0x20
> github.com/pelletier/go-toml.(*tomlParser).run(0xc8200182a0)
>     /Users/lander/go/src/github.com/pelletier/go-toml/parser.go:30 +0x46
> github.com/pelletier/go-toml.parseToml(0xc820018240, 0x8b)
>     /Users/lander/go/src/github.com/pelletier/go-toml/parser.go:361 +0x274
> github.com/pelletier/go-toml.Load(0x1b81e0, 0x8b, 0x0, 0x0, 0x0)
>     /Users/lander/go/src/github.com/pelletier/go-toml/toml.go:373 +0x8e
> main.main()
>     /var/folders/6y/xxqr1vqn6q7c_ttvdgjt7p1w0000gn/T/gfig.go:19 +0x33
> 
> goroutine 5 [chan send]:
> github.com/pelletier/go-toml.(*tomlLexer).emit(0xc8200142d0, 0x11)
>     /Users/lander/go/src/github.com/pelletier/go-toml/lexer.go:62 +0xb8
> github.com/pelletier/go-toml.(*tomlLexer).lexInsideKeyGroupArray(0xc8200142d0, 0xc82000a780)
>     /Users/lander/go/src/github.com/pelletier/go-toml/lexer.go:486 +0xe6
> github.com/pelletier/go-toml.(*tomlLexer).(github.com/pelletier/go-toml.lexInsideKeyGroupArray)-fm(0xc82000a780)
>     /Users/lander/go/src/github.com/pelletier/go-toml/lexer.go:467 +0x20
> github.com/pelletier/go-toml.(*tomlLexer).run(0xc8200142d0)
>     /Users/lander/go/src/github.com/pelletier/go-toml/lexer.go:35 +0x46
> created by github.com/pelletier/go-toml.lexToml
>     /Users/lander/go/src/github.com/pelletier/go-toml/lexer.go:586 +0xd0
> exit status 2
> 
> ```
> 
> Output ran through [panicparse](https://github.com/maruel/panicparse):
> 
> 
> ```
> panic: interface conversion: interface is *toml.TomlTree, not []*toml.TomlTree [recovered]
>     panic: interface conversion: interface is *toml.TomlTree, not []*toml.TomlTree
> 
> exit status 2
> 1: running
>     go-toml toml.go:368   Load.func1(0xc820085f08)
>     go-toml parser.go:110 (*tomlParser).parseGroupArray(#2, #1)
>     go-toml parser.go:80  parseGroupArray)-fm(#1)
>     go-toml parser.go:30  (*tomlParser).run(#2)
>     go-toml parser.go:361 parseToml(0xc820018240, 0x8b)
>     go-toml toml.go:373   Load(0x1b81e0, 0x8b, 0, 0, 0)
>     main    gfig.go:19    main()
> 1: chan send [Created by go-toml.lexToml @ lexer.go:586]
>     go-toml lexer.go:62   (*tomlLexer).emit(#4, 0x11)
>     go-toml lexer.go:486  (*tomlLexer).lexInsideKeyGroupArray(#4, #3)
>     go-toml lexer.go:467  lexInsideKeyGroupArray)-fm(#3)
>     go-toml lexer.go:35   (*tomlLexer).run(#4)
> 
> ```
> 
> 
> This panic was discovered with [go-fuzz](https://github.com/dvyukov/go-fuzz).

----

#### Hang

These need some work since the output used is `<crashFile>.output`, which contains
a lot of information

