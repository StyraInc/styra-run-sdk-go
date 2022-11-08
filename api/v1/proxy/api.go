package proxy

type GetDataResponse struct {
	Result interface{} `json:"result"`
}

type PutDataResponse struct{}

type DeleteDataResponse struct{}

type QueryRequest struct {
	Input interface{} `json:"input"`
}

type QueryResponse struct {
	Result interface{} `json:"result"`
}

type CheckRequest struct {
	Input interface{} `json:"input"`
}

type CheckResponse struct {
	Result bool `json:"result"`
}

type BatchProxyQuery struct {
	Path  string      `json:"path"`
	Input interface{} `json:"input,omitempty"`
}

type BatchProxyRequest struct {
	Items []*BatchProxyQuery `json:"items"`
	Input interface{}        `json:"input,omitempty"`
}

type BatchProxyResult struct {
	Result interface{} `json:"result,omitempty"`
}

type BatchProxyResponse struct {
	Result []*BatchProxyResult `json:"result"`
}
