package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/akamensky/argparse"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

var (
	parser            = argparse.NewParser("civars", "Prints provided string to stdout")
	token             = parser.String("t", "token", &argparse.Options{Required: true, Help: "Gitlab Token"})
	projectId         = parser.String("p", "projectId", &argparse.Options{Required: true, Help: "Gitlab Project ID"})
	scope             = parser.String("s", "scope", &argparse.Options{Required: true, Help: "EnvironmentScope for variables"})
	outputDir         = parser.String("o", "dir", &argparse.Options{Required: false, Help: "Output directory, if not set CWD will be used."})
	logLevel  *string = parser.Selector("d", "log-level", []string{"INFO", "DEBUG", "WARN", "ERROR"}, nil)
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)

}

func WriteOutput(path string, filename string, value string) error {
	file := filepath.Join(path, filename)

	// If the directory does not exist, create it
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err2 := f.Write([]byte(value))
	if err2 != nil {
		return err2
	}

	return nil

}

func main() {

	// Parse the command line arguments
	err := parser.Parse(os.Args)

	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		log.Error(parser.Usage(err))
	}

	// set log level depending on env
	switch strings.ToLower(*logLevel) {
	case "debug":
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}

	// Initialize a connection to the Gitlab API
	log.Debugf("Using token %s", *token)
	git, err := gitlab.NewClient(*token)
	if err != nil {
		// log.Fatal(err)
		log.Fatal(err)

	}

	// Get the project by id to get the parent group
	project, _, err := git.Projects.GetProject(*projectId, &gitlab.GetProjectOptions{})
	if err != nil {
		log.Fatal(err)
	}

	group, _, err := git.Groups.GetGroup(project.Namespace.ID, &gitlab.GetGroupOptions{})
	if err != nil {
		log.Fatalf("Failed to get parent group: %v", err)
	}

	log.Debugf("Found group: %v", group.FullName)

	//List all projects group
	projectsInGroup, _, err := git.Groups.ListGroupProjects(group.ID, &gitlab.ListGroupProjectsOptions{})
	if err != nil {
		log.Fatal("Failed to get projects in group: %v", err)
	}

	if log.GetLevel() == log.DebugLevel {
		for _, projectInGroup := range projectsInGroup {
			log.Debugf("Project: %v", projectInGroup.Name)
		}
	}
	var wg sync.WaitGroup
	// Get length of slice projectsInGroup to use in wait group for Go routines
	projectsInGroupLength := len(projectsInGroup)
	// Add slice to wait group
	wg.Add(projectsInGroupLength)
	for _, projectInGroup := range projectsInGroup {
		// Spawn a goroutine per project
		go func(projectInGroup *gitlab.Project) {
			defer wg.Done()
			variables, _, err := git.ProjectVariables.ListVariables(projectInGroup.ID, &gitlab.ListProjectVariablesOptions{})
			if err != nil {
				log.Fatalf("Failed to get variables in project: %v", err)
			}

			for _, variable := range variables {
				if variable.EnvironmentScope == *scope && variable.VariableType == *gitlab.VariableType(gitlab.FileVariableType) {
					log.Debugf("Found variable: %v with scope: %v and type: %v", variable.Key, variable.EnvironmentScope, *gitlab.VariableType(gitlab.FileVariableType))

					n := variable.Key
					err := WriteOutput(*outputDir, n, variable.Value)
					if err != nil {
						log.Fatal(err)
					}

					log.Infof("Wrote key: %v to: %v/%v", variable.Key, *outputDir, n)
				}
			}
		}(projectInGroup)
	}
	wg.Wait()

}
