// nolint:lll
// Generates the authzadapter adapter's resource yaml. It contains the adapter's configuration, name,
// supported template names (metric in this case), and whether it is session or no-session based.
//go:generate $GOPATH/src/istio.io/istio/bin/mixer_codegen.sh -a mixer/adapter/authzadapter/config/config.proto -x "-s=false -n authzadapter -t authorization"

package authzadapter

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"

	"istio.io/api/mixer/adapter/model/v1beta1"
	policy "istio.io/api/policy/v1beta1"
	"istio.io/istio/mixer/adapter/authzadapter/config"
	"istio.io/istio/mixer/pkg/status"
	"istio.io/istio/mixer/template/authorization"
	"istio.io/pkg/log"
)

type (
	// Server is basic server interface
	Server interface {
		Addr() string
		Close() error
		Run(shutdown chan error)
	}

	// AuthzAdapter supports authorization template.
	AuthzAdapter struct {
		listener net.Listener
		server   *grpc.Server
	}
)

var _ authorization.HandleAuthorizationServiceServer = &AuthzAdapter{}

// HandleAuthorization handler the request
func (s *AuthzAdapter) HandleAuthorization(ctx context.Context, r *authorization.HandleAuthorizationRequest) (*v1beta1.CheckResult, error) {
	log.Infof("received request %v\n", *r)

	cfg := &config.Params{}

	if r.AdapterConfig != nil {
		if err := cfg.Unmarshal(r.AdapterConfig.Value); err != nil {
			log.Errorf("error unmarshalling adapter config: %v", err)
			return nil, err
		}
	}

	log.Infof("Config: %+v\n", *cfg)

	subjectProps := decodeValueMap(r.Instance.Subject.Properties)
	log.Infof("AuthorizationHeader: %v\n", subjectProps["authorization_header"])

	authzHeader := fmt.Sprintf("%v", subjectProps["authorization_header"])
	if authzHeader != "" {
		headerParts := strings.Split(strings.TrimSpace(authzHeader), " ")

		if len(headerParts) >= 2 {
			authzType := headerParts[0]
			authzContent := headerParts[1]
			log.Infof("authzContent: %v\n", authzContent)

			if authzType == "Basic" {
				decoded, _ := base64.StdEncoding.DecodeString(authzContent)
				log.Infof("decoded: %s\n", decoded)
				basicAuthzParts := strings.Split(string(decoded), ":")
				clientID := basicAuthzParts[0]
				clientSecret := basicAuthzParts[1]

				log.Infof("clientID: %v\n", clientID)
				log.Infof("clientSecret: %v\n", clientSecret)
			}
		}
	}

	log.Infof("Action: %+v\n", *(r.Instance.Action))

	return &v1beta1.CheckResult{
		Status: status.OK,
	}, nil
}

// Addr returns the listening address of the server
func (s *AuthzAdapter) Addr() string {
	return s.listener.Addr().String()
}

// Run starts the server run
func (s *AuthzAdapter) Run(shutdown chan error) {
	shutdown <- s.server.Serve(s.listener)
}

// Close gracefully shuts down the server; used for testing
func (s *AuthzAdapter) Close() error {
	if s.server != nil {
		s.server.GracefulStop()
	}

	if s.listener != nil {
		_ = s.listener.Close()
	}

	return nil
}

// NewAuthzAdapter creates a new adapter that listens at provided port.
func NewAuthzAdapter(addr string) (Server, error) {
	if addr == "" {
		addr = "0"
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", addr))
	if err != nil {
		return nil, fmt.Errorf("unable to listen on socket: %v", err)
	}
	s := &AuthzAdapter{
		listener: listener,
	}
	fmt.Printf("listening on \"%v\"\n", s.Addr())
	s.server = grpc.NewServer()
	authorization.RegisterHandleAuthorizationServiceServer(s.server, s)
	return s, nil
}

func decodeValue(in interface{}) interface{} {
	switch t := in.(type) {
	case *policy.Value_StringValue:
		return t.StringValue
	case *policy.Value_Int64Value:
		return t.Int64Value
	case *policy.Value_DoubleValue:
		return t.DoubleValue
	default:
		return fmt.Sprintf("%v", in)
	}
}

func decodeValueMap(in map[string]*policy.Value) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = decodeValue(v.GetValue())
	}
	return out
}
