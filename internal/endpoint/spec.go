package endpoint

import (
	"fmt"
	"strings"
)

type Kind string

const (
	KindTCP   Kind = "tcp"
	KindUnix  Kind = "unix"
	KindNPipe Kind = "npipe"
)

type Spec struct {
	Kind    Kind
	Address string
	Raw     string
}

func Parse(raw string) (Spec, error) {
	switch {
	case strings.HasPrefix(raw, "tcp://"):
		return Spec{Kind: KindTCP, Address: strings.TrimPrefix(raw, "tcp://"), Raw: raw}, nil
	case strings.HasPrefix(raw, "unix://"):
		return Spec{Kind: KindUnix, Address: strings.TrimPrefix(raw, "unix://"), Raw: raw}, nil
	case strings.HasPrefix(raw, "npipe://"):
		return Spec{Kind: KindNPipe, Address: normalizeNPipeAddress(strings.TrimPrefix(raw, "npipe://")), Raw: raw}, nil
	default:
		return Spec{}, fmt.Errorf("unsupported endpoint %q", raw)
	}
}

func (s Spec) String() string {
	return s.Raw
}

func normalizeNPipeAddress(address string) string {
	replacer := strings.NewReplacer("/", `\`)
	return replacer.Replace(address)
}
