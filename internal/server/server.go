package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ijaidev/kraftui/internal/api"
	"github.com/ijaidev/kraftui/internal/kraft"
	embedfs "github.com/ijaidev/kraftui/internal/ui"
	"github.com/ijaidev/kraftui/log"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

const (
	autoListenPort     = 5200
	autoListenAttempts = 5
	readHeaderTimeout  = 10 * time.Second
	idleTimeout        = 60 * time.Second
	shutdownTimeout    = 10 * time.Second
)

var _ api.StrictServerInterface = (*Server)(nil)

type Server struct {
	version      string
	kraftVersion string
	kraft        *kraft.Client
	port         int
}

func NewServer(port int, version string, kraftVersion string, kraftClient *kraft.Client) *Server {
	return &Server{
		port:         port,
		version:      version,
		kraftVersion: kraftVersion,
		kraft:        kraftClient,
	}
}

func (s *Server) Listen(ctx context.Context) error {
	ui, err := embedfs.UiHandler()
	if err != nil {
		return fmt.Errorf("embed frontend: %w", err)
	}

	swagger, err := api.GetSpec()
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}
	swagger.Servers = nil

	validatedAPI := buildAPIRouter(s, swagger)

	mux := http.NewServeMux()
	mux.Handle("/api/", validatedAPI)
	mux.Handle("/", ui)

	listener, err := s.getListener()
	if err != nil {
		return err
	}

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	errCh := make(chan error, 1)
	go func() {
		log.G().Info("server listening", "version", s.version, "url", fmt.Sprintf("http://localhost:%d", s.port))
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.G().Info("received interruption, shutting down")
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("error shutting down the server: %w", err)
	}
	<-errCh
	return nil
}

func buildAPIRouter(s api.StrictServerInterface, swagger *openapi3.T) http.Handler {
	strictOptions := api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeAPIError(w, http.StatusBadRequest, "invalid_query", err.Error())
		},
	}

	strictHandler := api.NewStrictHandlerWithOptions(s, nil, strictOptions)
	apiRouter := api.Handler(strictHandler)

	validatorMiddleware := nethttpmiddleware.OapiRequestValidatorWithOptions(swagger, &nethttpmiddleware.Options{
		ErrorHandler: func(w http.ResponseWriter, message string, statusCode int) {
			writeAPIError(w, statusCode, "invalid_query", message)
		},
	})

	return validateUndocumentedQueryParams(swagger)(validatorMiddleware(apiRouter))
}

func validateUndocumentedQueryParams(swagger *openapi3.T) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pathItem := swagger.Paths.Find(r.URL.Path)
			if pathItem != nil {
				op := pathItem.GetOperation(r.Method)
				if op != nil {
					allowed := make(map[string]bool)
					for _, paramRef := range op.Parameters {
						if paramRef != nil && paramRef.Value != nil && paramRef.Value.In == openapi3.ParameterInQuery {
							allowed[paramRef.Value.Name] = true
						}
					}
					for name := range r.URL.Query() {
						if !allowed[name] {
							writeAPIError(w, http.StatusBadRequest, "invalid_query", fmt.Sprintf("unsupported query parameter %q", name))
							return
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Server) getListener() (net.Listener, error) {
	if s.port != 0 {
		return net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.port))
	}

	for i := range autoListenAttempts {
		currentPort := autoListenPort + i
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", currentPort))
		if err == nil {
			s.port = currentPort
			return listener, nil
		}

		if isPortOccupiedError(err) {
			log.G().Warn("port is occupied, checking next", "port", currentPort)
			continue
		}
		return nil, fmt.Errorf("critical binding error on port %d: %w", currentPort, err)
	}

	return nil, fmt.Errorf("exhausted all %d ports starting from %d", autoListenAttempts, autoListenPort)
}

func (s *Server) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	if s.kraft == nil || s.kraftVersion == "" {
		return api.GetHealth503JSONResponse{
			ServiceUnavailableJSONResponse: api.ServiceUnavailableJSONResponse(kraftUnavailableError()),
		}, nil
	}
	return api.GetHealth200JSONResponse{
		Status:       api.Ok,
		Version:      s.version,
		KraftVersion: s.kraftVersion,
	}, nil
}

func (s *Server) ListMachines(ctx context.Context, request api.ListMachinesRequestObject) (api.ListMachinesResponseObject, error) {
	if s.kraft == nil {
		return kraftFailure[api.ListMachines502JSONResponse](kraftUnavailableError()), nil
	}
	items, err := s.kraft.ListMachines(ctx, request.Params)
	if err != nil {
		return kraftFailure[api.ListMachines502JSONResponse](kraftCommandFailed(ctx, "list_machines", err)), nil
	}
	return api.ListMachines200JSONResponse(items), nil
}

func (s *Server) ListPackages(ctx context.Context, request api.ListPackagesRequestObject) (api.ListPackagesResponseObject, error) {
	if s.kraft == nil {
		return kraftFailure[api.ListPackages502JSONResponse](kraftUnavailableError()), nil
	}
	items, err := s.kraft.ListPackages(ctx, request.Params)
	if err != nil {
		return kraftFailure[api.ListPackages502JSONResponse](kraftCommandFailed(ctx, "list_packages", err)), nil
	}
	return api.ListPackages200JSONResponse(items), nil
}

func (s *Server) ListNetworks(ctx context.Context, request api.ListNetworksRequestObject) (api.ListNetworksResponseObject, error) {
	if s.kraft == nil {
		return kraftFailure[api.ListNetworks502JSONResponse](kraftUnavailableError()), nil
	}
	items, err := s.kraft.ListNetworks(ctx, request.Params)
	if err != nil {
		return kraftFailure[api.ListNetworks502JSONResponse](kraftCommandFailed(ctx, "list_networks", err)), nil
	}
	return api.ListNetworks200JSONResponse(items), nil
}

func (s *Server) ListVolumes(ctx context.Context, request api.ListVolumesRequestObject) (api.ListVolumesResponseObject, error) {
	if s.kraft == nil {
		return kraftFailure[api.ListVolumes502JSONResponse](kraftUnavailableError()), nil
	}
	items, err := s.kraft.ListVolumes(ctx, request.Params)
	if err != nil {
		return kraftFailure[api.ListVolumes502JSONResponse](kraftCommandFailed(ctx, "list_volumes", err)), nil
	}
	return api.ListVolumes200JSONResponse(items), nil
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(api.Error{Code: code, Message: message})
}

func kraftUnavailableError() api.Error {
	return api.Error{Code: "kraft_unavailable", Message: "Kraft CLI is unavailable"}
}

func kraftCommandFailed(ctx context.Context, op string, err error) api.Error {
	attrs := []any{"op", op, "error", err}
	var cmdErr *kraft.CommandError
	if errors.As(err, &cmdErr) {
		if cmdErr.Command != "" {
			attrs = append(attrs, "command", cmdErr.Command)
		}
		if cmdErr.Stderr != "" {
			attrs = append(attrs, "stderr", cmdErr.Stderr)
		}
	}
	log.G().ErrorContext(ctx, "Kraft command failed", attrs...)
	return api.Error{Code: "kraft_command_failed", Message: "Kraft command failed"}
}

func kraftFailure[T ~struct{ api.KraftFailureJSONResponse }](err api.Error) T {
	return T{api.KraftFailureJSONResponse(err)}
}
