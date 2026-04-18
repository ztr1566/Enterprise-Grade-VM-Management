package audit

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger initializes the structured JSON audit logger using Uber-Go/Zap.
func InitLogger() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.OutputPaths = []string{"stdout"}
	
	var err error
	Logger, err = config.Build()
	if err != nil {
		panic(err)
	}
}

// LogCredentialAccess records an audit entry for any credential access attempt.
func LogCredentialAccess(userID, vmID string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	
	Logger.Info("credential_access",
		zap.String("user_id", userID),
		zap.String("vm_id", vmID),
		zap.String("status", status),
	)
}
