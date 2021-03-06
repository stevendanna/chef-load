package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-chef/chef"
)

func chefClientRun(nodeClient chef.Client, nodeName string, runList []string, getCookbooks bool, apiGetRequests []string, sleepDuration int) {
	node, err := nodeClient.Nodes.Get(nodeName)
	if err != nil {
		statusCode := getStatusCode(err)
		if statusCode == 404 {
			// Create a Node object
			// TODO: should have a constructor for this
			node = chef.Node{
				Name:                nodeName,
				Environment:         "_default",
				ChefType:            "node",
				JsonClass:           "Chef::Node",
				RunList:             []string{},
				AutomaticAttributes: map[string]interface{}{},
				NormalAttributes:    map[string]interface{}{},
				DefaultAttributes:   map[string]interface{}{},
				OverrideAttributes:  map[string]interface{}{},
			}

			_, err = nodeClient.Nodes.Post(node)
			if err != nil {
				fmt.Println("Couldn't create node. ", err)
			}
		} else {
			fmt.Println("Couldn't get node: ", err)
		}
	}

	rl := map[string][]string{"run_list": runList}
	data, err := chef.JSONReader(rl)
	if err != nil {
		fmt.Println(err)
	}

	var cookbooks map[string]json.RawMessage

	err = func(cookbooks *map[string]json.RawMessage) error {
		req, err := nodeClient.NewRequest("POST", "environments/_default/cookbook_versions", data)
		res, err := nodeClient.Do(req, nil)
		if err != nil {
			// can't print res here if it is nil
			// fmt.Println(res.StatusCode)
			fmt.Println(err)
			return err
		}
		defer res.Body.Close()

		return json.NewDecoder(res.Body).Decode(cookbooks)
	}(&cookbooks)

	if getCookbooks {
		downloadCookbooks(&nodeClient, cookbooks)
	}

	for _, apiGetRequest := range apiGetRequests {
		req, _ := nodeClient.NewRequest("GET", apiGetRequest, nil) //, data)
		res, err := nodeClient.Do(req, nil)
		if err != nil {
			// can't print res here if it is nil
			// fmt.Println(res.StatusCode)
			// TODO: should this be handled better than just skipping over it?
			fmt.Println(err)
			continue
		}
		defer res.Body.Close()
		// res.Body.Close()
	}

	time.Sleep(time.Duration(sleepDuration) * time.Second)

	_, err = nodeClient.Nodes.Put(node)
	if err != nil {
		fmt.Println("Couldn't update node: ", err)
	}
}
