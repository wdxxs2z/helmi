package constants

const (
	LookupRegex = `\{\{\s*lookup\s*\(\s*'(?P<type>[\w]+)'\s*,\s*'(?P<path>[\w/:]+)'\s*\)\s*\}\}`
	LookupRegexType = "type"
	LookupRegexPath = "path"
	LookupValue    = "value"
	LookupCluster  = "cluster"
	LookupUsername = "username"
	LookupPassword = "password"
	LookupEnv      = "env"
	LookupRelease  = "release"
)

const (
	FetchServiceCatalog = "catalog"
	InstanceIDLogKey = "instance-id"
	BindingIDLogKey = "binding-id"
	DetailsLogKey = "details"
	AcceptsIncompleteLogKey = "acceptsIncomplete"
)