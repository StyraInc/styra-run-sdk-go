package shared

import rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"

type OnUserBindingTouched func(user *rbac.User) (int, error)
