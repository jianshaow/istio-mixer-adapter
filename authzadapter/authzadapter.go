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
	"strconv"
	"strings"

	"github.com/gogo/googleapis/google/rpc"
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

	// AuthzInfo for policy checking.
	AuthzInfo struct {
		clientID        string
		authzType       string
		targetNamespace string
		targetService   string
		targetPath      string
		requstMethod    string
		requestPriority int
	}

	// AuthzContext for policy checking.
	AuthzContext struct {
		authzInfo     *AuthzInfo
		adapterConfig *config.Params
	}
)

var _ authorization.HandleAuthorizationServiceServer = &AuthzAdapter{}

// HandleAuthorization handler the request
func (s *AuthzAdapter) HandleAuthorization(ctx context.Context, request *authorization.HandleAuthorizationRequest) (*v1beta1.CheckResult, error) {
	log.Infof("received request %v\n", *request)

	context := &AuthzContext{}
	context.authzInfo = &AuthzInfo{}
	context.adapterConfig = &config.Params{}

	err := parseAdapterConfig(context, *request)
	if err != nil {
		return nil, err
	}

	authzInfoStatus := parseAuthzInfo(context, *request)
	if !status.IsOK(authzInfoStatus) {
		return &v1beta1.CheckResult{
			Status: authzInfoStatus,
		}, nil
	}

	priorityStatus := parsePriority(context, *request)
	if !status.IsOK(priorityStatus) {
		return &v1beta1.CheckResult{
			Status: priorityStatus,
		}, nil
	}

	log.Infof("AuthzContext: %+v", *context)

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

func parseAdapterConfig(context *AuthzContext, request authorization.HandleAuthorizationRequest) error {
	if request.AdapterConfig != nil {
		if err := context.adapterConfig.Unmarshal(request.AdapterConfig.Value); err != nil {
			log.Errorf("error unmarshalling adapter config: %v", err)
			return err
		}
	}

	log.Infof("AdapterConfig: %+v\n", *context.adapterConfig)

	return nil
}

func parseAuthzInfo(context *AuthzContext, request authorization.HandleAuthorizationRequest) rpc.Status {
	subjectProps := decodeValueMap(request.Instance.Subject.Properties)

	log.Infof("AuthorizationHeader: %v\n", subjectProps["authorization_header"])

	authzHeader := fmt.Sprintf("%v", subjectProps["authorization_header"])

	if authzHeader == "" {
		log.Info("no authorization header")
		return status.WithUnauthenticated("no authorization header...")
	}

	headerParts := strings.Split(strings.TrimSpace(authzHeader), " ")

	if len(headerParts) == 2 {
		authzType := headerParts[0]
		authzContent := headerParts[1]
		log.Infof("authzContent: %v\n", authzContent)

		if authzType == "Basic" {
			context.authzInfo.authzType = authzType
			s := parsekBasicCredential(context.authzInfo, authzContent)
			if !status.IsOK(s) {
				return s
			}
		}
	} else {
		log.Infof("wrong authorization header: %v\n", authzHeader)
		return status.WithUnauthenticated("wrong authorization header...")
	}

	context.authzInfo.requstMethod = request.Instance.Action.Method
	context.authzInfo.targetNamespace = request.Instance.Action.Namespace
	context.authzInfo.targetPath = request.Instance.Action.Path
	context.authzInfo.targetService = request.Instance.Action.Service

	return status.OK
}

func parsekBasicCredential(authzInfo *AuthzInfo, credential string) rpc.Status {
	decoded, decodeErr := base64.StdEncoding.DecodeString(credential)
	if decodeErr != nil {
		log.Infof("wrong basic credential: %v, error: %v\n", credential, decodeErr)
		return status.WithInvalidArgument("Wrong basic credential...")
	}

	log.Infof("decoded: %s\n", decoded)
	basicAuthzParts := strings.Split(string(decoded), ":")
	clientID := basicAuthzParts[0]
	clientSecret := basicAuthzParts[1]
	s := authenticate(clientID, clientSecret)
	if !status.IsOK(s) {
		return s
	}
	authzInfo.clientID = clientID

	return status.OK
}

func authenticate(clientID string, clientSecret string) rpc.Status {
	return status.OK
}

func parsePriority(context *AuthzContext, request authorization.HandleAuthorizationRequest) rpc.Status {
	actionProps := decodeValueMap(request.Instance.Action.Properties)

	priorityHeader := fmt.Sprintf("%v", actionProps["priority_header"])

	if priorityHeader != "" {
		priority, err := strconv.Atoi(priorityHeader)
		if err != nil {
			log.Infof("wrong priority: %v\n", priorityHeader)
			return status.WithInvalidArgument("Wrong priority header...")
		}
		log.Infof("requestPriority: %v\n", priority)
	}

	log.Infof("Action: %+v\n", *(request.Instance.Action))
	return status.OK
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
