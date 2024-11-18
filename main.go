package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

func main() {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon ECS service client
	client := ecs.NewFromConfig(cfg)

	// Return a list of existing clusters
	outputListClusters, err := client.ListClusters(context.TODO(), &ecs.ListClustersInput{
		MaxResults: aws.Int32(1),
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, arn := range outputListClusters.ClusterArns {
		log.Printf("Cluster ARN: %s", arn)
	}

	cluster := strings.Split(outputListClusters.ClusterArns[0], "/")[1]
	log.Printf("Cluster: %s", cluster)

	outputListServices, err := client.ListServices(context.TODO(), &ecs.ListServicesInput{
		Cluster:    &cluster,
		MaxResults: aws.Int32(1),
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, arn := range outputListServices.ServiceArns {
		log.Printf("Service ARN: %s", arn)
	}

	service := strings.Split(outputListServices.ServiceArns[0], "/")[1]
	log.Printf("Service: %s", service)

	log.Printf("Setting desired count to 0...")
	_, err = client.UpdateService(context.TODO(), &ecs.UpdateServiceInput{
		Service:      &service,
		Cluster:      &cluster,
		DesiredCount: aws.Int32(0),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Wait for running tasks to go to 0
	taskArns := []string{""}
	for len(taskArns) > 0 {
		outputListTasksOutput, err := client.ListTasks(context.TODO(), &ecs.ListTasksInput{
			ServiceName:   &service,
			Cluster:       &cluster,
			DesiredStatus: "Running",
			MaxResults:    aws.Int32(1),
		})
		if err != nil {
			log.Fatal(err)
		}
		taskArns = outputListTasksOutput.TaskArns
		log.Printf("Waiting for task count to go to 0...")
		time.Sleep(5 * time.Second)
	}

	log.Printf("Setting desired count to 1...")
	_, err = client.UpdateService(context.TODO(), &ecs.UpdateServiceInput{
		Service:      &service,
		Cluster:      &cluster,
		DesiredCount: aws.Int32(1),
	})
	if err != nil {
		log.Fatal(err)
	}
}
