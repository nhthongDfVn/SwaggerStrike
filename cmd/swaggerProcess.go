package cmd

import (
	"context"
	"fmt"
	"strings"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"errors"
	"encoding/json"
	"os"
)

var doc *openapi3.T
var docErr error



func processSwaggerFile(path string, mode string){
	var loader *openapi3.Loader

	if checkIsOpenAPIVersion(path) == false {
		printTask("[+] May be Swagger 2, we will convert it to OpenAPI version 3")
		var doc2 openapi2.T

		doc2 = parseSwagger(path)

		doc, docErr = openapi2conv.ToV3(&doc2)
		if docErr != nil {
			panic(docErr)
		}

		outputFile :=  "convert-swagger3.json"


		json, err := json.Marshal(doc)
		  if err != nil {
		}
		os.WriteFile(outputFile, json, 0666)
		path = outputFile
	} 

	

	isValid, errRead := checkIfFileExists(path)

	if isValid == false {
		panic(errRead)
	}
	
	ctx := context.Background()

	loader = &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}

	doc, docErr = loader.LoadFromFile(path)
		
	if docErr != nil {
		panic(docErr)
	}

	if err := doc.Validate(loader.Context); err != nil {
		panic(err)
	}

	// expand $ref
	if err := loader.ResolveRefsIn(doc, nil); err != nil {
		panic(err)
	}

	printTask("References resolved successfully")

	ExtractSecuritySchemes(&doc.Components.SecuritySchemes)
	getServerList(&doc.Servers)

	for path, pathItem := range doc.Paths.Map() {
		if SpecificPath != "ALL" {
			sPathArray := strings.Split(SpecificPath, ",")
			if StringInArray(path,sPathArray) == false{
				continue
			}
		}
		
		for method, operation := range pathItem.Operations(){
			method = strings.ToUpper(method)
			parameters := ParameterProcess(operation)

			if method == "POST" || method == "PUT" || method == "PATCH"{
			 	if operation.RequestBody == nil || operation.RequestBody.Value == nil {
					fmt.Println("No request body specified for this operation")
					return
				}

				for contentType, mediaType := range operation.RequestBody.Value.Content {
					if mediaType.Schema != nil {
						requestBody:= ExtractBodyValues(mediaType)
						for _, baseURL := range TargetURL{
							CallAction(baseURL,path,parameters,requestBody,contentType,method,mode)
						}
					}
					
				}

			} else {
				for _, baseURL := range TargetURL{
					CallAction(baseURL,path,parameters,nil,"",method,mode)
				}
			}
		}
	}
}

func CallAction(baseURL string,path string, parameters []interface{}, requestBody []interface{}, contentType string, method string,mode string){
	if method == "POST" || method == "PUT" || method == "PATCH" {
		if mode == "listall"{
			if len(config.Profiles) > 1{
				CurrentProfile = config.Profiles[0].Name
			}
			listAllStatus(baseURL,path,parameters,requestBody,contentType,method)
			CurrentProfile = ""
		} else if mode == "unauth"{
			UnauthCheckFunc(baseURL,path,parameters,requestBody,contentType,method)
		} else if mode == "idor"{
			if len(config.Profiles) > 1{
				CurrentProfile = config.Profiles[0].Name
			}
			sequentialFuzz(baseURL,path,parameters,requestBody,contentType,method)
			CurrentProfile = ""
		} else if mode == "privilege" {
			authorizationFuzz(baseURL,path,parameters,requestBody,contentType,method)
		}
	} else {
		if mode == "listall"{
			if len(config.Profiles) > 1{
				CurrentProfile = config.Profiles[0].Name
			}
			listAllStatus(baseURL,path,parameters,nil,"",method)
			CurrentProfile = ""
		} else if mode == "unauth"{
			UnauthCheckFunc(baseURL,path,parameters,nil,"",method)
		} else if mode == "idor"{
			if len(config.Profiles) > 1{
				CurrentProfile = config.Profiles[0].Name
			}
			sequentialFuzz(baseURL,path,parameters,nil,"",method)
			CurrentProfile = ""
		} else if mode == "privilege" {
			authorizationFuzz(baseURL,path,parameters,nil,"",method)
		}
	}

}


