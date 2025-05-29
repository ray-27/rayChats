package config

import (
	"raychat/proto/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Client *GrpcManager
)

type GrpcManager struct {
	conn       *grpc.ClientConn
	AuthClient pb.AuthServiceClient
}

func NewGrpcManager(serverAdd string) (*GrpcManager, error) {
	conn, err := grpc.NewClient(serverAdd, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GrpcManager{
		conn:       conn,
		AuthClient: pb.NewAuthServiceClient(conn),
	}, nil
}

// Close closes the connection
func (cm *GrpcManager) Close() error {
	return cm.conn.Close()
}
