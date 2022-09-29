# styra-run-sdk-go

The Styra Run `golang` SDK.

## How to install

First, add the SDK as a dependency:

`go get github.com/StyraInc/styra-run-sdk-go`

## Initialize the client

The client wraps the core Styra Run API. You can initialize it as follows:

```golang
package main

import (
    api "github.com/styrainc/styra-run-sdk-go/api/v1"
)

func main() {
    client := api.New(
        &api.Settings{
            Token:             "",
            Url:               "",
            DiscoveryStrategy: api.Simple,
            MaxRetries:        3,
        },
    )
}
```

### Token

In Styra Run, you interact with projects created through the UI or API. Projects have a list of environments. Here, `Token` refers to an environment-specific token within a project.

### Url

The client is bound to a specific environment within a project. The url should be of the following format:

`https://api-test.styra.com/v1/projects/{user}/{project}/envs/{env}`

Here, `user`, `project`, and `env` refer to a specific Styra Run user, project and environment, respectively.

### Discovery

When initialized, the client enumerates all data plane urls that can be used to call into Styra Run. The `DiscoveryStrategy` parameter controls how the client chooses a specific url. Here's a table describing the various strategies. Note that more clever strategies may be added in the future.

| Strategy | Description |
| --- | --- |
| `Simple` | Given a list of urls, start with the first one in the list. In the event of a failure, try the next available url in the list in a round robin fashion. |


`MaxRetries` controls how many times the client will retry in the event of certain errors.

## Use the client

Once the client has been initialized, you can use it to interact with Styra Run. The following sections describe the available operations.

### GetData

```golang
path := "rbac/user_bindings/acmecorp"

var result interface{}

err := client.GetData(ctx, path, &result)
```

The client will automatically serialize and deserialize data to and from Styra Run and `json` tags are fully supported. This is true of all SDK operations. Make sure to pass the `data` parameter by reference if it's not a reference type. 

### PutData

```golang
err := client.PutData(ctx, path,
    map[string]interface{}{
        "alice": []string{"ADMIN"},
        "bob":   []string{"VIEWER"},
    },
}
```

### DeleteData

```golang
err := client.DeleteData(ctx, path)
```

### Query

This executes a policy rule query within Styra Run and emits the response. Here, `input` is arbitrary structured data that's used as input to the policy rule.

```golang
query := "tickets/resolve/allow"
input := map[string]string{
    "tenant":  "acmecorp",
    "subject": "alice",
}

var result interface{}

err := client.Query(ctx, query, input, &result)
```

### Check

The same as `Query`, but returns `true` if the Styra Run response is `{"result": true}` and `false` otherwise.

```golang
ok, err := client.Check(ctx, query, input)
```

### BatchQuery

Allows you to execute multiple queries at once. Note that the client will seamlessly issue multiple requests to Styra Run if the batch size exceeds the Styra Run API limit. Results and potential errors are set by reference within each `Query` instance, and the order of the queries is preserved. You can pass in a global input data structure that's used as a fallback if each query doesn't set it's input field.

```golang
queries := make([]api.Query, 0)

for i := 0; i < 64; i++ {
    queries = append(
        queries,
        api.Query{
            Path:  query,
            Input: input,
        },
    )
}

err := client.BatchQuery(ctx, queries, nil)
```

## Initialize RBAC

The RBAC management API wraps the default RBAC policies within Styra Run. RBAC stands for role-based access control. To use RBAC you will first need to initialize it:

```golang
import (
    rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
)

func main() {
    // Initialize the client ..

    myRbac := rbac.New(
        &rbac.Settings{
            Client: client,
        },
    )
}
```

## Use RBAC

Once RBAC is initialized, you can use it to perform the following operations:

### GetRoles

This allows you to retrieve all available user roles as a list of `string`'s.

```golang
session := &api.Session{
    Tenant:  "acmecorp",
    Subject: "alice",
}

result, err := myRbac.GetRoles(ctx, session)
```

Note that all RBAC calls are guarded with an authorization check. This is to ensure that the person performing the request has the appropriate permissions to make said request.

### ListUserBindingsAll

This emits _all_ user bindings.

```golang
result, err := myRbac.ListUserBindingsAll(ctx, session)
```

### ListUserBindings

This emits user bindings for the specified users.

```golang
users := []*rbac.User{
    {Id: "alice"},
    {Id: "bob"},
    {Id: "bryan"},
    {Id: "cesar"},
}

result, err := myRbac.ListUserBindings(ctx, session, users)
```

### GetUserBinding

```golang
bruce := &rbac.User{
    Id: "bruce",
}

result, err := myRbac.GetUserBinding(ctx, session, bruce)
```

### PutUserBinding

```golang
binding := &rbac.UserBinding{
    Roles: []string{"OWNER"},
}

err := myRbac.PutUserBinding(ctx, session, bruce, binding)
```

### DeleteUserBinding

```golang
err := myRbac.DeleteUserBinding(ctx, session, bruce)
```

## Proxies

To make it easier for the programmer to serve the SDK in their own web servers we provide proxy implementations of some client and all RBAC functions. The proxies generate routes that look like the following:

```golang
type Route struct {
    Path    string
    Method  string
    Handler http.HandlerFunc
}
```

Here, `Path` is the suffix used for a route. For example, for the `GetRoles` RBAC function, `Path` is `/roles` and `Method` is `GET`. You can use this information to install the handlers in your own web server.

