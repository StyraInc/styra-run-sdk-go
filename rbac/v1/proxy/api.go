package proxy

type GetRolesResponse struct {
	Result []string `json:"result"`
}

type ListUserBinding struct {
	Id    string   `json:"id"`
	Roles []string `json:"roles"`
}

type ListUserBindingsResponse struct {
	Result []*ListUserBinding `json:"result"`
	Page   interface{}        `json:"page"`
}

type GetUserBindingResponse struct {
	Result []string `json:"result"`
}

type PutUserBindingRequest []string

type PutUserBindingResponse struct{}

type DeleteUserBindingResponse struct{}
