package server

import (
	"context"
	"encoding/json"
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

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	listner, err := s.getListner()
	if err != nil {
		return err
	}
	errCh := make(chan error, 1)
	go func() {
		log.G.Info("server listening", "version", s.version, "url", fmt.Sprintf("http://localhost:%d", s.port))
		if err := srv.Serve(listner); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.G.Info("received interruption, shutting down")
	case err := <-errCh:
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err = srv.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("error shutting down the server: %w", err)
	}
	return nil
}

func buildAPIRouter(s api.StrictServerInterface, swagger *openapi3.T) http.Handler {
	strictOptions := api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.Error{
				Code:    "invalid_query",
				Message: err.Error(),
			})
		},
	}

	strictHandler := api.NewStrictHandlerWithOptions(s, nil, strictOptions)
	apiRouter := api.Handler(strictHandler)

	validatorMiddleware := nethttpmiddleware.OapiRequestValidatorWithOptions(swagger, &nethttpmiddleware.Options{
		ErrorHandler: func(w http.ResponseWriter, message string, statusCode int) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			_ = json.NewEncoder(w).Encode(api.Error{
				Code:    "invalid_query",
				Message: message,
			})
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
							w.Header().Set("Content-Type", "application/json")
							w.WriteHeader(http.StatusBadRequest)
							_ = json.NewEncoder(w).Encode(api.Error{
								Code:    "invalid_query",
								Message: fmt.Sprintf("unsupported query parameter %q", name),
							})
							return
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Server) getListner() (net.Listener, error) {
	maxAttempts := 5

	if s.port != 0 {
		address := fmt.Sprintf("127.0.0.1:%d", s.port)
		return net.Listen("tcp", address)
	}

	startPort := 5200
	for i := range maxAttempts {
		currentPort := startPort + i
		address := fmt.Sprintf("127.0.0.1:%d", currentPort)

		listener, err := net.Listen("tcp", address)
		if err == nil {
			s.port = currentPort
			return listener, nil
		}

		if isPortOccupiedError(err) {
			log.G.Warn("port is occupied, checking next", "port", currentPort)
		} else {
			return nil, fmt.Errorf("critical binding error on port %d: %w", currentPort, err)
		}
	}

	return nil, fmt.Errorf("exhausted all %d ports starting from %d", maxAttempts, startPort)
}

func (s *Server) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	if s.kraft == nil || s.kraftVersion == "" {
		return api.GetHealth503JSONResponse{
			ServiceUnavailableJSONResponse: api.ServiceUnavailableJSONResponse(api.Error{Code: "kraft_unavailable", Message: "Kraft CLI is unavailable"}),
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
		return api.ListMachines502JSONResponse{
			KraftFailureJSONResponse: api.KraftFailureJSONResponse(api.Error{Code: "kraft_unavailable", Message: "Kraft CLI is unavailable"}),
		}, nil
	}
	items, err := s.kraft.ListMachines(ctx, request.Params)
	if err != nil {
		log.G.Error("Kraft command failed", "error", err)
		return api.ListMachines502JSONResponse{
			KraftFailureJSONResponse: api.KraftFailureJSONResponse(api.Error{Code: "kraft_command_failed", Message: "Kraft command failed"}),
		}, nil
	}
	return api.ListMachines200JSONResponse(items), nil
}

func (s *Server) ListPackages(ctx context.Context, request api.ListPackagesRequestObject) (api.ListPackagesResponseObject, error) {
	if s.kraft == nil {
		return api.ListPackages502JSONResponse{
			KraftFailureJSONResponse: api.KraftFailureJSONResponse(api.Error{Code: "kraft_unavailable", Message: "Kraft CLI is unavailable"}),
		}, nil
	}
	items, err := s.kraft.ListPackages(ctx, request.Params)
	if err != nil {
		log.G.Error("Kraft command failed", "error", err)
		return api.ListPackages502JSONResponse{
			KraftFailureJSONResponse: api.KraftFailureJSONResponse(api.Error{Code: "kraft_command_failed", Message: "Kraft command failed"}),
		}, nil
	}
	return api.ListPackages200JSONResponse(items), nil
}

func (s *Server) ListNetworks(ctx context.Context, request api.ListNetworksRequestObject) (api.ListNetworksResponseObject, error) {
	if s.kraft == nil {
		return api.ListNetworks502JSONResponse{
			KraftFailureJSONResponse: api.KraftFailureJSONResponse(api.Error{Code: "kraft_unavailable", Message: "Kraft CLI is unavailable"}),
		}, nil
	}
	items, err := s.kraft.ListNetworks(ctx, request.Params)
	if err != nil {
		log.G.Error("Kraft command failed", "error", err)
		return api.ListNetworks502JSONResponse{
			KraftFailureJSONResponse: api.KraftFailureJSONResponse(api.Error{Code: "kraft_command_failed", Message: "Kraft command failed"}),
		}, nil
	}
	return api.ListNetworks200JSONResponse(items), nil
}

func (s *Server) ListVolumes(ctx context.Context, request api.ListVolumesRequestObject) (api.ListVolumesResponseObject, error) {
	if s.kraft == nil {
		return api.ListVolumes502JSONResponse{
			KraftFailureJSONResponse: api.KraftFailureJSONResponse(api.Error{Code: "kraft_unavailable", Message: "Kraft CLI is unavailable"}),
		}, nil
	}
	items, err := s.kraft.ListVolumes(ctx, request.Params)
	if err != nil {
		log.G.Error("Kraft command failed", "error", err)
		return api.ListVolumes502JSONResponse{
			KraftFailureJSONResponse: api.KraftFailureJSONResponse(api.Error{Code: "kraft_command_failed", Message: "Kraft command failed"}),
		}, nil
	}
	return api.ListVolumes200JSONResponse(items), nil
}
