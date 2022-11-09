package discovery

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	rerrors "github.com/styrainc/styra-run-sdk-go/internal/errors"
	"github.com/styrainc/styra-run-sdk-go/internal/rest"
)

const (
	gatewayUrlFormat = "%s/gateways"
	maxRetries       = 3
)

var (
	noGatewaysError = errors.New("no gateways")
	badGatewayCodes = map[int]bool{
		502: true, // bad gateway
		503: true, // service unavailable
		504: true, // gateway timeout
	}
)

// TODO: Add sorted strategy.

type StrategyType uint

const (
	Simple StrategyType = iota
)

type Aws struct {
	Region string `json:"region"`
	ZoneId string `json:"zone_id"`
	Zone   string `json:"zone"`
}

type Gateway struct {
	Url string `json:"gateway_url"`
	Aws *Aws   `json:"aws"`
}

type Strategy interface {
	Init(ctx context.Context, gateways []*Gateway) error
	Next()
	Gateway() string
}

type Request func(url string) error

type ExecutorSettings struct {
	Token        string
	Url          string
	StrategyType StrategyType
	MaxRetries   int
	Client       *http.Client
}

type Executor interface {
	Try(ctx context.Context, request Request) error
}

type executor struct {
	settings *ExecutorSettings
	strategy Strategy
	mutex    sync.Mutex
}

func NewExecutor(settings *ExecutorSettings) Executor {
	return &executor{
		settings: settings,
	}
}

func (e *executor) Try(ctx context.Context, request Request) error {
	if err := e.initialized(ctx); err != nil {
		return err
	}

	var err error
	for i := 0; i < e.settings.MaxRetries; i++ {
		gateway := e.strategy.Gateway()

		if err = request(gateway); err == nil {
			return nil
		} else if httpError, ok := err.(rerrors.HttpError); !ok {
			return err
		} else if _, ok := badGatewayCodes[httpError.Code()]; !ok {
			return err
		} else {
			e.mutex.Lock()

			// This must be guarded because multiple requests could
			// be in flight with the same gateway. If both fail, we
			// want to advance to the next gateway exactly once.
			if e.strategy.Gateway() == gateway {
				e.strategy.Next()
			}

			e.mutex.Unlock()
		}
	}

	return err
}

func (e *executor) initialized(ctx context.Context) error {
	if e.strategy != nil {
		return nil
	}

	if e.settings.MaxRetries <= 0 {
		e.settings.MaxRetries = maxRetries
	}

	gateways, err := e.gateways(ctx)
	if err != nil {
		return err
	} else if len(gateways) == 0 {
		return noGatewaysError
	}

	var strategy Strategy
	switch e.settings.StrategyType {
	case Simple:
		strategy = NewSimpleStrategy()
	default:
		strategy = NewSimpleStrategy()
	}

	if err := strategy.Init(ctx, gateways); err != nil {
		return err
	}

	e.strategy = strategy

	return nil
}

func (e *executor) gateways(ctx context.Context) ([]*Gateway, error) {
	response := &struct {
		Result []*Gateway `json:"result"`
	}{}

	rest := &rest.Rest{
		Url:     fmt.Sprintf(gatewayUrlFormat, e.settings.Url),
		Method:  http.MethodGet,
		Client:  e.settings.Client,
		Headers: e.bearer(),
		Decoder: rerrors.HttpErrorDecoder(response),
	}
	if err := rest.Execute(ctx); err != nil {
		return nil, err
	}

	return response.Result, nil
}

func (e *executor) bearer() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", e.settings.Token),
	}
}
