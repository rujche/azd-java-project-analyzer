package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"ajpa/analyzer"
	"ajpa/converter"

	"github.com/azure/azure-dev/cli/azd/pkg/project"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	cwd := flag.String("cwd", dir, "change working directory")
	// todo: add other flags like:
	// 1. output dir.
	// 2. output to console.
	// 3. log level.
	flag.Parse()

	result, err := analyzer.AnalyzeJavaProject(*cwd)
	if err != nil {
		fmt.Println(err)
		return
	}
	config, err := converter.ProjectAnalysisResultToAzdProjectConfig(result)
	if err != nil {
		fmt.Println(err)
		return
	}
	path := filepath.Join(*cwd, "azure.yaml")
	err = project.Save(context.TODO(), &config, path)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("File generated: ", path)
}
