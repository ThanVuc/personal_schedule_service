package initialize

import (
	"context"
	"fmt"
	"net"
	"personal_schedule_service/global"
	"personal_schedule_service/internal/grpc/controller"
	"personal_schedule_service/internal/grpc/wire"
	"personal_schedule_service/pkg/settings"
	"personal_schedule_service/proto/personal_schedule"
	"sync"

	"github.com/thanvuc/go-core-lib/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type PersonalScheduleServer struct {
	logger             log.Logger
	config             *settings.Server
	labelServiceServer *controller.LabelController
	goalServiceServer  *controller.GoalController
}

func NewPersonalScheduleService() *PersonalScheduleServer {
	return &PersonalScheduleServer{
		logger:             global.Logger,
		config:             &global.Config.Server,
		labelServiceServer: wire.InjectLabelController(),
		goalServiceServer:  wire.InjectGoalController(),
	}
}

func (ps *PersonalScheduleServer) runServers(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go ps.runServiceServer(ctx, wg)
}

// create server factory
func (ps *PersonalScheduleServer) createServer() *grpc.Server {
	server := grpc.NewServer()

	personal_schedule.RegisterLabelServiceServer(server, ps.labelServiceServer)
	personal_schedule.RegisterGoalServiceServer(server, ps.goalServiceServer)

	return server
}

func (ps *PersonalScheduleServer) runServiceServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	lis, err := ps.createListener()
	if err != nil {
		ps.logger.Error("Failed to create listener",
			"", zap.Error(err),
		)
		return
	}

	// Create a new gRPC server instance
	server := ps.createServer()

	// Gracefully handle server shutdown
	go ps.gracefullyShutdownServer(ctx, server)

	// Server listening on the specified port
	ps.serverListening(server, lis)
}

func (ps *PersonalScheduleServer) gracefullyShutdownServer(ctx context.Context, server *grpc.Server) {
	<-ctx.Done()
	ps.logger.Info("gRPC server is shutting down...", "")
	server.GracefulStop()
	ps.logger.Info("gRPC server stopped gracefully!", "")
}

func (ps *PersonalScheduleServer) serverListening(server *grpc.Server, lis net.Listener) {
	ps.logger.Info(fmt.Sprintf("gRPC server listening on %s:%d", ps.config.Host, lis.Addr().(*net.TCPAddr).Port), "")
	if err := server.Serve(lis); err != nil {
		if err == grpc.ErrServerStopped {
			ps.logger.Info("gRPC server exited normally", "")
		} else {
			ps.logger.Error("Failed to serve gRPC server",
				"", zap.Error(err),
			)
		}
	}
}

func (ps *PersonalScheduleServer) createListener() (net.Listener, error) {
	err := error(nil)
	lis := net.Listener(nil)
	lis, err = net.Listen("tcp", fmt.Sprintf("%s:%d", ps.config.Host, ps.config.PersonalSchedulePort))
	if err != nil {
		ps.logger.Error("Failed to listen: %v", "", zap.Error(err))
		return nil, err
	}

	return lis, nil
}