func ParameterProcess(operation *openapi3.Operation)([]interface{}){
	var array []interface{}

	for _, param := range operation.Parameters{
		if param.Ref != "" {
			extractArray := ExtractAllParameters(param.Ref)

			for _,data := range extractArray{
				array = append(array,data)
			}
			
		} else {
			if param.Value.Schema.Ref != "" {
				fmt.Println("[P] Ref: ", param.Value.Schema.Ref)
			} else {
				array = append(array,updateReqParam(param.Value.In,param.Value.Name,param.Value.Schema.Value.Type,param.Value.Schema.Value.Format))
			}
		}
		
	}

	return array

}


func ExtractBodyValues(mediaType *openapi3.MediaType)([]interface{}) {
	var array []interface{}

	if mediaType.Schema != nil {
		if mediaType.Schema.Value != nil && mediaType.Schema.Value.Type == "object" {
				for propName, propSchema := range mediaType.Schema.Value.Properties {
					if propSchema.Ref != ""{
						referencedSchema := propSchema.Ref
						if strings.Contains(referencedSchema,"components/schemas/"){
							schemaArray := ExtractAllSchemas(referencedSchema)
							var rParam RequestParam
							rParam.Name = propName
							rParam.Type = "Appitems"

							rParam.Value = schemaArray
							array = append(array,rParam)
						} else {
							fmt.Println("HAS A ERROR WHEN EXTRACT",propSchema.Value.Items)
						}

					} else if propSchema.Value.Type == "array"{
						if propSchema.Value.Items != nil && propSchema.Value.Items.Ref != ""{
							referencedSchema := propSchema.Value.Items.Ref

							if strings.Contains(referencedSchema,"components/schemas/"){
								schemaArray := ExtractAllSchemas(referencedSchema)
								var rParam RequestParam
								rParam.Name = propName
								rParam.Type = "Apparray"

								rParam.Value = schemaArray
								array = append(array,rParam)
							} else {
								fmt.Println("HAS A ERROR WHEN EXTRACT",propSchema.Value.Items)
							}
							
						} else {
							array = append(array,updateReqParam("1340",propName,"array",propSchema.Value.Items.Value.Format))
						}

					} else {
						array = append(array,updateReqParam("1339",propName,propSchema.Value.Type,propSchema.Value.Format))
					}
				}
			} else if mediaType.Schema.Value.Type == "array"{
				referencedSchema := mediaType.Schema.Value.Items.Ref
				schemaArray := ExtractAllSchemas(referencedSchema)
				var rParam RequestParam
				rParam.Name = "hehe"
				rParam.Type = "ApparrayItems"

				rParam.Value = schemaArray
				array = append(array,rParam)
			} else if mediaType.Schema.Value.OneOf != nil   {
				for index := range mediaType.Schema.Value.OneOf{
					ref := mediaType.Schema.Value.OneOf[index].Ref
					tempArray := ExtractAllSchemas(ref)

					for _, v := range tempArray{
						var tempParam RequestParam
						tempParam = v.(RequestParam)
						tempParam.In = "body"
						array = append(array, tempParam)
					}
				}

			} else if mediaType.Schema.Value.AnyOf != nil {
				for index := range mediaType.Schema.Value.AnyOf{
					ref := mediaType.Schema.Value.AnyOf[index].Ref
					tempArray := ExtractAllSchemas(ref)
					
					for _, v := range tempArray{
					 	var tempParam RequestParam
					 	tempParam = v.(RequestParam)
					 	tempParam.In = "body"
					 	array = append(array, tempParam)
					}
				}
			} else if mediaType.Schema.Value.AllOf != nil{
				for index := range mediaType.Schema.Value.AllOf{
					ref := mediaType.Schema.Value.AllOf[index].Ref
					tempArray := ExtractAllSchemas(ref)

					for _, v := range tempArray{
						var tempParam RequestParam
						tempParam = v.(RequestParam)
						tempParam.In = "body"
						array = append(array, tempParam)
					}
				}
			} else {
				fmt.Println("No object schema defined for this content type")
			}
		} else {
			fmt.Println("No schema defined for this content type")
		}

	return array

}


