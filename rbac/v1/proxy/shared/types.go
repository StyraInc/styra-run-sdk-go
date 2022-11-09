package shared

import (
	"net/http"

	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
)

type GetUsers func(r *http.Request, bytes []byte) ([]*rbac.User, interface{}, error)

type OnBeforeAccess func(user *rbac.User) (int, error)
