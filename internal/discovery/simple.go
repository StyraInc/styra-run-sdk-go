package discovery

import "context"

type simple struct {
	gateways []*Gateway
	index    int
}

func NewSimpleStrategy() Strategy {
	return &simple{}
}

func (s *simple) Init(ctx context.Context, gateways []*Gateway) error {
	s.gateways = gateways

	return nil
}

func (s *simple) Next() {
	s.index++

	if s.index == len(s.gateways) {
		s.index = 0
	}
}

func (s *simple) Gateway() string {
	return s.gateways[s.index].Url
}
