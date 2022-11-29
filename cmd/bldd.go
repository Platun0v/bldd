package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/platun0v/bldd/pkg"
	"github.com/spf13/cobra"
	"os"
	"sort"
)

type outputFormat string

const (
	outputFormatTXT  outputFormat = "txt"
	outputFormatJSON outputFormat = "json"
)

var (
	outputFormatFlag outputFormat = outputFormatTXT
)

func (e *outputFormat) String() string {
	return string(*e)
}

func (e *outputFormat) Set(v string) error {
	switch v {
	case "txt", "json":
		*e = outputFormat(v)
		return nil
	default:
		return errors.New("invalid output format")
	}
}

func (e *outputFormat) Type() string {
	return "string"
}

// RootCmd - Bldd(backward ldd) is a tool that shows all EXECUTABLE files that use specified shared library files.
var RootCmd = &cobra.Command{
	Use:   "Bldd [OPTIONS] EXECUTABLES",
	Short: "Bldd(backward ldd) is a tool that shows all EXECUTABLE files that use specified shared library files",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")

		files := args

		if len(files) == 0 {
			fmt.Println("No files specified")
			return
		}

		// Remove trailing slash
		if len(output) > 0 && output[len(output)-1] == '/' {
			output = output[:len(output)-1]
		}

		for i, file := range files {
			if file[len(file)-1] == '/' {
				files[i] = file[:len(file)-1]
			}
		}

		err := Bldd(output, outputFormatFlag, files)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.Flags().StringP("output", "o", "", "Output file. Default is stdout")
	RootCmd.MarkFlagFilename("output")
	RootCmd.Flags().VarP(&outputFormatFlag, "format", "f", "Output format. Available formats: txt, json")
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type ElfFileCount struct {
	count int
	files []string
}

type Files struct {
	Count int      `json:"count"`
	Files []string `json:"files"`
	Lib   string   `json:"lib"`
}

func Bldd(output string, format outputFormat, files []string) error {
	var elfFiles []string
	for _, file := range files {
		gotFiles, err := pkg.FindElf(file)
		if err != nil {
			return err
		}
		elfFiles = append(elfFiles, gotFiles...)
	}
	x32Libs := make(map[string]*ElfFileCount)
	x64Libs := make(map[string]*ElfFileCount)

	for _, elfFile := range elfFiles {
		x32, x64, err := pkg.Ldd(elfFile)

		if err != nil {
			return err
		}

		for _, lib := range x32 {
			if _, ok := x32Libs[lib]; ok {
				x32Libs[lib].count++
				x32Libs[lib].files = append(x32Libs[lib].files, elfFile)
			} else {
				x32Libs[lib] = &ElfFileCount{count: 1, files: []string{elfFile}}
			}
		}

		for _, lib := range x64 {
			if _, ok := x64Libs[lib]; ok {
				x64Libs[lib].count++
				x64Libs[lib].files = append(x64Libs[lib].files, elfFile)
			} else {
				x64Libs[lib] = &ElfFileCount{count: 1, files: []string{elfFile}}
			}
		}
	}

	x32Files := make([]Files, 0, len(x32Libs))
	for lib := range x32Libs {
		x32Files = append(x32Files, Files{Count: x32Libs[lib].count, Files: x32Libs[lib].files, Lib: lib})
	}

	x64Files := make([]Files, 0, len(x64Libs))
	for lib := range x64Libs {
		x64Files = append(x64Files, Files{Count: x64Libs[lib].count, Files: x64Libs[lib].files, Lib: lib})
	}

	sort.Slice(x32Files, func(i, j int) bool {
		return x32Files[i].Count > x32Files[j].Count
	})

	sort.Slice(x64Files, func(i, j int) bool {
		return x64Files[i].Count > x64Files[j].Count
	})

	resText := ""
	if format == outputFormatTXT {
		if len(x32Files) > 0 {
			resText += "---------- i386 (x86) ----------\n"
			resText += RenderTXT(x32Files)
			resText += "\n"
		}
		if len(x64Files) > 0 {
			resText += "---------- x86_64 (amd64) ----------\n"
			resText += RenderTXT(x64Files)
		}
	} else if format == outputFormatJSON {
		jsonFiles := RenderJsonFile{X32: x32Files, X64: x64Files}
		json, err := json.Marshal(jsonFiles)
		if err != nil {
			return err
		}
		resText = string(json)
	}

	if len(output) > 0 {
		err := os.WriteFile(output, []byte(resText), 0644)
		if err != nil {
			return err
		}
	} else {
		fmt.Println(resText)
	}

	return nil
}

func RenderTXT(f []Files) string {
	res := ""
	for _, file := range f {
		res += fmt.Sprintf("%s (%d execs)\n", file.Lib, file.Count)
		for _, elfFile := range file.Files {
			res += fmt.Sprintf("\t%s\n", elfFile)
		}
	}
	return res
}

type RenderJsonFile struct {
	X32 []Files `json:"x32"`
	X64 []Files `json:"x64"`
}
