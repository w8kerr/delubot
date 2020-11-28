package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

// PrintJSON prints the interface as json to stdout
func PrintJSON(obj interface{}) {
	bytes, _ := json.MarshalIndent(obj, "\t", "\t")
	fmt.Println(string(bytes))
}

// PrintPipe prints a pipe in a format that can be imported into Studio 3T
func PrintPipe(obj interface{}, collection string) {
	bytes, _ := json.MarshalIndent(obj, "", "  ")
	str := string(bytes)

	re := regexp.MustCompile(`("[0-9a-f]{24}")`)
	str = re.ReplaceAllString(str, `ObjectId($1)`)

	re2 := regexp.MustCompile(`("(?:[1-9]\d{3}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1\d|2[0-8])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31)|(?:[1-9]\d(?:0[48]|[2468][048]|[13579][26])|(?:[2468][048]|[13579][26])00)-02-29)T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:Z|[+-][01]\d:[0-5]\d)")`)
	str = re2.ReplaceAllString(str, `ISODate($1)`)

	str = fmt.Sprintf(`db.getCollection("%s").aggregate(%s)`, collection, str)

	fmt.Println(str)
}

// PrintJSONStr Outputs a json representation as a string
func PrintJSONStr(obj interface{}) string {
	bytes, _ := json.MarshalIndent(obj, "\t", "\t")
	return string(bytes)
}

// PrintBody Prints the body of the HTTP request in the given context
func PrintBody(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	fmt.Println(string(body))
}

// JSTTime Print a timestamp in Japanese format
func JSTTime(t time.Time) string {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		fmt.Println(err)
	}
	t = t.In(loc)

	return t.Format("2006年01月02日　15:04:05　JST")
}
