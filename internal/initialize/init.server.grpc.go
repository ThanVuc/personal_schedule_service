package initialize

import (
	"context"
	"fmt"
	"net"
	"personal_schedule_service/global"
	"personal_schedule_service/pkg/settings"
	"sync"

	"github.com/thanvuc/go-core-lib/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type AuthServer struct {
	logger log.Logger
	config *settings.Server
}

func NewAuthService() *AuthServer {
	return &AuthServer{
		logger: global.Logger,
		config: &global.Config.Server,
	}
}

func (as *AuthServer) runServers(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go as.runServiceServer(ctx, wg)
}

// create server factory
func (as *AuthServer) createServer() *grpc.Server {
	server := grpc.NewServer()

	return server
}

func (as *AuthServer) runServiceServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	lis, err := as.createListener()
	if err != nil {
		as.logger.Error("Failed to create listener",
			"", zap.Error(err),
		)
		return
	}

	// Create a new gRPC server instance
	server := as.createServer()

	// Gracefully handle server shutdown
	go as.gracefullyShutdownServer(ctx, server)

	// Server listening on the specified port
	as.serverListening(server, lis)
}

func (as *AuthServer) gracefullyShutdownServer(ctx context.Context, server *grpc.Server) {
	<-ctx.Done()
	as.logger.Info("gRPC server is shutting down...", "")
	server.GracefulStop()
	as.logger.Info("gRPC server stopped gracefully!", "")
}

func (as *AuthServer) serverListening(server *grpc.Server, lis net.Listener) {
	as.logger.Info(fmt.Sprintf("gRPC server listening on %s:%d", as.config.Host, lis.Addr().(*net.TCPAddr).Port), "")
	if err := server.Serve(lis); err != nil {
		if err == grpc.ErrServerStopped {
			as.logger.Info("gRPC server exited normally", "")
		} else {
			as.logger.Error("Failed to serve gRPC server",
				"", zap.Error(err),
			)
		}
	}
}

func (as *AuthServer) createListener() (net.Listener, error) {
	err := error(nil)
	lis := net.Listener(nil)
	lis, err = net.Listen("tcp", fmt.Sprintf("%s:%d", as.config.Host, as.config.PersonalSchedulePort))
	if err != nil {
		as.logger.Error("Failed to listen: %v", "", zap.Error(err))
		return nil, err
	}

	return lis, nil
}
