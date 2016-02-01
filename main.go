package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

const AppVersion = "0.1.1"

var quit = make(chan int)

func main() {
	fConfig := flag.String("config", "", "Configuration file to load")
	fHelp := flag.Bool("help", false, "Print this help")
	fNodes := flag.String("nodes", "", "Number of nodes making chef-client runs")
	fRuns := flag.String("runs", "", "Number of chef-client runs each node should make, 0 value will make infinite runs")
	fSampleConfig := flag.Bool("sample-config", false, "Print out full sample configuration")
	fVersion := flag.Bool("version", false, "Print chef-load version")
	flag.Parse()

	if *fHelp {
		fmt.Println("Usage of chef-load:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *fVersion {
		fmt.Println("chef-load", AppVersion)
		os.Exit(0)
	}

	if *fSampleConfig {
		printSampleConfig()
		os.Exit(0)
	}

	var (
		config *chefLoadConfig
		err    error
	)

	if *fConfig != "" {
		config, err = loadConfig(*fConfig)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Usage of chef-load:")
		flag.PrintDefaults()
		return
	}

	if *fNodes != "" {
		config.Nodes, _ = strconv.Atoi(*fNodes)
	}

	if *fRuns != "" {
		config.Runs, _ = strconv.Atoi(*fRuns)
	}

	// Early exit if we can't read the client_key, avoiding a messy stacktrace
	f, err := os.Open(config.ClientKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read client key %v:\n\t%v\n", config.ClientKey, err)
		os.Exit(1)
	}
	f.Close()

	var data map[string]interface{}
	if config.OhaiData == "" {
		fmt.Fprintf(os.Stderr, "No ohai data provide, nearly empty node objects will be used!\n")
	} else {
		data, err = read_ohai_data(config.OhaiData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}

	numNodes := config.Nodes
	for i := 0; i < numNodes; i++ {
		nodeName := config.NodeNamePrefix + "-" + strconv.Itoa(i)
		go startNode(nodeName, *config, data)
	}
	for i := 0; i < numNodes; i++ {
		<-quit // Wait to be told to exit.
	}
}

func read_ohai_data(path string) (map[string]interface{}, error) {
	var data map[string]interface{}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open specified ohai data (%v): %v", path, err)
	}
	defer f.Close()

	parser := json.NewDecoder(f)
	err = parser.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("could not parse ohai data (%v):%v", path, err)
	}
	return data, nil
}
