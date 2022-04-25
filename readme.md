
# JRPC [![MIT License](https://img.shields.io/apm/l/atomic-design-ui.svg?)](https://github.com/tterb/atomic-design-ui/blob/master/LICENSEs)


JSON RPC 2.0 implementation using websockets. Register a struct with a functions with the server and the jrpc communication will be automatically handled. Create a subscription type of function and it will send regular updates to the client.




## Example

One way would be to clone this repo and run: ```go run example/*```


Or just use Server example file:

```go
package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kroksys/jrpc"
	"github.com/kroksys/jrpc/registry"
)

func main() {
	jrpcServer := jrpc.NewServer()
	if err := jrpcServer.Register("example", Example{}); err != nil {
		log.Panicln(err)
	}

	r := gin.Default()
	r.GET("/ws", jrpcServer.WebsocketHandlerGin)

	log.Printf("JSON RPC 2.0 server started. Address: localhost:3333/ws\n")
	err := r.Run("localhost:3333")
	if err != nil {
		log.Println("jrpc server stopped")
	}
}

type Example struct{}

func (Example) Simple(x, y int) (int, error) {
	return x + y, nil
}

func (Example) SimpleNoParams() (int, error) {
	return 5, errors.New("simple error")
}

func (Example) SimpleWithContext(ctx context.Context, x, y int) (int, error) {
	return x + y, nil
}

func (Example) Subscription(sub *registry.Subscription) error {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		select {
		case <-sub.Unsubscribe:
			return nil
		default:
			if !sub.Notify("Hello") {
				return nil
			}
			time.Sleep(time.Second)
		}
	}
	return nil
}
```

When the server is running connect to the ```ws://localhost:3333/ws```server (i.e. using postman) and send json data bellow to trigger methods and get response.
```json
{"jsonrpc":"2.0","method":"example.Simple", "params": [2, 3], "id":2865}
// Response: {"jsonrpc":"2.0","result":5,"id":2865}

{"jsonrpc":"2.0","method":"example.SimpleNoParams","id":2866}
// Error response because function returns error:
// {"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error","data":"simple error"},"id":2866}

{"jsonrpc":"2.0","method":"example.SimpleWithContext", "params": [2, 3], "id":2866}
// Response: {"jsonrpc":"2.0","result":5,"id":2866}

// Subscribe to regular updates usgin request
{"jsonrpc":"2.0","method":"example.subscribe.Subscription","id":2868}
{"jsonrpc":"2.0","method":"example.unsubscribe.Subscription","id":2869}

// Or using notifications
{"jsonrpc":"2.0","method":"example.subscribe.Subscription"}
{"jsonrpc":"2.0","method":"example.unsubscribe.Subscription"}

```
## Authors

- [@kroksys](https://www.github.com/kroksys)

