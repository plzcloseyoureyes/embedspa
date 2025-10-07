
# embedspa
##### Embed Single Page Applications like React with Go Embed, Features: Auto Index, ETag support, Strip prefix for specific paths.
##### Features:
* Returns index.html contents for unexist routes
* ETag support for caching
* Strip prefix for specific paths.

# Gin usage example:
```go
package main
import (
    "embed"
    "io/fs"
    "github.com/l10r/embedspa"
    "github.com/gin-gonic/gin"
)

//go:embed dist
var reactAppEmbed embed.FS

func main() {
    r  := gin.Default()
    embedFS, _  := fs.Sub(reactAppEmbed, "dist")
    spaExample  := embedspa.NewEmbedSPAHandler(embedFS).
        StripPrefixURL("").
        SetIndexPath("index.html")
    r.GET("/*any", gin.WrapF(spaExample.ServeHTTP))
    r.Run()
}
``` 
