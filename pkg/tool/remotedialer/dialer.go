/*
Copyright 2021 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package remotedialer

import (
	"context"
	"net"
)

type Dialer func(ctx context.Context, network, address string) (net.Conn, error)

func (s *Server) HasSession(clientKey string) bool {
	_, err := s.sessions.getDialer(clientKey)
	return err == nil
}

func (s *Server) Dialer(clientKey string) Dialer {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		d, err := s.sessions.getDialer(clientKey)
		if err != nil {
			return nil, err
		}

		return d(ctx, network, address)
	}
}
