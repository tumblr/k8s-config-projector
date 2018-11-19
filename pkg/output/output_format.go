package output

import (
	"fmt"
	"regexp"
)

// BuildFileOutputName - Creates the output file name in the form of {namespace}--{name}--{timestamp}.yaml
func BuildFileOutputName(namespace string, name string, timestamp int64) string {
	return fmt.Sprintf("%s--%s--%d.yaml", namespace, name, timestamp)
}

// FormatName - Format name to k8s compatible string to only include alpha numeric and -.
func FormatName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9-.]`)
	return re.ReplaceAllString(name, `$1-$2`)
}
