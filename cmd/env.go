package cmd


var UserAgent string
var swaggerURL string
var EnvHeaders []string
var apiTarget string
var timeout int64
var Proxy string
var EnvCookies []string
var SpecificPath string
var ConfigFile string
var TargetURL []string


var OutputName string
var OutputFormat string
var config Config
var CurrentProfile string
var ConnectionErrCount = 0
const MAX_REQUEST_ERROR = 5
