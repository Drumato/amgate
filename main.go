package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"log/slog"

	"github.com/Drumato/amgate/pkg/config"
	"github.com/Drumato/amgate/pkg/server"
	"github.com/cockroachdb/errors"
	"github.com/labstack/echo/v4"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/samber/lo"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Setup
	e := echo.New()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	loglevelS := os.Getenv("LOG_LEVEL")

	var loglevel slog.Level
	switch strings.ToLower(loglevelS) {
	case "debug":
		loglevel = slog.LevelDebug
	case "info":
		loglevel = slog.LevelInfo
	case "warn":
		loglevel = slog.LevelWarn
	case "error":
		loglevel = slog.LevelError
	default:
		loglevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: loglevel}))
	slog.SetDefault(logger)

	kubeconfigPath := os.Getenv("AMGATE_KUBECONFIG_PATH")
	kubeconfigPath = lo.If(kubeconfigPath != "", kubeconfigPath).Else(filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	k8sClient, err := NewClient(kubeconfigPath)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create k8s client", slog.String("error", err.Error()))
		stop()
		os.Exit(1)
	}

	cfg, err := config.LoadFromConfigMap(ctx, k8sClient)
	if err != nil {
		logger.ErrorContext(ctx, "failed to read config from configmap", slog.String("error", err.Error()))
		stop()
		os.Exit(1)
	}

	if err := cfg.ValidateAndDefault(); err != nil {
		logger.ErrorContext(ctx, "failed to validate config", slog.String("error", err.Error()))
		stop()
		os.Exit(1)
	}

	s := server.New(e, &cfg,
		server.WithK8sClient[struct{}](k8sClient),
		server.WithLogger[struct{}](logger),
	)

	// Start server
	if err := s.Start(ctx); err != nil {
		logger.ErrorContext(ctx, "failed to create k8s client", slog.String("error", err.Error()))
		stop()
		os.Exit(1)
	}
}

func NewClient(kubeconfigFilePath string) (client.Client, error) {
	cfg, err := loadKubeconfigFromFile(kubeconfigFilePath)
	if err != nil {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		k8sClient, err := client.New(cfg, client.Options{})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return k8sClient, nil
	}

	k8sClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return k8sClient, nil
}

func loadKubeconfigFromFile(path string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return config, err
}
