package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/olekukonko/tablewriter"
	"google.golang.org/protobuf/types/descriptorpb"
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
	table.SetHeader([]string{"Method", "Path", "OperationID"})

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
	table.SetHeader([]string{"File", "Service", "Method", "Input Type", "Output Type", "Streaming"})
	for _, protoFile := range protoFiles {
		parser := protoparse.Parser{
			// required for google proto files
			ImportPaths:           []string{"."},
			IncludeSourceCodeInfo: true,
			InferImportPaths:      true,
		}
		fds, err := parser.ParseFiles(protoFile)
		if err != nil {
			return fmt.Errorf("error parsing Proto file %s: %v", protoFile, err)
		}

		for _, file := range fds {
			for _, service := range file.GetServices() {
				for _, method := range service.GetMethods() {
					descriptor := method.AsMethodDescriptorProto()
					streaming := "No"
					if descriptor.GetClientStreaming() || descriptor.GetServerStreaming() {
						streaming = "Yes"
					}

					table.Append([]string{
						file.GetName(),
						service.GetName(),
						method.GetName(),
						descriptor.GetInputType(),
						descriptor.GetOutputType(),
						streaming,
					})
				}
			}
		}
	}

	table.Render()
	return nil
}

// generateGrpcurlCommand generates a grpcurl command for a given service and method
func GrpCurlCommand(protoFile, serviceName, methodName string) error {
	var grpCurl string
	parser := protoparse.Parser{
		ImportPaths:           []string{"."},
		IncludeSourceCodeInfo: true,
		InferImportPaths:      true,
	}

	fds, err := parser.ParseFiles(protoFile)
	if err != nil {
		return fmt.Errorf("error parsing Proto file %s: %v", protoFile, err)
	}

	serviceFound := false
	methodFound := false
	for _, file := range fds {
		for _, service := range file.GetServices() {
			if service.GetName() == serviceName {
				serviceFound = true
				for _, method := range service.GetMethods() {
					if method.GetName() == methodName {
						message, err := createJSONRequestBody(method.GetInputType().AsDescriptorProto().GetField())
						if err != nil {
							return fmt.Errorf("error creating JSON request body: %v", err)
						}
						grpCurl = fmt.Sprintf("grpcurl -plaintext -proto %s -d '%s' localhost:50051 %s/%s",
							protoFile, message, service.GetFullyQualifiedName(), method.GetName())
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

func createJSONRequestBody(fields []*descriptorpb.FieldDescriptorProto) (string, error) {
	fieldsMap := make(map[string]interface{})
	fmt.Println(fields)
	for _, field := range fields {
		if field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
			fieldsMap[field.GetJsonName()] = []interface{}{}
		} else {
			fieldsMap[field.GetJsonName()] = ""
		}
	}
	jsonData, err := json.Marshal(fieldsMap)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
