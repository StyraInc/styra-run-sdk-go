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

The client is bound to a specific environment within a project. You can find this url in the environment section of your project's overview page.

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
    queries = append(queries,
        api.Query{
            Path:  query,
            Input: input,
        },
    )
}

err := client.BatchQuery(ctx, queries, input)
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
user := &rbac.User{
    Id: "bruce",
}

result, err := myRbac.GetUserBinding(ctx, session, user)
```

### PutUserBinding

```golang
binding := &rbac.UserBinding{
    Roles: []string{"OWNER"},
}

err := myRbac.PutUserBinding(ctx, session, user, binding)
```

### DeleteUserBinding

```golang
err := myRbac.DeleteUserBinding(ctx, session, user)
```

## Proxies

To make it easier for the programmer to serve the SDK in their own web servers we provide proxy implementations of most client functions and all RBAC functions. Each proxy emits the following:

```golang
type Proxy struct {
    Method  string
    Handler http.HandlerFunc
}
```

Here, `Method` is the HTTP method that's expected and `Handler` is the HTTP handler function that you can embed in your own web server. How you bind said handler to a route is entirely up to you and what HTTP routing library you are using. Each proxy implementation is self contained and has it's own set of requirements.

As an example, in the code below we use the `gorilla/mux` HTTP router to serve the various proxies. First we'll need some code like the following:

```golang
"github.com/gorilla/mux"
"github.com/styrainc/styra-run-sdk-go/types"

// Initialize the client ..
client := ..

// Initialize rbac ..
myRbac := ..

// Instantiate a new `gorilla/mux` router.
router := mux.NewRouter()

// Extracts HTTP route parameters. For example, for a route 
// like `/foo/{id}`, extracts the `id` parameter.
key := func(key string) types.GetVar {
    return func(r *http.Request) string {
        return mux.Vars(r)[key]
    }
}

// Given a proxy, install it with a specific route and method.
install := func(proxy *types.Proxy, path string) {
    router.HandleFunc(path, proxy.Handler).Methods(proxy.Method)
}

// A function to extract session information. There are a few
// provided, but you will likely want to roll your own.
getSession := types.SessionFromValues(tenant, subject)
```

## Client proxies

The following sections show all proxies minimally configured. Some proxies have additional settings. Please see the code for each proxy for further details. Also, default implementations for some callbacks can be found here:

```golang
"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/shared/defaults.go"
```

### Query

```golang
"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/query"

// Query.
install(query.New(
    &query.Settings{
        Client:  client,
        GetPath: key("path"),
    }), "/query/{path:.*}",
)
```

```
POST /query/tickets/resolve/allow
{
    "input": {
        "tenant": "acmecorp",
        "subject": "alice"
    }
}

->

{
    "result": true
}
```

### Check

```golang
"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/check"

// Check.
install(check.New(
    &check.Settings{
        Client:  client,
        GetPath: key("path"),
    }), "/check/{path:.*}",
)
```

```
POST /check/tickets/resolve/allow
{
    "input": {
        "tenant": "acmecorp",
        "subject": "alice"
    }
}

->

{
    "result": true
}
```

### BatchQuery

```golang
"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/batch_query"
"github.com/styrainc/styra-run-sdk-go/api/v1/proxy/shared"

// Batch query.
install(batch_query.New(
    &batch_query.Settings{
        Client:        client,
        OnModifyInput: shared.DefaultOnModifyInput(getSession),
    }), "/batch_query",
)
```

Here we're using the `OnModifyInput` callback to inject session information into every `input` section of the request body.

```
POST /batch_query
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

## RBAC proxies

The following sections show all proxies minimally configured. Some proxies have additional settings. Please see the code for each proxy for further details. Also, default implementations for some callbacks can be found here:

```golang
"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/shared/defaults.go"
```

### GetRoles

```golang
"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/get_roles"

// Get roles.
install(get_roles.New(
    &get_roles.Settings{
        Rbac:       myRbac,
        GetSession: getSession,
    }), "/roles",
)
```

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

### ListUserBindingsAll

```golang
"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/list_user_bindings_all"

// List user bindings all.
install(list_user_bindings_all.New(
    &list_user_bindings_all.Settings{
        Rbac:       myRbac,
        GetSession: getSession,
    }), "/user_bindings_all",
)
```

```
GET /user_bindings_all

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
            "id": "billy",
            "roles": [
                "VIEWER"
            ]
        },
        {
            "id": "bob",
            "roles": [
                "MASTER"
            ]
        }
    ]
}
```

### ListUserBindings

```golang
"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/list_user_bindings"
"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/shared"

// In this version we control the list of users 
// and lookup bindings for them directly.
users := []*rbac.User{
    {Id: "alice"},
    {Id: "bob"},
    {Id: "bryan"},
    {Id: "cesar"},
    {Id: "emily"},
    {Id: "gary"},
    {Id: "henry"},
    {Id: "kevin"},
    {Id: "lynn"},
    {Id: "jiri"},
    {Id: "larry"},
    {Id: "alan"},
}

// List user bindings.
install(list_user_bindings.New(
    &list_user_bindings.Settings{
        Rbac:       myRbac,
        GetSession: getSession,
        GetUsers:   shared.DefaultGetUsers(users, 3),
    }), "/user_bindings",
)
```

```
GET /user_bindings?page=1

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
        "index": 1,
        "total": 4
    }
}
```

Here, the input bytes to `GetUsers` is the string `"1"` from the `page=1` query parameter. The output should contain a list of users for that page and an `interface{}` that serves as the value for the `"page"` key in the response. Notice that how the users are paged is ultimately up to the programmer. It's important to note, however, that the [frontend SDK](https://github.com/StyraInc/styra-run-sdk-js) assumes the above structure so you must follow suit if you want to use it.

### GetUserBinding

```golang
"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/get_user_binding"

// Get user binding.
install(get_user_binding.New(
    &get_user_binding.Settings{
        Rbac:       myRbac,
        GetSession: getSession,
        GetId:      key("id"),
    }), "/user_bindings/{id}",
)
```

```
GET /user_bindings/alice

->

{
    "result": [
        "ADMIN"
    ]
}
```

### PutUserBinding

```golang
"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/put_user_binding"

// Put user binding.
install(put_user_binding.New(
    &put_user_binding.Settings{
        Rbac:       myRbac,
        GetSession: getSession,
        GetId:      key("id"),
    }), "/user_bindings/{id}",
)
```

```
PUT /user_bindings/alice
[
    "VIEWER"
]

->

{}
```

### DeleteUserBinding

```golang
"github.com/styrainc/styra-run-sdk-go/rbac/v1/proxy/delete_user_binding"

// Delete user binding.
install(delete_user_binding.New(
    &delete_user_binding.Settings{
        Rbac:       myRbac,
        GetSession: getSession,
        GetId:      key("id"),
    }), "/user_bindings/{id}",
)
```

```
DELETE /user_bindings/alice

->

{}
```
