# Starting and Getting Responses from Temporal Workflows in Go: A Developer's Guide

This guide provides a practical pathway for Go developers to effectively start Temporal workflow executions and retrieve their responses, covering essential concepts and implementation steps.

## Introduction to Temporal

Temporal is a distributed, scalable orchestration engine that helps you build and run reliable workflows for your applications. It handles state persistence, automatic retries, and complex coordination logic between services[2]. The platform consists of a programming framework (client library) and a managed service (backend)[2].

## Setting Up Your Environment

### Installing the Temporal Go SDK

```bash
go get go.temporal.io/sdk
```

### Connecting to Temporal Service

```go
import (
    "go.temporal.io/sdk/client"
)

// Create a Temporal Client to communicate with the Temporal Service
temporalClient, err := client.Dial(client.Options{
    HostPort: client.DefaultHostPort, // Defaults to "127.0.0.1:7233"
})
if err != nil {
    log.Fatalln("Unable to create Temporal Client", err)
}
defer temporalClient.Close()
```

For Temporal Cloud connections:
```go
// For Temporal Cloud
temporalClient, err := client.Dial(client.Options{
    HostPort:  "your-namespace.tmprl.cloud:7233",
    Namespace: "your-namespace",
    ConnectionOptions: client.ConnectionOptions{
        TLS: &tls.Config{},
    },
})
```

## Defining a Simple Workflow

```go
import (
    "time"
    "go.temporal.io/sdk/workflow"
)

// Define your workflow function
func GreetingWorkflow(ctx workflow.Context, name string) (string, error) {
    // Set activity options
    ao := workflow.ActivityOptions{
        TaskQueue:              "greeting-tasks",
        StartToCloseTimeout:    time.Minute,
        ScheduleToCloseTimeout: time.Minute,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // Execute activity and get result
    var result string
    err := workflow.ExecuteActivity(ctx, GreetingActivity, name).Get(ctx, &result)
    if err != nil {
        return "", err
    }
    
    return result, nil
}

// Define your activity function
func GreetingActivity(ctx context.Context, name string) (string, error) {
    return "Hello, " + name + "!", nil
}
```

## Creating a Worker

```go
import (
    "go.temporal.io/sdk/worker"
)

func startWorker(c client.Client) {
    // Create worker options
    w := worker.New(c, "greeting-tasks", worker.Options{})
    
    // Register workflow and activity with the worker
    w.RegisterWorkflow(GreetingWorkflow)
    w.RegisterActivity(GreetingActivity)
    
    // Start the worker
    err := w.Run(worker.InterruptCh())
    if err != nil {
        log.Fatalln("Unable to start worker", err)
    }
}
```

## Starting Workflow Executions

```go
// Define workflow options
workflowOptions := client.StartWorkflowOptions{
    ID:        "greeting-workflow-" + uuid.New().String(),
    TaskQueue: "greeting-tasks",
}

// Start the workflow execution
workflowRun, err := temporalClient.ExecuteWorkflow(
    context.Background(), 
    workflowOptions, 
    GreetingWorkflow, 
    "Temporal Developer"
)
if err != nil {
    log.Fatalln("Unable to execute workflow", err)
}

// Get workflow ID and run ID for future reference
fmt.Printf("Started workflow: WorkflowID: %s, RunID: %s\n", 
    workflowRun.GetID(), 
    workflowRun.GetRunID())
```

## Getting Responses from Workflow Executions

### 1. Synchronous Response

```go
// Wait for workflow completion and get result
var result string
err = workflowRun.Get(context.Background(), &result)
if err != nil {
    log.Fatalln("Unable to get workflow result", err)
}
fmt.Printf("Workflow result: %s\n", result)
```

### 2. Retrieving Results Later

```go
// Get workflow result using workflow ID and run ID
workflowID := "greeting-workflow-123"
runID := "run-id-456"

// Retrieve the workflow handle
workflowRun = temporalClient.GetWorkflow(context.Background(), workflowID, runID)

// Get the result
var result string
err = workflowRun.Get(context.Background(), &result)
if err != nil {
    log.Fatalln("Unable to get workflow result", err)
}
```

### 3. Using Queries to Get Workflow State

```go
// Define query handler in your workflow
func GreetingWorkflow(ctx workflow.Context, name string) (string, error) {
    // Set up state variable
    greeting := ""
    
    // Register query handler
    err := workflow.SetQueryHandler(ctx, "getGreeting", func() (string, error) {
        return greeting, nil
    })
    if err != nil {
        return "", err
    }
    
    // Workflow logic...
    greeting = "Hello, " + name + "!"
    
    return greeting, nil
}

// Query the workflow state from client
response, err := temporalClient.QueryWorkflow(context.Background(), 
    workflowID, runID, "getGreeting")
if err != nil {
    log.Fatalln("Unable to query workflow", err)
}

var greeting string
err = response.Get(&greeting)
if err != nil {
    log.Fatalln("Unable to decode query result", err)
}
fmt.Printf("Current greeting: %s\n", greeting)
```

### 4. Message Passing with Signals

```go
// In your workflow, set up a signal channel
func GreetingWorkflow(ctx workflow.Context, name string) (string, error) {
    // Create signal channel
    updateNameChannel := workflow.GetSignalChannel(ctx, "update_name")
    
    for {
        // Wait for signal or timeout
        selector := workflow.NewSelector(ctx)
        selector.AddReceive(updateNameChannel, func(c workflow.ReceiveChannel, more bool) {
            var newName string
            c.Receive(ctx, &newName)
            name = newName
            // Process updated name...
        })
        
        // Add timeout to exit workflow
        selector.Select(ctx)
    }
}

// Send signal to workflow
err = temporalClient.SignalWorkflow(context.Background(), 
    workflowID, runID, "update_name", "New Name")
if err != nil {
    log.Fatalln("Unable to signal workflow", err)
}
```

