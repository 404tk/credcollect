CredCollect
===========

Automatic credential collection and storage with `CredCollect`.  
Only plaintext passwords (excluding cookies and tokens) are extracted.

**Supported Module**

- Browser
- Navicat
- FileZilla
- WinScp
- Seeyon OA
- Docker Hub

Usage
-----
Command line execution
```shell
credcollect -h
```

CredCollect as a library  
```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/404tk/credcollect/runner"
)

func main() {
	options := &runner.Options{Silent: true}
	res := options.Enumerate()
	r, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(r))
}
```

Contributing
------------

1.  Fork it
2.  Create your feature branch (`git checkout -b my-new-feature`)
3.  Commit your changes (`git commit -am 'Add some feature'`)
4.  Push to the branch (`git push origin my-new-feature`)
5.  Create new Pull Request

License
-------

This repo is released under the [MIT License](http://www.opensource.org/licenses/MIT).