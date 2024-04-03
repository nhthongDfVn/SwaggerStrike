/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"
	"github.com/spf13/cobra"
)


// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
    Use:   "main",
    Short: "A brief description of your application",
    Long: ``,
    Run: func(cmd *cobra.Command, args []string) {
    },
}


func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}


func init() {
	printBanner()

	rootCmd.AddCommand(modeListAllStatus)
	rootCmd.AddCommand(modeListUnAuth)
	rootCmd.AddCommand(modeIDORFuzz)
	rootCmd.AddCommand(modeListPrivilege)
	

	rootCmd.PersistentFlags().StringVarP(&UserAgent, "agent", "a", "SwaggerStrike (bot)", "Config User-Agent string")
	//rootCmd.PersistentFlags().StringVarP(&OutputFormat, "format", "f", "txt", "Declare the format of the output (json/txt).")
	//rootCmd.PersistentFlags().StringVarP(&OutputName, "output", "o", "scan_result", "Declare the name of the output.")
	rootCmd.PersistentFlags().StringArrayVarP(&EnvHeaders, "headers", "H", nil, "Add custom headers, separated by a colon (\"Name:Value\"). Multiple flags are accepted. (Only using for swaggerURL authentication)")
	rootCmd.PersistentFlags().StringArrayVarP(&EnvCookies, "cookies", "C", nil, "Add custom cookies, separated by a colon (\"Name:Value\"). Multiple flags are accepted. (Only using for swaggerURL authentication)")
	rootCmd.PersistentFlags().StringVarP(&Proxy, "proxy", "p", "NOPROXY", "Proxy host and port. Example: http://127.0.0.1:8080")
	rootCmd.PersistentFlags().StringVarP(&SpecificPath, "specificPath", "P", "ALL", "Add more specific paths, separated by ,")
	rootCmd.PersistentFlags().StringVarP(&ConfigFile, "fileConfig", "c", "profile.yaml", "Add more specific paths, separated by ,")
	rootCmd.PersistentFlags().StringVarP(&apiTarget, "target", "T", "", "Manually set a target for the requests to be made if separate from the host the documentation resides on.")
	rootCmd.PersistentFlags().Int64VarP(&timeout, "timeout", "t", 15, "Set the request timeout")
	rootCmd.PersistentFlags().StringVarP(&swaggerURL, "url", "u", "", "Loads the documentation file from a URL (json, yaml format). If openAPI document in HTML format, the tool will try to parse it.")
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	err := rootCmd.MarkPersistentFlagRequired("url")
    if err != nil {
        printTitle("-url flag required. ")
        os.Exit(1)
    }  

    err = rootCmd.MarkPersistentFlagRequired("target")
    if err != nil {
        printTitle("-target flag required. ")
        os.Exit(1)
    }  
}



