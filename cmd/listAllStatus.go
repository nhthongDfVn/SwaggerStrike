package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"encoding/json"
	"os"
)

var ResultTable []ResponseInfo

var modeListAllStatus = &cobra.Command{
	Use:   "listall",
	Short: "Send all requests and see the response status",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		readYamlFile(ConfigFile)
		printTitle("--------- LIST ALL STATUS ---------")
		path := fetchSwaggerURL()
		processSwaggerFile(path,"listall")
	},
}


func init() {
}


func listAllStatus(BaseURL string,Path string, Parameter []interface{}, requestBody []interface{}, ContentType string, method string){

	var StatusCode int 
	//var Body string

	Response, err := GenerateRequests(BaseURL,Path,Parameter,requestBody,ContentType,method,true)

	if err != nil {
		printError(fmt.Sprintf("%s",err))
	} else {
		defer Response.Body.Close()
		StatusCode = Response.StatusCode
		BodyByte, _ := ioutil.ReadAll(Response.Body)
		Body := string(BodyByte)
		ContentLength := len(Body)

		info := updateResponseInfo(BaseURL+Path,method,ContentLength,StatusCode,ContentType)
		ResultTable = append(ResultTable,info)
		writeLog(StatusCode,BaseURL+Path, method,"")
	}
}

func listAllStatusOutput(){
	if OutputFormat == "json"{
		json, err := json.Marshal(ResultTable)
		  if err != nil {
		}
		os.WriteFile(OutputName, json, 0666)
	} else if OutputFormat == "txt" {

		isExist,_ := checkIfFileExists(OutputName)
		if isExist == true{
			os.Remove(OutputName)
		}

		file, err := os.Create(OutputName)
		if err != nil {
			printError("Error creating file:" + fmt.Sprintf("%s",err))
			return
		}

		for _, line := range ResultTable {
			newLine := "Method: %-10s\t Status: %-5d\t Content-length: %-10d\t BaseURL: %s\n"
			_, err := fmt.Fprintf(file, newLine,line.Method,line.StatusCode,line.ContentLength,line.BaseURL)
			if err != nil {
				printError("Error writing to file:" + fmt.Sprintf("%s",err))
				return
			}
		}
	}
}





