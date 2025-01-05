package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func main() {
	shutdown := initTracer()
	defer shutdown()

	userService := os.Getenv("USER_SERVICE")
	projectService := os.Getenv("PROJECT_SERVICE")
	taskService := os.Getenv("TASK_SERVICE")
	notificationService := os.Getenv("NOTIFICATION_SERVICE")
	activityHistoryService := os.Getenv("ACTIVITY_HISTORY_SERVICE")
	port := os.Getenv("PORT")

	if userService == "" || projectService == "" || taskService == "" || notificationService == "" || activityHistoryService == "" {
		log.Fatal("One or more service addresses are not set in the environment variables")
	}

	if port == "" {
		port = "8443"
	}

	http.Handle("/api/users/", enableCORS(otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, userService)
	}), "UsersEndpoint")))

	http.Handle("/api/projects/", enableCORS(otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, projectService)
	}), "ProjectsEndpoint")))

	http.Handle("/api/tasks/", enableCORS(otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, taskService)
	}), "TasksEndpoint")))

	http.Handle("/api/notifications/", enableCORS(otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, notificationService)
	}), "NotificationsEndpoint")))

	http.Handle("/api/activities/", enableCORS(otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, activityHistoryService)
	}), "ActivityHistoryEndpoint")))

	log.Printf("API Gateway is running on HTTPS port %s...", port)
	if err := http.ListenAndServeTLS(":"+port, "certificates/cert.crt", "certificates/cert.key", nil); err != nil {
		log.Fatalf("Failed to start API Gateway HTTPS server: %v", err)
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

func proxyToService(w http.ResponseWriter, r *http.Request, serviceURL string) {
	client := &http.Client{
		Transport: otelhttp.NewTransport(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}),
	}

	if !strings.HasPrefix(serviceURL, "http") {
		serviceURL = "https://" + serviceURL
	}
	newPath := strings.TrimPrefix(r.URL.Path, "/api")
	newPath = strings.TrimSuffix(newPath, "/")
	fullURL := serviceURL + newPath

	log.Printf("Forwarding request to: %s", fullURL)

	req, err := http.NewRequestWithContext(r.Context(), r.Method, fullURL, r.Body)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error forwarding request to %s: %v", fullURL, err)
		http.Error(w, "Failed to forward request: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
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
