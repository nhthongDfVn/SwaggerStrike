
package cmd


import (
	"io/ioutil"
	"gopkg.in/yaml.v3"
	"fmt"
	"log"
	"strings"
	"reflect"
	"strconv"
	"os"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"encoding/json"
	"net/http"
	"net/url"
	"github.com/fatih/color"
	"time"
)

func IsVersionSupport(version string) (bool){
	if version != ""{
		return true
	} else {
		return false
	}
}

func checkIfFileExists(filename string) (bool, error) {
    // Check if file exists
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        return false, err
    }

    // Read file contents
    _, err := ioutil.ReadFile(filename)
    if err != nil {
        return false, err
    }

    return true, nil
}



func checkIsOpenAPIVersion(filePath string) (bool) {
    content, _ := ioutil.ReadFile(filePath)
    fileContent := string(content)

    if strings.Contains(fileContent, "\"swagger\":") || strings.Contains(fileContent, "swagger:") {
        return false
    } else if strings.Contains(fileContent, "\"openapi\":") || strings.Contains(fileContent, "openapi:") {
        return true
    } else {
    	printError("Current version not support")
    	os.Exit(0)
    	return false
    }
}



func GetValueFromConfig(dataName string,value interface{})(interface{}){
	valConfig := config.Parameters[dataName]
	if valConfig != ""{
		if strings.Contains(valConfig,"-"){
			parts := strings.SplitN(valConfig, "-", 2)
			if len(parts) != 2 || reflect.TypeOf(parts[0]).Kind() == reflect.Int || reflect.TypeOf(parts[1]).Kind() == reflect.Int{
		    	return value
		   } else {
		   		from, err:= strconv.Atoi(parts[0])
		   		if err!= nil {
		   			return value
		   		}
		   		return from
		   }
		}else if strings.Contains(valConfig,","){
			parts := strings.Split(valConfig, ",")
			return parts[0]

		} else {
			return valConfig
		}
	} else {
		return value
	}
}

func GenerateValue(dataName string,dataType string, format string, example interface{}) interface{} {
	FullRequestParam = append(FullRequestParam,dataName)

    switch dataType {
    case "number":
        switch format {
        case "float":
            return  GetValueFromConfig(dataName,float32(3.14))
        case "double":
            return GetValueFromConfig(dataName,float64(3.14159265359))
        default:
            return GetValueFromConfig(dataName,3)
        }
    case "integer":
        switch format {
        case "int32":
            return GetValueFromConfig(dataName,int32(4))
        case "int64":
            return GetValueFromConfig(dataName,int64(5))
        default:
            return GetValueFromConfig(dataName,6)
        }
    case "string":
        switch format {
        case "date":
            return GetValueFromConfig(dataName,"2017-07-21")
        case "date-time":
            return GetValueFromConfig(dataName,"2017-07-21T17:32:28Z")
        case "password":
            return GetValueFromConfig(dataName,"want password")
        case "byte":
            return GetValueFromConfig(dataName,"U3dhZ2dlciByb2Nrcw==")
        case "binary":
            return GetValueFromConfig(dataName,"need file content")
        default:
            return GetValueFromConfig(dataName,"1")
        }
    case "boolean":
    	return GetValueFromConfig(dataName,true)
    default:
        return "unsupported data type: " + dataType
    }
}

func readYamlFile(filename string) {
    yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}
}



func parseSwagger(filename string)(openapi2.T){
	input, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var doc openapi2.T
	if err = json.Unmarshal(input, &doc); err != nil {
		printError("Error when read file: " + filename)
	}

	return doc

}

func StringInArray(keyword string, list []string) bool {
    for _, item := range list {
        if item == keyword {
            return true
        }
    }
    return false
}


func ConvertToString(value interface{})(string){
	return fmt.Sprintf("%v", value)
}


func updateResponseInfo(BaseURL string, Method string,ContentLength int,StatusCode int,ContentType string)ResponseInfo{
	var info ResponseInfo

	info.BaseURL = BaseURL
	info.Method = Method
	info.ContentLength = ContentLength
	info.StatusCode = StatusCode
	info.ContentType = ContentType
	return info
}


