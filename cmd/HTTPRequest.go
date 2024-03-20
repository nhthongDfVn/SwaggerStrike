
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"crypto/tls"
	"reflect"
	"os"
	"strings"
	"time"
)

var FullRequestParam []string


type HTTPRequest struct {
	Method 	string
	BaseURL string
	Path    string
	Headers	map[string]string
	Cookies map[string]string
	Body map[string]any
	Query map[string]any
	ContentType string
}


func NewHTTPRequest() *HTTPRequest {

	FullRequestParam = []string{}
	return &HTTPRequest{
		Method:      "",
		BaseURL: 	 "/",
		Path:		 "/",
		Headers:     make(map[string]string),
		Cookies:     make(map[string]string),
		Body:        make(map[string]interface{}),
		Query:       make(map[string]interface{}),
		ContentType: "",
	}
}


func GenTransportConfig(skipTLS bool)(*http.Transport,error){

	var transport *http.Transport

	if Proxy != "NOPROXY"{
		proxyURL, errProxy := url.Parse(Proxy)
		   	if errProxy != nil {
		        fmt.Println("Error parsing proxy URL:", errProxy)
		        return nil,errProxy
		    }

		    if skipTLS == true {
		    	transport = &http.Transport{
		    		TLSClientConfig: &tls.Config{
			            InsecureSkipVerify: true,
			        },
	        		Proxy: http.ProxyURL(proxyURL),
	    		}
	    	} else {
	    		transport = &http.Transport{
	        		Proxy: http.ProxyURL(proxyURL),
	    		}
	    	}

	} else {
		if skipTLS == true {
		    transport = &http.Transport{
		    	TLSClientConfig: &tls.Config{
			        InsecureSkipVerify: true,
			    },
	    	}
	    } else {
	    	transport = &http.Transport{}
	    }
	}

	return transport,nil
}



func GenerateRequests(BaseURL string,Path string, Parameter []interface{}, requestBody []interface{}, ContentType string, method string, skipTLS bool)(*http.Response, error){

		var req = NewHTTPRequest()
		var transport *http.Transport
		var errProxy error

		transport,errProxy = GenTransportConfig(skipTLS)

		if errProxy != nil {
			return nil,errProxy
		}

    	client := &http.Client{Transport: transport,Timeout: time.Duration(timeout) * time.Second,}
		var httpRequest *http.Request
		var err error
		var requestBodyData []byte


		if BaseURL == "/"{
			BaseURL = getHostFromURI(swaggerURL)
		}


		for _, param := range Parameter{
			param := param.(RequestParam)
			if param.In == "path"{
				value:= GenerateValue(param.Name,param.Type,param.Format,nil)
				Path = strings.ReplaceAll(Path, "{" +param.Name + "}",ConvertToString(value))
			} else if param.In == "query" {
				if param.Type == "array"{
					param.Type = "string"
					value:= GenerateValue(param.Name,param.Type,param.Format,nil)
					req.Query[param.Name] = value
				} else {
					value:= GenerateValue(param.Name,param.Type,param.Format,nil)
					req.Query[param.Name] = value
				}
				
			} else if param.In == "header"{
				value:= GenerateValue(param.Name,param.Type,param.Format,nil)
				req.Headers[param.Name] = value.(string)
			} else if param.In == "cookie"{
				value:= GenerateValue(param.Name,param.Type,param.Format,nil)
				req.Cookies[param.Name] = value.(string)
			} else {
				printError("!!!!! Note: invalid param In type")
			}
		}


		for _, profile := range config.Profiles {
			if profile.Name == CurrentProfile {

				for key, value := range profile.Header {
					req.Headers[key] = value
				}

				for key, value := range profile.Cookie {
					req.Cookies[key] = value
				}

				for key, value := range profile.Query {
					req.Query[key] = value
				}

				for key, value := range profile.Data {
					req.Body[key] = value
				}
			}
		}

		values := url.Values{}
		for key, value := range req.Query {
			values.Add(key, fmt.Sprintf("%v", value))
		}
		pathQuery := fmt.Sprintf("?%s",values.Encode())
		if len(pathQuery) <= 1{
			pathQuery = ""
		}
		
		if method == "POST" || method == "PUT" || method == "PATCH"{

			for _, reqBody := range requestBody {
				reqBody := reqBody.(RequestParam)
				if reqBody.Type == "array"{
					if checkRequestParamInterface(reqBody.Value) == true{
						value := extractValueRecursive(reqBody.Value)
						arrayTemp := [...]interface{}{value,value}
						req.Body[reqBody.Name] = arrayTemp
					} else {
						value:= GenerateValue(reqBody.Name,"string",reqBody.Format,nil)
						arrayTemp := [...]interface{}{value,value}
						req.Body[reqBody.Name] = arrayTemp
					}
					
					
				} else if reqBody.Type == "Appitems"{
					value := extractValueRecursive(reqBody.Value)
					req.Body[reqBody.Name] = value
				} else if reqBody.Type == "ApparrayItems"{
					value := extractValueRecursive(reqBody.Value)
					req.Body["ApparrayItems"] = value
				} else if reqBody.Type != "Apparray"{
					value:= GenerateValue(reqBody.Name,reqBody.Type,reqBody.Format,nil)
					req.Body[reqBody.Name] = value
				} else {
					value := extractValueRecursive(reqBody.Value)
					arrayTemp := [...]interface{}{value}
					req.Body[reqBody.Name] = arrayTemp
				}
			}

			if req.Body["ApparrayItems"] != nil {
				newArray := [...]interface{}{req.Body["ApparrayItems"]}
				requestBodyData = getBodyByContentType(ContentType,newArray)
			}  else {
				requestBodyData= getBodyByContentType(ContentType,req.Body)
			}  
		}

		if method == "POST" || method == "PUT" || method == "PATCH"{
			httpRequest, err = http.NewRequest(method, BaseURL+ Path + pathQuery,bytes.NewBuffer(requestBodyData))
		} else {
			httpRequest, err = http.NewRequest(method, BaseURL + Path + pathQuery,nil)
		}

		if err != nil{	
			return nil, err
		}

		if ContentType != ""{
			httpRequest.Header.Set("Content-Type", ContentType)
		}

		for key, value := range req.Headers {
			httpRequest.Header.Set(key, value)
		}

		for key, value := range req.Cookies {
			cookie := &http.Cookie{
					Name:  key,
					Value: value,
			}

			httpRequest.AddCookie(cookie)
		}

		if len(UserAgent) > 0 {
			httpRequest.Header.Set("User-Agent", UserAgent)
		}


	httpRespone, err := client.Do(httpRequest)

	if err != nil {
		if ConnectionErrCount > MAX_REQUEST_ERROR {
			printError("TOO MUCH NETWORK ERROR. SKIP PROGRAM")
			os.Exit(1)
		} else {
			ConnectionErrCount = ConnectionErrCount + 1
		}
	}
	return httpRespone,err
}

