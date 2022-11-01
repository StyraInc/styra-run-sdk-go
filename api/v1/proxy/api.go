package proxy

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
