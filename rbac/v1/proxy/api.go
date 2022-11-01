package proxy

import (
	rbac "github.com/styrainc/styra-run-sdk-go/rbac/v1"
)

type GetRolesResponse struct {
	Result []string `json:"result"`
}

type ListUserBindingsResponse struct {
	Result []*rbac.UserBinding `json:"result"`
	Page   interface{}         `json:"page,omitempty"`
}

type GetUserBindingResponse struct {
	Result []string `json:"result"`
}

type PutUserBindingRequest []string

type PutUserBindingResponse struct{}

type DeleteUserBindingResponse struct{}
