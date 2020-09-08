package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/fullstorydev/grpcurl"
	"github.com/gdong42/grpc-mate/metadata"
	"github.com/gdong42/grpc-mate/proxy/reflection"
	"github.com/gdong42/grpc-mate/proxy/stub"
	"google.golang.org/grpc"
)

type Proxy struct {
	cc         *grpc.ClientConn
	reflector  reflection.Reflector
	stub       stub.Stub
	descSource grpcurl.DescriptorSource
}

func (p *Proxy) Invoke(ctx context.Context,
	serviceName, methodName string,
	message []byte,
	md *metadata.Metadata,
) ([]byte, error) {
	invocation, err := p.reflector.CreateInvocation(serviceName, methodName, message)
	if err != nil {
		fmt.Println("123123")
		//return nil, err
	}

	outputMsg, err := p.stub.InvokeRPC(ctx, invocation, md)
	if err != nil {
		return nil, err
	}
	m, err := outputMsg.MarshalJSON()
	if err != nil {
		return nil, errors.New("failed to marshal output JSON")
	}
	return m, err
}
