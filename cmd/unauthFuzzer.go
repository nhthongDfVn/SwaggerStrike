package cmd

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"github.com/spf13/cobra"
	"regexp"
)

var modeListUnAuth = &cobra.Command{
	Use:   "unauth",
	Short: "Testing for Bypassing Authentication Schema",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		checkValid()
		readYamlFile(ConfigFile)
		printTitle("--------- LIST ALL UNAUTH ---------")
		path := fetchSwaggerURL()
		processSwaggerFile(path,"unauth")
	},
}


func init() {
}


func checkUnauthStatusCode(statusCode int)(bool) {
	switch statusCode {
		case 400,404,405,415,429:
			return false
		case 401,403:
			return false
		case 301,302:
			return true
		default:
			return true
	}
}

func checkUnauthResponseBody(response *http.Response) (bool) {

	bodyBytes, err := ioutil.ReadAll(response.Body)
    responseBody := string(bodyBytes)

    pattern := "(?i)" + strings.Join(config.UnauthorizedString[:], "|")
	re, err := regexp.Compile(pattern)

	if err != nil {
		printError("Error compiling regex pattern:" + fmt.Sprintf("%s",err))
		panic(err)
	}

	if re.MatchString(responseBody) == true {
		return false
	}

	return true

}


func UnauthCheckFunc(BaseURL string,Path string, Parameter []interface{}, requestBody []interface{}, ContentType string, method string){
	Response, err := GenerateRequests(BaseURL,Path,Parameter,requestBody,ContentType,method,true)
	
	
	if err != nil {
		printError(fmt.Sprintf("%s",err))
	} else {
		defer Response.Body.Close()
		StatusCode := Response.StatusCode

		if checkUnauthStatusCode(StatusCode) == true{
			if checkUnauthResponseBody(Response) == true {
				writeLog(StatusCode,BaseURL+Path, method,"[+] Detect Bypass Authentication endpoint")
			}
		}
	}
}