## Error Handling and Retries

```go
// Configure retry policy
retryPolicy := &temporal.RetryPolicy{
    InitialInterval:    time.Second,
    BackoffCoefficient: 2.0,
    MaximumInterval:    time.Minute * 5,
    MaximumAttempts:    5,
}

// Apply retry policy to activity options
ao := workflow.ActivityOptions{
    TaskQueue:              "greeting-tasks",
    StartToCloseTimeout:    time.Minute,
    ScheduleToCloseTimeout: time.Minute,
    RetryPolicy:            retryPolicy,
}
```

## Getting Workflow Information

```go
// Inside a workflow, get workflow execution info
info := workflow.GetInfo(ctx)
workflowID := info.WorkflowExecution.ID
runID := info.WorkflowExecution.RunID[7]
```

## Workflow Run ID

To get the current run ID within a workflow (useful for self-termination)[7]:
```go
// Inside a workflow
runID := workflow.GetInfo(ctx).WorkflowExecution.RunID
```

## Conclusion

This guide provides the essential steps to start and get responses from Temporal workflow executions in Go. Temporal offers a powerful framework for building reliable, distributed applications with durable execution state. For more advanced features, refer to the official Temporal documentation and explore the sample applications[12].

Remember that Temporal is particularly valuable for scenarios involving:
- Long-running, potentially multi-step processes
- Coordination between multiple services
- Processes requiring automatic retries
- Workflows that need to maintain state even through system failures[14]

By leveraging Temporal's fault-tolerance capabilities, you can build applications that reliably execute complex business logic while focusing on your business requirements rather than infrastructure concerns.

Sources
[1] Go SDK developer guide | Temporal Platform Documentation https://docs.temporal.io/develop/go
[2] temporal - Go Packages https://pkg.go.dev/go.temporal.io/sdk/temporal
[3] Workflow message passing - Go SDK - Temporal Docs https://docs.temporal.io/develop/go/message-passing
[4] temporalio/sdk-go: Temporal Go SDK - GitHub https://github.com/temporalio/sdk-go
[5] Temporal Client - Go SDK https://docs.temporal.io/develop/go/temporal-clients
[6] README.md - Temporal Go SDK samples - GitHub https://github.com/temporalio/samples-go/blob/main/README.md
[7] Temporal, How to get RunID while being inside a workflow to ... https://stackoverflow.com/questions/73229921/temporal-how-to-get-runid-while-being-inside-a-workflow-to-terminate-the-curren
[8] Run your first Temporal application with the Go SDK https://learn.temporal.io/getting_started/go/first_program_in_go/
[9] Go SDK developer guide | Temporal Platform Documentation https://docs.temporal.io/develop/go/
[10] Build a Temporal Application from scratch in Go https://learn.temporal.io/getting_started/go/hello_world_in_go/
[11] workflow package - go.temporal.io/sdk/workflow - Go Packages https://pkg.go.dev/go.temporal.io/sdk/workflow
[12] temporalio/samples-go: Temporal Go SDK samples - GitHub https://github.com/temporalio/samples-go
[13] workflow - Go Packages https://pkg.go.dev/go.temporal.io/temporal/workflow
[14] When to use a Workflow tool (Temporal) vs a Job Queue - Reddit https://www.reddit.com/r/golang/comments/1as23yb/when_to_use_a_workflow_tool_temporal_vs_a_job/
[15] workflowcheck command - go.temporal.io/sdk/contrib/tools ... https://pkg.go.dev/go.temporal.io/sdk/contrib/tools/workflowcheck
[16] Implementing Temporal IO in Golang Microservices Architecture https://www.softwareletters.com/p/implementing-temporal-io-golang-microservices-architecture-stepbystep-guide
[17] Using Temporal and Go SDK for flows orchestration : r/golang - Reddit https://www.reddit.com/r/golang/comments/1dy2np1/using_temporal_and_go_sdk_for_flows_orchestration/
[18] Temporal Workflow | Temporal Platform Documentation https://docs.temporal.io/workflows
[19] Intro to Temporal with Go SDK - YouTube https://www.youtube.com/watch?v=-KWutSkFda8
[20] Temporal SDK : r/golang - Reddit https://www.reddit.com/r/golang/comments/15kwzke/temporal_sdk/
[21] Core application - Go SDK | Temporal Platform Documentation https://docs.temporal.io/develop/go/core-application
[22] Get started with Temporal and Go https://learn.temporal.io/getting_started/go/
[23] temporal: when testing, how do I pass context into workflows and ... https://stackoverflow.com/questions/69577516/temporal-when-testing-how-do-i-pass-context-into-workflows-and-activities
[24] Workflow with Temporal - Capten.AI https://capten.ai/learning-center/10-learn-temporal/understand-temporal-workflow/workflow/
[25] client package - go.temporal.io/sdk/client - Go Packages https://pkg.go.dev/go.temporal.io/sdk/client
[26] documentation-samples-go/yourapp/your_workflow_definition_dacx ... https://github.com/temporalio/documentation-samples-go/blob/main/yourapp/your_workflow_definition_dacx.go