func getBodyByContentType(contentType string, requestBody interface{})([]byte){
	var result []byte

	switch contentType {
	case "application/json","*/*": 
		var err error
		result, err = json.Marshal(requestBody)
		if err != nil {
			printError(fmt.Sprintf("%s",err))
			return []byte("default body")
		}
	case "application/x-www-form-urlencoded":
		values := url.Values{}

		for key, value := range requestBody.(map[string]interface{}) {
			value1 := ConvertToString(value)
		 	values.Add(key, value1)
		}
		result = []byte(values.Encode())

	case "text/plain":
		result = []byte("default text plain content")
			
	default: 
		result = []byte("default body")
	}

	return result
}


func extractValueRecursive(requestBody interface{})(map[string]interface{}){
	array := make(map[string]interface{})

	if requestBody != nil {
		for _, reqBody := range requestBody.([]interface{}) {
			reqBody := reqBody.(RequestParam)
			if reqBody.Type == "array"{
				if checkRequestParamInterface(reqBody.Value) == true{
					value := extractValueRecursive(reqBody.Value)
					arrayTemp := [...]interface{}{value,value}
					array[reqBody.Name] = arrayTemp
				} else {
					value:= GenerateValue(reqBody.Name,"string",reqBody.Format,nil)
					arrayTemp := [...]interface{}{value,value}
					array[reqBody.Name] = arrayTemp
				}
				
			} else if reqBody.Type == "Appitems"{
					value := extractValueRecursive(reqBody.Value)
					array[reqBody.Name] = value
			} else if reqBody.Type == "ApparrayItems"{
					value := extractValueRecursive(reqBody.Value)
					array["ApparrayItems"] = value
			} else if reqBody.Type!= "Apparray" {
				value:= GenerateValue(reqBody.Name,reqBody.Type,reqBody.Format,nil)
				array[reqBody.Name] = value
			} else {
				value := extractValueRecursive(reqBody.Value)
				arrayTemp := [...]interface{}{value}
				array[reqBody.Name] = arrayTemp
			}
			
		}
	}

	return array
}

func checkRequestParamInterface(requestBody interface{})(bool){
	if requestBody != nil {
		 for _, reqBody := range requestBody.([]interface{}){
		 	if reflect.TypeOf(reqBody).String() == "cmd.RequestParam" {
		 		return true
		 	} else {
		 		return false
		 	}
		 }
		 return false
	} else {
		return false
	}
}
