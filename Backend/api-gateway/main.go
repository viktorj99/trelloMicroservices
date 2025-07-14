package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sony/gobreaker"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var (
	userService            string
	projectService         string
	taskService            string
	notificationService    string
	port                   string
	serviceCircuitBreakers = make(map[string]*gobreaker.CircuitBreaker)
)

func main() {
	shutdown := initTracer()
	defer shutdown()

	userService = os.Getenv("USER_SERVICE")
	projectService = os.Getenv("PROJECT_SERVICE")
	taskService = os.Getenv("TASK_SERVICE")
	notificationService = os.Getenv("NOTIFICATION_SERVICE")
	port = os.Getenv("PORT")
	if port == "" {
		port = "8443"
	}

	if userService == "" || projectService == "" || taskService == "" || notificationService == "" {
		log.Fatal("Missing service env vars")
	}

	initializeCircuitBreakers()

	http.Handle("/api/users/", enableCORS(otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyWithMechanisms(w, r, userService)
	}), "UsersEndpoint")))

	http.Handle("/api/projects/", enableCORS(otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyWithMechanisms(w, r, projectService)
	}), "ProjectsEndpoint")))

	http.Handle("/api/tasks/", enableCORS(otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyWithMechanisms(w, r, taskService)
	}), "TasksEndpoint")))

	http.Handle("/api/notifications/", enableCORS(otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyWithMechanisms(w, r, notificationService)
	}), "NotificationsEndpoint")))

	log.Printf("API Gateway running on HTTPS port %s...", port)
	if err := http.ListenAndServeTLS(":"+port, "certificates/cert.crt", "certificates/cert.key", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func enableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, user_id, captcha_token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func proxyWithMechanisms(w http.ResponseWriter, r *http.Request, serviceURL string) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	breaker, ok := serviceCircuitBreakers[serviceURL]
	if !ok || breaker == nil {
		log.Printf("Circuit breaker not found for service: %s", serviceURL)
		fallbackResponse(w)
		return
	}

	_, err := breaker.Execute(func() (interface{}, error) {
		return nil, callWithRetry(w, r, serviceURL)
	})

	if err != nil {
		log.Printf("Request failed for service %s: %v", serviceURL, err)
		fallbackResponse(w)
	}
}

func callWithRetry(w http.ResponseWriter, r *http.Request, serviceURL string) error {
	backoffConfig := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3)
	return backoff.Retry(func() error {
		return forwardRequest(w, r, serviceURL)
	}, backoffConfig)
}

func forwardRequest(w http.ResponseWriter, r *http.Request, serviceURL string) error {
	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: otelhttp.NewTransport(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}),
	}

	newPath := strings.TrimPrefix(r.URL.Path, "/api")
	newPath = strings.TrimSuffix(newPath, "/")
	if !strings.HasPrefix(serviceURL, "http") {
		serviceURL = "https://" + serviceURL
	}
	fullURL := serviceURL + newPath

	log.Printf("Forwarding request to: %s", fullURL)

	req, err := http.NewRequestWithContext(r.Context(), r.Method, fullURL, r.Body)
	if err != nil {
		return err
	}
	req.Header = r.Header

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}

func fallbackResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte("Service temporarily unavailable. Please try again later."))
}

func initializeCircuitBreakers() {
	serviceURLs := []string{
		userService,
		projectService,
		taskService,
		notificationService,
	}
	for _, serviceURL := range serviceURLs {
		serviceCircuitBreakers[serviceURL] = gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        serviceURL,
			MaxRequests: 5,
			Interval:    time.Minute,
			Timeout:     30 * time.Second,
		})
	}
}

func initTracer() func() {
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	if jaegerEndpoint == "" {
		jaegerEndpoint = "http://jaeger:14268/api/traces"
	}

	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		log.Fatalf("failed to create Jaeger exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("api-gateway"),
		)),
	)

	otel.SetTracerProvider(tp)
	return func() {
		_ = tp.Shutdown(context.Background())
	}
}