The client proxy exposes the `BatchQuery` function which is used by the [frontend SDK](https://github.com/StyraInc/styra-run-sdk-js) to make UI rendering decisions. The rbac proxy exposes all RBAC functions.

## Initialize the client proxy

```golang
package main

import (
    api "github.com/styrainc/styra-run-sdk-go/api/v1"
    "github.com/styrainc/styra-run-sdk-go/api/v1/proxy"
)

func main() {
    // Initialize the client ..
    
    myProxy := proxy.New(
        &proxy.Settings{
            Client:    client,
            Callbacks: proxy.DefaultCallbacks(
                &proxy.DefaultCallbackSettings{
                    GetSession: api.SessionFromValues("acmecorp", "alice"),
                },
            ),
        },
    )
}
```

Here, `Callbacks` is a struct containing the following functions:

| Callback | Description | Required |
| --- | --- | --- |
| `GetSession` | Extracts `session` information from the http request. There are several implementations provided that pull `session` information from HTTP cookies, the `context`, and so on. You can also implement your own. | yes |
| `OnModifyBatchQueryInput` | Allows the programmer to inject values into each query input field and the global input field. The default implementation automatically injects the tenant and subject. | no |

## Use the client proxy

The proxy emits routes that have the following shape:

### BatchQuery

```
POST /
{
    "items": [
        {
            "path": "tickets/resolve/allow"
        },
        {
            "path": "tickets/resolve/allow",
            "input": {
                "tenant": "acmecorp",
                "subject": "alice"
            }
        }
    ],
    "input": {
        "tenant": "acmecorp",
        "subject": "alice"
    }
}

->

{
    "result": [
        {
            "result": true
        },
        {
            "result": true
        }
    ]
}
```

## Initialize the RBAC proxy

```golang
package main

import (
    "github.com/gorilla/mux"

    api "github.com/styrainc/styra-run-sdk-go/api/v1"
    rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
    "github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy"
)

var (
    users = []*rbac.User{
        {Id: "alice"},
        {Id: "bob"},
        {Id: "bryan"},
        {Id: "cesar"},
        {Id: "emily"},
    }
)

func main() {
    // Initialize the client ..

    myProxy := proxy.New(
        &proxy.Settings{
            Client:    client,
            GetUrlVar: func(r *http.Request, key string) string {
                return mux.Vars(r)[key]
            },
            Callbacks: proxy.ArrayCallbacks(
                &proxy.ArrayCallbackSettings{
                    GetSession: api.SessionFromValues("acmecorp", "alice"),
                    Users:      users,
                    PageSize:   2,
                },
            ),
        },
    )
}
```

Here, `GetUrlVar` tells the proxy how to extract url parameters. For example, if the route's `Path` is `/user_bindings/{id}`, this function should emit the part of the url that corresponds to the string `"id"`. How this value is extracted depends on the web server library you are using. The example above assumes the [gorilla/mux](https://github.com/gorilla/mux) library.

`Callbacks` is a struct containing the following functions:

| Callback | Description | Required |
| --- | --- | --- |
| `GetSession` | This is the same as the client proxy. | yes |
| `GetUsers` | This callback is used by the `ListUserBindings` proxy. It allows the proxy to page users (and hence user bindings) in and out. If omitted, all user bindings are emitted and pagination is ignored. See below for more details. | no |
| `OnGetUserBinding` | Control whether `GetUserBinding` is allowed. | no |
| `OnPutUserBinding` | Control whether `PutUserBinding` is allowed. | no |
| `OnDeleteUserBinding` | Control whether `DeleteUserBinding` is allowed. | no |

## Use the RBAC proxy

### GetRoles 

 ``` 
GET /roles 

-> 

{ 
    "result": [ 
        "ADMIN", 
        "VIEWER" 
    ] 
} 
``` 

### ListUserBindings 

``` 
GET /user_bindings?page=3 

-> 

{ 
    "result": [ 
        { 
            "id": "alice", 
            "roles": [ 
                "ADMIN" 
            ] 
        }, 
        { 
            "id": "bob", 
            "roles": [ 
                "VIEWER" 
            ] 
        }, 
        { 
            "id": "bryan", 
            "roles": [] 
        } 
    ], 
    "page": { 
        "index": 0, 
        "total": 4 
    } 
} 
``` 

Here, the input bytes to `GetUsers` is the string `"3"` from the `page=3` query parameter. The output should contain a list of users for that page and an `interface{}` that serves as the value for the `"page"` key in the response. Notice that how the users are paged is ultimately up to the programmer. It's important to note, however, that the [frontend SDK](https://github.com/StyraInc/styra-run-sdk-js) assumes the above structure so you must follow suit if you want to use it. See `rbac/v1/proxy/callbacks.go` for a concrete example that stores users in an array.

If the programmer does not provide a `GetUsers` callback, the proxy will internally call `ListUserBindingsAll`, emitting _all_ user bindings and ignoring pagination.

### GetUserBinding

```
GET /user_bindings/{id}

->

{
    "result": [
        "ADMIN"
    ]
}
```

### PutUserBinding

```
PUT /user_bindings/{id}
[
    "VIEWER"
]

->

{}
```

### DeleteUserBinding

```
DELETE /user_bindings/{id}

->

{}
```

## Examples

If you do not wish to write your own web server, we provide the following concrete implementations which you can find in `examples/v1/proxies`:

* [gorilla/mux](https://github.com/gorilla/mux)
