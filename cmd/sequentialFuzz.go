package cmd

import (
	"fmt"
	"strings"
	"strconv"
	"reflect"
	"math"
	"io/ioutil"
	"regexp"
	"github.com/spf13/cobra"
)

var ResponseTable []ResponseInfo

var DiffStatus int
var DiffContent int

const MAX_DiffStatus = 3
const MAX_DiffContent = 3
const MAX_LenDifferent = 5

var FirstStatusCode int
var FirstBody string


var modeIDORFuzz = &cobra.Command{
	Use:   "idor",
	Short: "Testing for Insecure direct object references(IDOR)",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		readYamlFile(ConfigFile)
		printTitle("--------- LIST ALL IDOR ---------")
		path := fetchSwaggerURL()
		processSwaggerFile(path,"idor")
	},
}


func init() {
}



func sequentialFuzz(BaseURL string,Path string, Parameter []interface{}, requestBody []interface{}, ContentType string, method string){
	wordlists := [...]string{"id","uuid","key","pk"}

	FirstResponse, err := GenerateRequests(BaseURL,Path,Parameter,requestBody,ContentType,method,true)
	
	if err != nil {
		printError(fmt.Sprintf("%s",err))
	} else {
		defer FirstResponse.Body.Close()
		FirstStatusCode = FirstResponse.StatusCode
		FirstBodyByte, _ := ioutil.ReadAll(FirstResponse.Body)
	    FirstBody = string(FirstBodyByte)

	}

	for _,param := range FullRequestParam{
		for _, keyword := range wordlists {
			new_param := strings.ToLower(param)
			if strings.Contains(new_param,keyword){
				DiffStatus = 0
				DiffContent = 0
				BruteforcePayload(BaseURL,Path,Parameter,requestBody,ContentType,method,param)
			}
		}
	}
}

func generateDefaultPayload()([]string){
	var payload []string
	for i := 0; i < 100; i++ {
		payload = append(payload,strconv.Itoa(i))
	}
	return payload
}

func generateBruteforcePayload(keyword string)([]string){
	// format 1: range a-b
	// format 2: has , in 
	var payload []string

	if strings.Contains(keyword,"-"){
		parts := strings.SplitN(keyword, "-", 2)
		if len(parts) != 2 || reflect.TypeOf(parts[0]).Kind() == reflect.Int || reflect.TypeOf(parts[1]).Kind() == reflect.Int{
	    	return generateDefaultPayload() 
	   } else {
	   		from, err:= strconv.Atoi(parts[0])
	   		if err!= nil {
	   			return generateDefaultPayload()
	   		}
	   		to, err:= strconv.Atoi(parts[1])
	   		if err!= nil {
	   			return generateDefaultPayload()
	   		}
		   	for i := from; i <= to; i++ {
				payload = append(payload,strconv.Itoa(i))
			}
	   }
	}else if strings.Contains(keyword,","){
		parts := strings.Split(keyword, ",")
		for _,part := range parts {
			payload = append(payload,part)
		}
	} else {
		payload = append(payload,keyword)
	}
	return payload
}


func BruteforcePayload(BaseURL string,Path string, Parameter []interface{}, requestBody []interface{}, ContentType string, method string,keyword string){
	// only number id
	var payloads []string
	var hasValue bool
	var oldValue string 
	var SecondStatusCode int 
	var SecondBody string

	if config.Parameters[keyword] != "" {
		oldValue = config.Parameters[keyword]
		payloads = generateBruteforcePayload(oldValue)
		hasValue = true

	} else {
		payloads =  generateDefaultPayload()
		hasValue = false
	}

	for _, payload := range payloads{
		config.Parameters[keyword] = payload
		SecondResponse, err := GenerateRequests(BaseURL,Path,Parameter,requestBody,ContentType,method,true)

		if err != nil {
			printError(fmt.Sprintf("%s",err))
		} else {
			defer SecondResponse.Body.Close()
			SecondStatusCode = SecondResponse.StatusCode
			SecondBodyByte, _ := ioutil.ReadAll(SecondResponse.Body)
	    	SecondBody = string(SecondBodyByte)
		}

		if checkAuthorzationStatusCode(FirstStatusCode,SecondStatusCode) && checkAuthorzationContent(FirstBody,SecondBody){
			message := "[+] Found IDOR in Param: " + keyword
			writeLog(SecondStatusCode,BaseURL+Path, method,message)
			break
		} else if checkAuthorzationStatusCode(FirstStatusCode,SecondStatusCode) || checkAuthorzationContent(FirstBody,SecondBody){
			message := "[+] May be IDOR in Param: " + keyword
			writeLog(SecondStatusCode,BaseURL+Path, method,message)
			break
		}
	}

	if hasValue {
		config.Parameters[keyword] = oldValue

	} else {
		config.Parameters[keyword] = ""
	}
}




func checkAuthorzationStatusCode(FirsttatusCode  int,SecondStatusCode int)(bool) {
	if DiffStatus > MAX_DiffStatus {
		return true
	} else {
		if FirstStatusCode != SecondStatusCode{
			if SecondStatusCode != 429 && SecondStatusCode != 403 && SecondStatusCode != 401{
				DiffStatus = DiffStatus + 1
			}
		}
	}
	return false
}

func checkAuthorzationContent(FirstBody string,SecondBody string)(bool) {
	// compare only when they has same status code
	if DiffContent > MAX_DiffContent {
		return true
	} else {
		pattern := "(?i)" + strings.Join(config.UnauthorizedString[:], "|")
		re, err := regexp.Compile(pattern)
		if err != nil {
			printError("Error compiling regex pattern:" + fmt.Sprintf("%s",err))
			panic(err)
		}

		if re.MatchString(SecondBody) == true && re.MatchString(FirstBody) == false {
			return false
		}

		pattern = "(?i)not found"
		re, err = regexp.Compile(pattern)
		if err != nil {
			printError("Error compiling regex pattern:" + fmt.Sprintf("%s",err))
			panic(err)
		}

		if re.MatchString(SecondBody) == true && re.MatchString(FirstBody) == false {
			DiffContent = DiffContent + 1
			return false
		}

		if math.Abs(float64(len(FirstBody) - len(SecondBody))) > MAX_LenDifferent {
			DiffContent = DiffContent + 1
		}
	}
	return false
}
