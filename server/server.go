package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/nathang15/go-tinystore/node"
	"github.com/nathang15/go-tinystore/pb"
	"github.com/nathang15/go-tinystore/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type CacheServer struct {
	router          *gin.Engine
	cache           *store.LRU
	logger          *zap.SugaredLogger
	nodesInfo       node.NodesInfo
	leaderId        string
	nodeId          string
	shutdownChannel chan bool
	decisionChannel chan string
	synced          chan bool
	mutex           sync.Mutex
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

	router := gin.New()
	router.Use(gin.Recovery())

	lru := store.Init(capacity)

	cacheServer := CacheServer{
		router:          router,
		cache:           &lru,
		logger:          sugaredLogger,
		nodesInfo:       nodesInfo,
		nodeId:          nodeId,
		leaderId:        NO_LEADER,
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

		value, err := s.cache.Get(key)
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

	s.cache.Put(key, value)
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

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node.Host, node.Port), grpc.WithTransportCredentials(creds))
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

// gRPC handler for getting item from cache
func (s *CacheServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	key, err := strconv.Atoi(req.Key)
	if err != nil {
		s.logger.Errorf("unable to convert %s to integer", req.Key)
		return nil, status.Errorf(codes.InvalidArgument, "key must be integer")
	}
	value, err := s.cache.Get(key)
	if err != nil {
		return &pb.GetResponse{Data: "key not found"}, nil
	}
	strvalue := strconv.Itoa(value)
	return &pb.GetResponse{Data: strvalue}, nil
}

// gRPC handler for putting item in cache
func (s *CacheServer) Put(ctx context.Context, req *pb.PutRequest) (*empty.Empty, error) {
	key, err := strconv.Atoi(req.Key)
	if err != nil {
		s.logger.Errorf("unable to convert key %s to integer", req.Key)
		return nil, status.Errorf(codes.InvalidArgument, "key must be integer")
	}
	value, err := strconv.Atoi(req.Value)
	if err != nil {
		s.logger.Errorf("unable to convert value %s to integer", req.Value)
		return nil, status.Errorf(codes.InvalidArgument, "value must be integer")
	}
	s.cache.Put(key, value)
	return &empty.Empty{}, nil
}
