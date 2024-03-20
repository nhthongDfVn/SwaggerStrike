package cmd 

import (
	"strings"
	"regexp"
	"io/ioutil"
	"github.com/spf13/cobra"
	"fmt"
)

var modeListPrivilege = &cobra.Command{
	Use:   "privilege",
	Short: "Testing for Broken Access Control(privilege)",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		checkValid()
		readYamlFile(ConfigFile)
		printTitle("--------- LIST ALL PRIVILEGED ENDPOINT ---------")
		path := fetchSwaggerURL()
		processSwaggerFile(path,"privilege")
	},
}


func init() {
}


func authorizationFuzz(BaseURL string,Path string, Parameter []interface{}, requestBody []interface{}, ContentType string, method string){

	var StatusCode int 
	var Body string
	if len(config.Profiles) < 2 {
		printError("CAN NOT EXECUTE: You need 2 role to run this mode. ")
		return
	}


	for _, profile := range config.Profiles {
		CurrentProfile = profile.Name

		Response, err := GenerateRequests(BaseURL,Path,Parameter,requestBody,ContentType,method,true)

		if err != nil {
			printError(fmt.Sprintf("%s",err))
		} else {
			defer Response.Body.Close()
			StatusCode = Response.StatusCode
			BodyByte, _ := ioutil.ReadAll(Response.Body)
		    Body = string(BodyByte)

		}

		if checkauthorizationCondition(StatusCode,Body) == false &&  StringInArray(profile.Name,GetProfileValidWithPath(Path)) == false {
			message := "Path: " + Path + " Profile: " + profile.Name
			writeLog(StatusCode,BaseURL+Path, method,message)
		}
	}
	
	CurrentProfile = ""
}


func GetProfileValidWithPath(UriPath string)([]string){

	var profileList []string
	var pattern string


	for _, path := range config.Decentralization {
		for _, pathName := range path.Paths{

			if strings.Contains(pathName.Path,"*"){
				pattern = "(?i)" + strings.ReplaceAll(pathName.Path, "/", "\\/")
				pattern = strings.ReplaceAll(pattern, "*", ".*")
			} else {
				pattern = "(?i)" + strings.ReplaceAll(pathName.Path, "/", "\\/")
			}

			re, err := regexp.Compile(pattern)
			if err != nil {
				printError("Error compiling regex pattern:" + fmt.Sprintf("%s",err))
				panic(err)
			}

			if re.MatchString(UriPath) == true{
				profileList = append(profileList,path.Name)
			}
		}
	}

	return profileList
}


func checkauthorizationCondition(StatusCode  int,Body string)(bool) {
	if StatusCode == 403 || StatusCode == 401 || StatusCode == 301 || StatusCode == 302  {
		return true
	}

	pattern := "(?i)" + strings.Join(config.UnauthorizedString[:], "|")
	re, err := regexp.Compile(pattern)
	if err != nil {
		printError("Error compiling regex pattern:" + fmt.Sprintf("%s",err))
		panic(err)
	}

	if re.MatchString(Body) == true {
		return true
	}

	return false
}
