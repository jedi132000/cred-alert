package main

import (
	"cred-alert/revok/api"
	"cred-alert/revok/web"
	"cred-alert/revokpb"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"code.cloudfoundry.org/lager"
	flags "github.com/jessevdk/go-flags"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
	"github.com/tedsuo/rata"
)

type Opts struct {
	Port uint16 `long:"port" default:"8080" description:"Port to listen on."`

	RPCServerAddress string `long:"rpc-server-address" description:"Address for RPC server." required:"true"`
	RPCServerPort    uint16 `long:"rpc-server-port" description:"Port for RPC server." required:"true"`

	CACertPath     string `long:"ca-cert-path" description:"Path to the CA certificate" required:"true"`
	ClientCertPath string `long:"client-cert-path" description:"Path to the client certificate" required:"true"`
	ClientKeyPath  string `long:"client-key-path" description:"Path to the client private key" required:"true"`
}

var (
	indexLayout        *template.Template
	organizationLayout *template.Template
	repositoryLayout   *template.Template
	logger             lager.Logger
)

func init() {
	bs, err := web.Asset("web/templates/index.html")
	if err != nil {
		log.Fatalf("failed loading asset: %s", err.Error())
	}
	indexLayout = template.Must(template.New("index.html").Parse(string(bs)))

	bs, err = web.Asset("web/templates/organization.html")
	if err != nil {
		log.Fatalf("failed loading asset: %s", err.Error())
	}
	organizationLayout = template.Must(template.New("organization.html").Parse(string(bs)))

	bs, err = web.Asset("web/templates/repository.html")
	if err != nil {
		log.Fatalf("failed loading asset: %s", err.Error())
	}
	repositoryLayout = template.Must(template.New("repository.html").Parse(string(bs)))

	logger = lager.NewLogger("credential-count-publisher")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))
}

func main() {
	var opts Opts

	logger.Info("starting")

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	serverAddr := fmt.Sprintf("%s:%d", opts.RPCServerAddress, opts.RPCServerPort)
	listenAddr := fmt.Sprintf(":%d", opts.Port)

	clientCert, err := tls.LoadX509KeyPair(
		opts.ClientCertPath,
		opts.ClientKeyPath,
	)

	rootCertPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile(opts.CACertPath)
	if err != nil {
		log.Fatalf("failed to read ca cert: %s", err.Error())
	}

	ok := rootCertPool.AppendCertsFromPEM(bs)
	if !ok {
		log.Fatalf("failed to append certs from pem: %s", err.Error())
	}

	transportCreds := credentials.NewTLS(&tls.Config{
		ServerName:   opts.RPCServerAddress,
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      rootCertPool,
	})

	dialOption := grpc.WithTransportCredentials(transportCreds)
	conn, err := grpc.Dial(serverAddr, dialOption)
	if err != nil {
		log.Fatalf("failed to create handler: %s", err.Error())
	}
	defer conn.Close()

	revokClient := revokpb.NewRevokClient(conn)

	handler, err := rata.NewRouter(web.Routes, rata.Handlers{
		web.Index:        api.NewIndexHandler(logger, indexLayout, revokClient),
		web.Organization: api.NewOrganizationHandler(logger, organizationLayout, revokClient),
		web.Repository:   api.NewRepositoryHandler(logger, repositoryLayout, revokClient),
	})

	if err != nil {
		log.Fatalf("failed to create handler: %s", err.Error())
	}

	runner := sigmon.New(http_server.New(listenAddr, handler))

	err = <-ifrit.Invoke(runner).Wait()
	if err != nil {
		logger.Error("running-server-failed", err)
	}
}
