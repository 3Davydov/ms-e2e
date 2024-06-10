package e2e

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/3Davydov/ms-proto/golang/order"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CreateOrderTestSuite struct {
	suite.Suite
	compose tc.ComposeStack
}

func (c *CreateOrderTestSuite) SetupSuite() {
	composeFilePaths := []string{"resources/docker-compose.yml"}

	compose, err := tc.NewDockerCompose(composeFilePaths...)
	if err != nil {
		log.Fatalf("failed to init docker compose")
	}
	c.compose = compose

	ctx := context.Background()

	log.Println("Starting Docker Compose...")
	var emt []tc.StackUpOption
	upErr := compose.Up(ctx, emt...)
	if upErr != nil {
		log.Fatalf("failed to bring up docker compose: %v", upErr)
	}
	time.Sleep(5 * time.Second)
}

func (c *CreateOrderTestSuite) Test_Should_Create_Order() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient("localhost:6969", opts...)
	if err != nil {
		log.Fatalf("Failed to connect order service. Err: %v", err)
	}

	defer conn.Close()

	orderClient := order.NewOrderClient(conn)
	createOrderResponse, errCreate := orderClient.Create(context.Background(), &order.CreateOrderRequest{
		UserId: 23,
		OrderItems: []*order.OrderItem{
			{
				ProductCode: "CAM123",
				Quantity:    3,
				UnitPrice:   1.23,
			},
		},
	})
	log.Println("ERR CREATE")
	log.Println(errCreate)
	c.Nil(errCreate)

	getOrderResponse, errGet := orderClient.Get(context.Background(), &order.GetOrderRequest{OrderId: createOrderResponse.OrderId})
	c.Nil(errGet)
	c.Equal(int64(23), getOrderResponse.UserId)
	orderItem := getOrderResponse.OrderItems[0]
	c.Equal(float32(1.23), orderItem.UnitPrice)
	c.Equal(int32(3), orderItem.Quantity)
	c.Equal("CAM123", orderItem.ProductCode)
}

func (c *CreateOrderTestSuite) TearDownSuite() {
	ctx := context.Background()
	var emt []tc.StackDownOption
	execError := c.compose.Down(ctx, emt...)
	if execError != nil {
		log.Fatalf("Could not shutdown compose stack: %v", execError)
	}
}

func TestCreateOrderTestSuite(t *testing.T) {
	suite.Run(t, new(CreateOrderTestSuite))
}
