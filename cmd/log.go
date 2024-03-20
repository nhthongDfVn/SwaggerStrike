package cmd

import (
	"log"

	colorable "github.com/mattn/go-colorable"
)

var logger = log.New(colorable.NewColorableStdout(), "", 0)

func writeLog(status int, target, method, errorMsg string) {
	switch status {
	case 200:
		logStatusOK(errorMsg,status, target, method)
	case 301, 302:
		logRedirected(errorMsg,status, target, method)
	case 401, 403:
		logForbiden(errorMsg,status, target, method)
	case 404:
		logNotFound(errorMsg,status, target, method)
	case 400:
		logBadRequest(errorMsg,status, target, method)
	case 0:
		logStatusOK(errorMsg,status, target, method)
	default:
		logDefault(errorMsg,status, target, method)
	}
}


func logStatusOK(message string,status int, target, method string) {
	if message == ""{
		message = "Accessible"
	}

	logger.Printf("\x1b[32m%-40s\t Status: %-5d\t Target: %-70s\t Method: %-8s\x1b[0m\n", message,status, target, method)
}

func logNotFound(message string,status int, target, method string) {

	if message == ""{
		message = "Not Found"
	}

	logger.Printf("\x1b[35m%-40s\t Status: %-5d\t Target: %-70s\t Method: %-8s\x1b[0m\n", message,status, target, method)
}

func logRedirected(message string,status int, target, method string) {

	if message == ""{
		message = "Redirected"
	}

	logger.Printf("\x1b[30m%-40s\t Status: %-5d\t Target: %-70s\t Method: %-8s\x1b[0m\n", message,status, target, method)
}

func logForbiden(message string,status int, target, method string) {

	if message == ""{
		message = "Forbiden"
	}

	logger.Printf("\x1b[31m%-40s\t Status: %-5d\t Target: %-70s\t Method: %-8s\x1b[0m\n", message,status, target, method)
}


func logBadRequest(message string,status int, target, method string) {

	if message == ""{
		message = "Bad request"
	}

	logger.Printf("\x1b[36m%-40s\t Status: %-5d\t Target: %-70s\t Method: %-8s\x1b[0m\n", message,status, target, method)
}


func logDefault(message string,status int, target, method string) {

	if message == ""{
		message = ""
	}

	logger.Printf("\x1b[29m%-40s\t Status: %-5d\t Target: %-70s\t Method: %-8s\x1b[0m\n", message,status, target, method)
}



func printError(message string){
	logger.Printf("\x1b[31mERROR: %s\x1b[0m\n", message)
}

func printTask(message string){
	logger.Printf("\x1b[32m %s\x1b[0m\n", message)
}


