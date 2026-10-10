package domainmodelsapp

import domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"

type Env struct {
	LoggerFormat domainmodelslogger.Format
	LoggerLevel  domainmodelslogger.Level

	PortGrpcServer uint16
	PortHttpServer uint16
}
