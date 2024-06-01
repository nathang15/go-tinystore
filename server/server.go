package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nathang15/go-tinystore/node"
	"github.com/nathang15/go-tinystore/pb"
	"github.com/nathang15/go-tinystore/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type CacheServer struct {
	router          *gin.Engine
	cache           *store.LRU
	logger          *zap.SugaredLogger
	nodesInfo       node.NodesInfo
	leaderId        int32
	nodeId          int32
	shutdownChannel chan bool
	decisionChannel chan int32
	synced          chan bool
	mutex           sync.Mutex
	electionStatus  bool
	pb.UnimplementedCacheServiceServer
}

type Pair struct {
	Key   int `json:"key"`
	Value int `json:"value"`
}

// Create gRPC server
func InitCacheServer(capacity int, configFile string, verbose bool) (*grpc.Server, *CacheServer) {
	sugaredLogger := GetSugaredZapLogger(verbose)
	nodesInfo := node.LoadNodesConfig(configFile)
	nodeId := node.GetCurrentNodeId(nodesInfo)

	router := gin.Default()

	lru := store.Init(capacity)

	cacheServer := CacheServer{
		router:          router,
		cache:           &lru,
		logger:          sugaredLogger,
		nodesInfo:       nodesInfo,
		nodeId:          nodeId,
		leaderId:        NO_LEADER,
		shutdownChannel: make(chan bool, 1),
		decisionChannel: make(chan int32, 1),
		synced:          make(chan bool, 1),
		electionStatus:  NO_ELECTION,
	}

	//routes
	router.GET("/cache/:key", cacheServer.GetHandler)
	router.POST("/cache/:key/:value", cacheServer.PutHandler)

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
	key, err := strconv.Atoi(c.Param("key"))
	if err != nil {
		s.logger.Errorf("Failed to parse key: %v", err)
		return
	}

	value, err := s.cache.Get(key)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	c.IndentedJSON(http.StatusOK, value)
}

// PutHandler Impementation
func (s *CacheServer) PutHandler(c *gin.Context) {
	var p Pair
	if err := c.BindJSON(&p); err != nil {
		s.logger.Errorf("Failed to bind JSON: %v", err)
		return
	}
	s.cache.Put(p.Key, p.Value)
	c.IndentedJSON(http.StatusCreated, p)
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

	channel, err := grpc.NewClient(fmt.Sprintf("%s:%d", node.Host, node.Port), grpc.WithTransportCredentials(creds))
	if err != nil {
		panic(err)
	}

	return pb.NewCacheServiceClient(channel)
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
