package grpc

import roomsv1 "github.com/olegetoya/booking/protos/gen/go/hotelsvc/rooms"

func (s *server) GetAll(ctx context.Context, req *roomsv1.GetAllRequest) (*roomsv1.GetAllResponse, error) {
	return &roomsv1.GetAllResponse{}, nil
}

func (s *server) Get(ctx context.Context, req *roomsv1.GetRequest) (*roomsv1.GetResponse, error) {
	return &roomsv1.GetResponse{}, nil
}
