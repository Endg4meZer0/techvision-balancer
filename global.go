package main

import "github.com/docker/docker/client"

var GlobalNodes Nodes = Nodes{}

var DockerClient *client.Client