package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/styrainc/styra-run-sdk-go/internal/discovery"
	"github.com/styrainc/styra-run-sdk-go/internal/utils"
)

const (
	dataPlaneUrlFormat      = "%s/data/%s"
	dataPlaneBatchUrlFormat = "%s/data_batch"
	batchLimit              = 20
)

type DiscoveryStrategy uint

const (
	Simple DiscoveryStrategy = iota
)

var (
	// Do this to simplify sdk initialization and avoid referencing other packages.
	discoveryStrategyToStrategyType = map[DiscoveryStrategy]discovery.StrategyType{
		Simple: discovery.Simple,
	}
)

type Query struct {
	Path   string
	Input  interface{}
	Result interface{}
	Error  *utils.ErrorResponse
}

type Settings struct {
	Token             string
	Url               string
	DiscoveryStrategy DiscoveryStrategy
	MaxRetries        int
	Client            *http.Client
}

type Client interface {
	GetData(ctx context.Context, path string, data interface{}) error
	PutData(ctx context.Context, path string, data interface{}) error
	DeleteData(ctx context.Context, path string) error
	Query(ctx context.Context, path string, input, result interface{}) error
	Check(ctx context.Context, path string, input interface{}) (bool, error)
	BatchQuery(ctx context.Context, queries []Query, input interface{}) error
}

type client struct {
	settings *Settings
	executor discovery.Executor
}

func New(settings *Settings) Client {
	return &client{
		settings: settings,
		executor: discovery.NewExecutor(
			&discovery.ExecutorSettings{
				Token:        settings.Token,
				Url:          settings.Url,
				StrategyType: discoveryStrategyToStrategyType[settings.DiscoveryStrategy],
				MaxRetries:   settings.MaxRetries,
				Client:       settings.Client,
			},
		),
	}
}

func (c *client) GetData(ctx context.Context, path string, data interface{}) error {
	return c.executor.Try(
		ctx,
		func(url string) error {
			return c.getData(ctx, url, path, data)
		},
	)
}

func (c *client) getData(ctx context.Context, url, path string, result interface{}) error {
	response := &struct {
		Result interface{} `json:"result"`
	}{
		Result: result,
	}

	rest := &utils.Rest{
		Url:     fmt.Sprintf(dataPlaneUrlFormat, url, path),
		Method:  http.MethodGet,
		Client:  c.settings.Client,
		Headers: c.bearer(),
		Decoder: utils.HttpErrorDecoder(response),
	}
	if err := rest.Execute(ctx); err != nil {
		return err
	}

	return nil
}

func (c *client) PutData(ctx context.Context, path string, data interface{}) error {
	return c.executor.Try(
		ctx,
		func(url string) error {
			return c.putData(ctx, url, path, data)
		},
	)
}

func (c *client) putData(ctx context.Context, url, path string, data interface{}) error {
	rest := &utils.Rest{
		Url:     fmt.Sprintf(dataPlaneUrlFormat, url, path),
		Method:  http.MethodPut,
		Client:  c.settings.Client,
		Headers: c.bearerAndJson(),
		Encoder: utils.JsonEncoder(data),
	}

	if err := rest.Execute(ctx); err != nil {
		return err
	}

	return nil
}

func (c *client) DeleteData(ctx context.Context, path string) error {
	return c.executor.Try(
		ctx,
		func(url string) error {
			return c.deleteData(ctx, url, path)
		},
	)
}

func (c *client) deleteData(ctx context.Context, url, path string) error {
	rest := &utils.Rest{
		Url:     fmt.Sprintf(dataPlaneUrlFormat, url, path),
		Method:  http.MethodDelete,
		Client:  c.settings.Client,
		Headers: c.bearer(),
	}
	if err := rest.Execute(ctx); err != nil {
		return err
	}

	return nil
}

func (c *client) Query(ctx context.Context, path string, input, result interface{}) error {
	return c.executor.Try(
		ctx,
		func(url string) error {
			return c.query(ctx, url, path, input, result)
		},
	)
}

func (c *client) query(ctx context.Context, url, path string, input, result interface{}) error {
	request := &struct {
		Input interface{} `json:"input"`
	}{
		Input: input,
	}

	response := &struct {
		Result interface{} `json:"result"`
	}{
		Result: result,
	}

	rest := &utils.Rest{
		Url:     fmt.Sprintf(dataPlaneUrlFormat, url, path),
		Method:  http.MethodPost,
		Client:  c.settings.Client,
		Headers: c.bearerAndJson(),
		Encoder: utils.JsonEncoder(request),
		Decoder: utils.HttpErrorDecoder(response),
	}

	if err := rest.Execute(ctx); err != nil {
		return err
	}

	return nil
}

func (c *client) Check(ctx context.Context, path string, input interface{}) (bool, error) {
	var result interface{}

	if err := c.Query(ctx, path, input, &result); err != nil {
		return false, err
	}

	if value, ok := result.(bool); ok && value {
		return true, nil
	} else {
		return false, nil
	}
}

func (c *client) BatchQuery(ctx context.Context, queries []Query, input interface{}) error {
	return c.executor.Try(
		ctx,
		func(url string) error {
			return c.batchQuery(ctx, url, queries, input)
		},
	)
}

func (c *client) batchQuery(ctx context.Context, url string, queries []Query, input interface{}) error {
	size := len(queries)

	for i := 0; i < size; i += batchLimit {
		var batch []Query

		if i+batchLimit < size {
			batch = queries[i : i+batchLimit]
		} else {
			batch = queries[i:size]
		}

		if err := c.doBatchQuery(ctx, url, batch, input); err != nil {
			return err
		}
	}

	return nil
}

func (c *client) doBatchQuery(ctx context.Context, url string, queries []Query, input interface{}) error {
	type item struct {
		Path  string      `json:"path"`
		Input interface{} `json:"input"`
	}

	request := &struct {
		Items []item      `json:"items"`
		Input interface{} `json:"input"`
	}{
		Items: make([]item, 0),
		Input: input,
	}

	for _, query := range queries {
		request.Items = append(
			request.Items,
			item{
				Path:  query.Path,
				Input: query.Input,
			},
		)
	}

	response := &struct {
		Result []struct {
			Result interface{}          `json:"result"`
			Error  *utils.ErrorResponse `json:"error"`
		} `json:"result"`
	}{}

	rest := &utils.Rest{
		Url:     fmt.Sprintf(dataPlaneBatchUrlFormat, url),
		Method:  http.MethodPost,
		Client:  c.settings.Client,
		Headers: c.bearerAndJson(),
		Encoder: utils.JsonEncoder(request),
		Decoder: utils.HttpErrorDecoder(response),
	}

	if err := rest.Execute(ctx); err != nil {
		return err
	}

	for i, item := range response.Result {
		queries[i].Result = item.Result
		queries[i].Error = item.Error
	}

	return nil
}

func (c *client) bearer() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", c.settings.Token),
	}
}

func (c *client) bearerAndJson() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", c.settings.Token),
		"Content-Type":  "application/json",
	}
}
