package cmd



type RequestParam struct {
	In string
	Name string
	Type string
	Format string
	Value interface{}
}

type Config struct {
	ID         string            `yaml:"id"`
	Info       Info              `yaml:"info"`
	Profiles   []Profile         `yaml:"profiles"`
	Parameters map[string]string `yaml:"parameters"`
	Decentralization      []DecInfo `yaml:"decentralization"`
	UnauthorizedString    []string  `yaml:"unauthorized_response"`
}

type Profile struct {
	Name   string            `yaml:"name"`
	Header map[string]string `yaml:"header"`
	Cookie map[string]string `yaml:"cookie"`
	Query  map[string]string `yaml:"query"`
	Data   map[string]string `yaml:"data"`
}

type Info struct {
	Name        string `yaml:"name"`
	Author      string `yaml:"author"`
	Description string `yaml:"description"`
}

type DecInfo struct {
	Name  string   `yaml:"name"`
	Paths []Path  `yaml:"paths"`
}

type Path struct {
	Path string `yaml:"path"`
}


type ResponseInfo struct {
	BaseURL       string 
	Method 		  string
	ContentLength int    
	StatusCode    int    
	ContentType   string 
}