func extractSchema(ref string, doc *openapi3.T) (*openapi3.SchemaRef, error) {
  
  parts := strings.SplitN(ref, "#/", 2)
  schemaName := strings.Split(parts[1], "components/schemas/")

  if len(parts) != 2 {
    return nil, errors.New("invalid $ref format")
  }

  if parts[0] == "" { // Reference within the same document
    if parts[1] == "" {
      return nil, errors.New("empty $ref fragment")
    }
    switch parts[1] {
    case "components/schemas/" + schemaName[1]:
      return doc.Components.Schemas[schemaName[1]], nil
    default:
      return nil, fmt.Errorf("unsupported $ref location: %s", parts[1])
    }
  } else {
    // Handle external references (not supported by kin-openapi)
    return nil, errors.New("external $ref not supported")
  }
}


func extractParameter(ref string, doc *openapi3.T) (*openapi3.ParameterRef, error) {
  
  parts := strings.SplitN(ref, "#/", 2)
  schemaName := strings.Split(parts[1], "components/parameters/")

  if len(parts) != 2 {
    return nil, errors.New("invalid $ref format")
  }

  if parts[0] == "" { // Reference within the same document
    if parts[1] == "" {
      return nil, errors.New("empty $ref fragment")
    }
    switch parts[1] {
    case "components/parameters/" + schemaName[1]:
      return doc.Components.Parameters[schemaName[1]], nil
    default:
      // Handle other potential locations based on your document structure
      return nil, fmt.Errorf("unsupported $ref location: %s", parts[1])
    }
  } else {
    // Handle external references (not supported by kin-openapi)
    return nil, errors.New("external $ref not supported")
  }
}


func ExtractAllParameters(ref string) ([]interface{}){
	var array []interface{}
	
	parameter, err := extractParameter(ref, doc)

	if err != nil {
		return []interface{}{nil}
	}

	data_in := parameter.Value.In

	tempArray := ExtractAllSchemas(parameter.Value.Schema.Ref)

	for _, v := range tempArray{
		var tempParam RequestParam
		tempParam = v.(RequestParam)
		tempParam.In = data_in
		array = append(array, tempParam)
	}

	return array
}



func ExtractAllSchemas(ref string) ([]interface{}){
	var array []interface{}
	schema, err := extractSchema(ref,doc)

	if err != nil {
		return []interface{}{nil}
	}
	if schema.Value.Type == "object"{
		for paramName, paramType := range schema.Value.Properties{
			
			if paramType.Value.Type == "array"{
				if strings.Contains(paramType.Value.Items.Ref,"components/schemas/"){
					var rParam RequestParam
					rParam.Name = paramName
					rParam.Type = "array"

					schemaResult := ExtractAllSchemas(paramType.Value.Items.Ref)

					rParam.Value = schemaResult
					array = append(array,rParam)
				} else {
					array = append(array,updateReqParam("1338",paramName,"array",paramType.Value.Items.Value.Format))
				}


			} else {
				array = append(array,updateReqParam("1337",paramName,paramType.Value.Type,paramType.Value.Format))
			}

		}
	}
	

	return array
}

func ExtractSecuritySchemes(securitySchemas *openapi3.SecuritySchemes){

	hasSecuritySchema := false

	
	for _, schema := range *securitySchemas {
		hasSecuritySchema = true
		if schema.Ref != ""{
			fmt.Println("[Layer 2]Security schema need to extract $ref: ", schema.Ref)
		} else {
			if schema.Value.Type == "http"{
				printTask("Type: "+ schema.Value.Type + " Schema:  " + schema.Value.Scheme)
			} else if schema.Value.Type == "apiKey"{
				printTask("Name: "+ schema.Value.Name + " In:  " + schema.Value.In)
			} else if schema.Value.Type == "oauth2"{
				printTask("Name: "+ schema.Value.Name + " In:  " + schema.Value.In)
			} else if schema.Value.Type == "openIdConnect"{
				printTask("Name: "+ schema.Value.Name + " In:  " + schema.Value.In)
			}

		}			
	}

	if hasSecuritySchema == true {
		printTask("Detect security schema. Please add for authentication process.")
	}

	printTask("----------------------------------")

}


func updateReqParam(valueIn string, valueName string,valueType string, valueFormat string)(RequestParam){
	var rParam RequestParam

	rParam.In = valueIn
	rParam.Name = valueName
	rParam.Type = valueType
	rParam.Format = valueFormat

	return rParam
}