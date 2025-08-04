package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bufbuild/protocompile"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

type OpenAPI struct {
	Paths map[string]PathItem `json:"paths"`
}

type Operation struct {
	OperationID string `json:"operationId"`
}

type PathItem struct {
	Get    *Operation `json:"get"`
	Post   *Operation `json:"post"`
	Put    *Operation `json:"put"`
	Delete *Operation `json:"delete"`
	Patch  *Operation `json:"patch"`
}

func OpenApiAnalyzer(specFiles []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Method", "Path", "OperationID")

	for _, specFile := range specFiles {
		data, err := os.ReadFile(specFile)
		if err != nil {
			panic(err)
		}

		var openapi OpenAPI
		if err := json.Unmarshal(data, &openapi); err == nil {
		} else if err := yaml.Unmarshal(data, &openapi); err == nil {
		} else {
			panic("Unsupported OpenAPI file format")
		}

		for path, pathItem := range openapi.Paths {
			if pathItem.Get != nil {
				table.Append([]string{"GET", path, pathItem.Get.OperationID})
			}
			if pathItem.Post != nil {
				table.Append([]string{"POST", path, pathItem.Post.OperationID})
			}
			if pathItem.Put != nil {
				table.Append([]string{"PUT", path, pathItem.Put.OperationID})
			}
			if pathItem.Patch != nil {
				table.Append([]string{"PATCH", path, pathItem.Patch.OperationID})
			}
			if pathItem.Delete != nil {
				table.Append([]string{"DELETE", path, pathItem.Delete.OperationID})
			}
		}
	}
	table.Render()
}

func ProtoAnalyzer(protoFiles []string) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("File", "Service", "Method", "Input Type", "Output Type", "Streaming")
	
	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{
			ImportPaths: []string{"."},
		},
	}
	
	ctx := context.Background()
	fds, err := compiler.Compile(ctx, protoFiles...)
	if err != nil {
		return fmt.Errorf("error compiling Proto files: %v", err)
	}

	for _, file := range fds {
		services := file.Services()
		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)
			methods := service.Methods()
			for j := 0; j < methods.Len(); j++ {
				method := methods.Get(j)
				streaming := "No"
				if method.IsStreamingClient() || method.IsStreamingServer() {
					streaming = "Yes"
				}

				table.Append([]string{
					string(file.Path()),
					string(service.Name()),
					string(method.Name()),
					string(method.Input().FullName()),
					string(method.Output().FullName()),
					streaming,
				})
			}
		}
	}

	table.Render()
	return nil
}

// generateGrpcurlCommand generates a grpcurl command for a given service and method
func GrpCurlCommand(protoFile, serviceName, methodName string) error {
	var grpCurl string
	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{
			ImportPaths: []string{"."},
		},
	}

	ctx := context.Background()
	fds, err := compiler.Compile(ctx, protoFile)
	if err != nil {
		return fmt.Errorf("error compiling Proto file %s: %v", protoFile, err)
	}

	serviceFound := false
	methodFound := false
	for _, file := range fds {
		services := file.Services()
		for i := 0; i < services.Len(); i++ {
			service := services.Get(i)
			if string(service.Name()) == serviceName {
				serviceFound = true
				methods := service.Methods()
				for j := 0; j < methods.Len(); j++ {
					method := methods.Get(j)
					if string(method.Name()) == methodName {
						// Create a simple JSON template based on the input message fields
						inputMsg := method.Input()
						fields := inputMsg.Fields()
						fieldsMap := make(map[string]interface{})
						for k := 0; k < fields.Len(); k++ {
							field := fields.Get(k)
							if field.IsList() {
								fieldsMap[string(field.Name())] = []interface{}{}
							} else {
								fieldsMap[string(field.Name())] = ""
							}
						}
						messageJSON, err := json.Marshal(fieldsMap)
						if err != nil {
							return fmt.Errorf("error creating JSON request body: %v", err)
						}
						grpCurl = fmt.Sprintf("grpcurl -plaintext -proto %s -d '%s' localhost:50051 %s/%s",
							protoFile, string(messageJSON), service.FullName(), method.Name())
						methodFound = true
						break
					}
				}
			}
			if serviceFound && methodFound {
				break
			}
		}
		if serviceFound && methodFound {
			break
		}
	}

	if !serviceFound {
		return fmt.Errorf("service %s not found", serviceName)
	}
	if !methodFound {
		return fmt.Errorf("method %s not found in service %s", methodName, serviceName)
	}
	fmt.Println(grpCurl)
	return nil
}

