package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/nathang15/go-tinystore/node"
	"github.com/nathang15/go-tinystore/pb"
	"github.com/nathang15/go-tinystore/ring"
	"github.com/nathang15/go-tinystore/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

const (
	PROD_DB = 0
	TEST_DB = 1
	SUCCESS = "OK"
)

type CacheServer struct {
	Ring            ring.Ring
	router          *gin.Engine
	cache           *store.LRU
	logger          *zap.SugaredLogger
	nodesInfo       node.NodesInfo
	leaderId        string
	nodeId          string
	groupId         string
	shutdownChannel chan bool
	decisionChannel chan string
	synced          chan bool
	electionStatus  bool
	pb.UnimplementedCacheServiceServer
}

type Pair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func InitCacheServer(capacity int, configFile string, verbose bool) (*grpc.Server, *CacheServer) {
	sugaredLogger := GetSugaredZapLogger(verbose)
	nodesInfo := node.LoadNodesConfig(configFile)
	nodeId := node.GetCurrentNodeId(nodesInfo)

	if _, ok := nodesInfo.Nodes[nodeId]; !ok {
		host, _ := os.Hostname()
		nodesInfo.Nodes[nodeId] = node.InitNode(nodeId, host, 8080, 5005)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	lru := store.Init(capacity)

	cacheServer := CacheServer{
		router:          router,
		cache:           &lru,
		logger:          sugaredLogger,
		nodesInfo:       nodesInfo,
		nodeId:          nodeId,
		shutdownChannel: make(chan bool, 1),
		decisionChannel: make(chan string, 1),
		synced:          make(chan bool, 1),
		electionStatus:  NO_ELECTION,
	}

	//routes
	router.GET("/get/:key", cacheServer.GetHandler)
	router.POST("/put", cacheServer.PutHandler)

	//Set up TLS
	credentials, err := LoadTLSCredentials()
	if err != nil {
		sugaredLogger.Fatalf("Failed to load TLS credentials: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(credentials))
	pb.RegisterCacheServiceServer(grpcServer, &cacheServer)
	reflection.Register(grpcServer)
	return grpcServer, &cacheServer
}

// GetHandler Impementation
func (s *CacheServer) GetHandler(c *gin.Context) {
	res := make(chan gin.H)
	go func(ctx *gin.Context) {
		key, err := strconv.Atoi(c.Param("key"))
		if err != nil {
			s.logger.Errorf("Failed to parse key: %v", err)
			return
		}

		value, err := s.cache.Get(c.Param("key"))
		if err != nil {
			res <- gin.H{"error": err.Error()}
		} else {
			res <- gin.H{"key": key, "value": value}
		}
	}(c.Copy())
	c.IndentedJSON(http.StatusOK, <-res)
}

// PutHandler Impementation
func (s *CacheServer) PutHandler(c *gin.Context) {
	var newPair Pair
	if err := c.BindJSON(&newPair); err != nil {
		s.logger.Errorf("unable to deserialize key-value pair from json: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := strconv.Atoi(newPair.Key)
	if err != nil {
		s.logger.Errorf("unable to convert key to integer: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	value, err := strconv.Atoi(newPair.Value)
	if err != nil {
		s.logger.Errorf("unable to convert value to integer: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.cache.Put(newPair.Key, newPair.Value)
	c.IndentedJSON(http.StatusCreated, gin.H{"key": key, "value": value})
}

// Set up mTLS config and creds
func LoadTLSCredentials() (credentials.TransportCredentials, error) {
	pemClientCA, err := os.ReadFile("certs/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to append client CA's certificate")
	}

	//load server cert and key
	sCert, err := tls.LoadX509KeyPair("certs/server-cert.pem", "certs/server-key.pem")
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{sCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		RootCAs:      certPool,
	}

	return credentials.NewTLS(config), nil
}

func NewGrpcClientForNode(node *node.Node) pb.CacheServiceClient {
	creds, err := LoadTLSCredentials()
	if err != nil {
		panic(fmt.Sprintf("Failed to create credentials: %v", err))
	}

	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", node.Host, node.GrpcPort), grpc.WithTransportCredentials(creds))
	if err != nil {
		panic(fmt.Sprintf("Failed to dial: %v", err))
	}

	return pb.NewCacheServiceClient(conn)
}

func GetSugaredZapLogger(verbose bool) *zap.SugaredLogger {
	var level zap.AtomicLevel
	if verbose {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	}

	config := zap.Config{
		Level:            level,
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	return logger.Sugar()
}

func (s *CacheServer) RunHttpServer(port int) {
	s.logger.Infof("HTTP server running on port %d", port)
	s.router.Run(fmt.Sprintf(":%d", port))
}

func (s *CacheServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	value, err := s.cache.Get(req.Key)
	if err != nil {
		return &pb.GetResponse{Data: "key not found"}, nil
	}
	return &pb.GetResponse{Data: value}, nil
}

func (s *CacheServer) Put(ctx context.Context, req *pb.PutRequest) (*empty.Empty, error) {
	s.cache.Put(req.Key, req.Value)
	return &empty.Empty{}, nil
}

func (s *CacheServer) ServerInitCacheClient(server_host string, server_port int) pb.CacheServiceClient {
	creds, err := LoadTLSCredentials()
	if err != nil {
		s.logger.Fatalf("failed to create credentials: %v", err)
	}

	var kacp = keepalive.ClientParameters{
		Time:                8 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: true,
	}

	addr := fmt.Sprintf("%s:%d", server_host, server_port)
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(creds),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithTimeout(time.Duration(time.Second)),
	)
	if err != nil {
		panic(err)
	}

	return pb.NewCacheServiceClient(conn)
}

func (s *CacheServer) RegisterNodeInternal() {
	for {
		if s.leaderId == NO_LEADER {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	leader := s.nodesInfo.Nodes[s.leaderId]
	local_node := s.nodesInfo.Nodes[s.nodeId]
	req := pb.Node{
		Id:       local_node.Id,
		Host:     local_node.Host,
		RestPort: local_node.RestPort,
		GrpcPort: local_node.GrpcPort,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if leader.GrpcClient == nil {
		client := s.ServerInitCacheClient(leader.Host, int(leader.GrpcPort))
		leader.SetGrpcClient(client)
	}

	_, err := leader.GrpcClient.RegisterNodeWithCluster(ctx, &req)
	if err != nil {
		s.logger.Infof("error registering node %s with cluster: %v", s.nodeId, err)
		return
	}
	s.logger.Infof("registered node %s with cluster", s.nodeId)
}

func (s *CacheServer) SetAllGrpcClients() {
	for _, node := range s.nodesInfo.Nodes {
		c := s.ServerInitCacheClient(node.Host, int(node.GrpcPort))
		node.SetGrpcClient(c)
		s.logger.Infof("created grpc client to %s", node.Id)
	}
}
