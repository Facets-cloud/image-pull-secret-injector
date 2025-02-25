package main

import (
	"flag"
	"os"
	"strings"
	"time"

	uberzap "go.uber.org/zap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func main() {
	var secretList string
	var debug bool

	flag.StringVar(&secretList, "secrets", "registry-secret-test-google-container,registry-secret-test-pass-azure", "Comma-separated list of secret names to inject")
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.Parse()

	opts := zap.Options{
		Development: true,
	}
	if !debug {
		opts.Level = uberzap.NewAtomicLevelAt(uberzap.InfoLevel)
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	var logger = log.Log.WithName("image-pull-secrets-injector")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: 9443,
		}),
		Metrics: server.Options{
			BindAddress: ":8080",
		},
		HealthProbeBindAddress: ":8081",
		LeaderElection:         true,
		LeaderElectionID:       "image-pull-secret-injector-leader",
		LeaseDuration:          &[]time.Duration{15 * time.Second}[0],
		RenewDeadline:          &[]time.Duration{10 * time.Second}[0],
		RetryPeriod:            &[]time.Duration{2 * time.Second}[0],
	})
	if err != nil {
		logger.Error(err, "could not create manager")
		os.Exit(1)
	}

	// Add health check endpoints
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	webhookServer := mgr.GetWebhookServer()
	decoder := admission.NewDecoder(mgr.GetScheme())

	secretNames := strings.Split(secretList, ",")
	for i := range secretNames {
		secretNames[i] = strings.TrimSpace(secretNames[i])
	}
	logger.Info("Configured secrets to inject", "secrets", secretNames)

	podMutator := &podMutator{
		SecretNames: secretNames,
		Client:      mgr.GetClient(),
		Log:         mgr.GetLogger(),
	}
	podMutator.InjectDecoder(decoder)

	webhookServer.Register("/mutate-v1-pod", &webhook.Admission{Handler: podMutator})

	ctx := ctrl.SetupSignalHandler()
	if err := mgr.Start(ctx); err != nil {
		logger.Error(err, "could not start manager")
		os.Exit(1)
	}
}
