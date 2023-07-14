package gin

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-es-indexer-go/api/groups"
	"github.com/multiversx/mx-chain-es-indexer-go/api/shared"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
)

const (
	webServerOffString = "off"
)

// ArgsWebServer holds the arguments needed for a webServer
type ArgsWebServer struct {
	Facade    shared.FacadeHandler
	ApiConfig config.ApiRoutesConfig
}

type webServer struct {
	sync.RWMutex
	facade     shared.FacadeHandler
	apiConfig  config.ApiRoutesConfig
	groups     map[string]shared.GroupHandler
	httpServer shared.HttpServerCloser
}

// NewWebServer will create a new instance of the webServer
func NewWebServer(args ArgsWebServer) (*webServer, error) {
	return &webServer{
		facade:    args.Facade,
		apiConfig: args.ApiConfig,
	}, nil
}

// StartHttpServer will start the http server
func (ws *webServer) StartHttpServer() error {
	ws.Lock()
	defer ws.Unlock()

	apiInterface := ws.apiConfig.RestApiInterface
	if apiInterface == webServerOffString {
		log.Debug("web server is turned off")
		return nil
	}

	var engine *gin.Engine
	gin.DefaultWriter = &ginWriter{}
	gin.DefaultErrorWriter = &ginErrorWriter{}
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	engine = gin.Default()
	cfg := cors.DefaultConfig()
	cfg.AllowAllOrigins = true
	cfg.AddAllowHeaders("Authorization")
	engine.Use(cors.New(cfg))

	err := ws.createGroups()
	if err != nil {
		return err
	}

	ws.registerRoutes(engine)

	s := &http.Server{Addr: apiInterface, Handler: engine}
	log.Debug("creating gin web sever", "interface", apiInterface)
	ws.httpServer, err = NewHttpServer(s)
	if err != nil {
		return err
	}

	log.Debug("starting web server")
	go ws.httpServer.Start()

	return nil
}

func (ws *webServer) createGroups() error {
	groupsMap := make(map[string]shared.GroupHandler)

	statusGroup, err := groups.NewStatusGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["status"] = statusGroup

	ws.groups = groupsMap

	return nil
}

func (ws *webServer) registerRoutes(ginRouter *gin.Engine) {
	for groupName, groupHandler := range ws.groups {
		log.Debug("registering gin API group", "group name", groupName)
		ginGroup := ginRouter.Group(fmt.Sprintf("/%s", groupName))
		groupHandler.RegisterRoutes(ginGroup, ws.apiConfig)
	}
}

// Close will handle the closing of inner components
func (ws *webServer) Close() error {
	var err error
	ws.Lock()
	if ws.httpServer != nil {
		err = ws.httpServer.Close()
	}
	ws.Unlock()

	if err != nil {
		err = fmt.Errorf("%w while closing the http server in gin/webServer", err)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ws *webServer) IsInterfaceNil() bool {
	return ws == nil
}