func fetchSwaggerURL()string{
	var filename string
	var openAPIString string
	
	filename = ""
	if strings.HasPrefix(swaggerURL,"http"){
			checkValid()
			if strings.HasSuffix(swaggerURL,".json"){
			
			openAPIString = getContentBody(swaggerURL)
			} else {
				htmlContent := getContentBody(swaggerURL)

				_, err := json.Marshal(htmlContent)

				 // check if html content is json
			    if err == nil {
			     openAPIString = htmlContent
			    } else{
			     openAPIString = extractSpecFromHtml(htmlContent)
			    }
			}
			filename = "openapi_data.json"
			os.WriteFile(filename, []byte(openAPIString), 0666)
			return filename
	} else {
		isExist,err :=checkIfFileExists(swaggerURL)
		if err!= nil {
			return ""
		}

		if isExist{
			return swaggerURL
		}
	}

	return filename
}

func getContentBody(url string)string{
	var req *http.Request
	var err error


	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}

	for _,header := range EnvHeaders{

		values := strings.Split(header,":")

		req.Header.Add(values[0], values[1])
	}
	
	for _, cookie := range EnvCookies {

			values := strings.Split(cookie,":")
			cookie := &http.Cookie{
					Name:  values[0],
					Value: values[1],
			}

			req.AddCookie(cookie)
		}

	// Send the request
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return ""
	}

	statusCode := response.StatusCode

	if statusCode != 200 {
		fmt.Println("Invalid status code: ", statusCode)
		return ""
	}

	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return ""
	}
	return string(body)
}



func extractSpecFromHtml(content string )string{
	contentArray := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")

	for _, line := range contentArray{
		if strings.Contains(line,"var spec = ") {
			startIndex := strings.Index(line, "{")
			endIndex := strings.LastIndex(line, "}")

			if startIndex == -1 || endIndex == -1 || endIndex <= startIndex {
				return ""
			}

			substring := line[startIndex : endIndex+1]
			return substring
		}
	}
	printError("Can not fetch Swagger/OpenAPI document")
    os.Exit(0)
	return ""

}

func getHostFromURI(uri string)string{
	parsedURI, err := url.Parse(uri)
    if err != nil {
        fmt.Println("Error parsing URI:", err)
        return ""
    }

    return parsedURI.Scheme + "//" + parsedURI.Host
}



func getServerList(doc *openapi3.Servers){

	if apiTarget != ""{
		TargetURL = append(TargetURL,apiTarget)
	} else {
		for _,server := range *doc{
			TargetURL = append(TargetURL,server.URL)
		}
	}

	if len(TargetURL) < 1{
		printTitle("No target uri specific, using default swagger uri")
		TargetURL = append(TargetURL,getHostFromURI(swaggerURL))
	}
}

func isValidURL(input string) error{
    _, err := url.ParseRequestURI(input)
    if err != nil {
    	fmt.Println(err)
        return err
    }


    u, err := url.Parse(input)
    if err != nil || u.Scheme == "" || u.Host == "" {
    	fmt.Println(err)
        return err
    }

    return nil
}

func checkValid(){
	    err := isValidURL(swaggerURL)
	    if err != nil {
	    	printTitle("url is not valid. ")
	    	os.Exit(1)
	    }
}


func printBanner() {
	 text := `
 __                                     __ _        _ _        
/ _\_      ____ _  __ _  __ _  ___ _ __/ _\ |_ _ __(_) | _____ 
\ \\ \ /\ / / _  |/ _  |/ _  |/ _ \ '__\ \| __| '__| | |/ / _ \
_\ \\ V  V / (_| | (_| | (_| |  __/ |  _\ \ |_| |  | |   <  __/
\__/ \_/\_/ \__,_|\__, |\__, |\___|_|  \__/\__|_|  |_|_|\_\___|
                  |___/ |___/                                  
`

    lines := strings.Split(text, "\n")
    //maxLength := 0

    // Print the beautified ASCII text with color
    for _, line := range lines {
        color.Red(line) // Example: You can change color.Cyan to any other color function from the color package
    }
    color.Red("                                                 @nhthongdfvn") 
    color.Red("                                                 Version: 1.0.0") 
}

func printTitle(text string) {
    color.Green(text)  
}

