package public

import (
	"fmt"
	"os"
	"time"

	"github.com/CMSgov/bcda-ssas-app/ssas"
	"github.com/CMSgov/bcda-ssas-app/ssas/constants"
	"github.com/CMSgov/bcda-ssas-app/ssas/service"
	"github.com/go-chi/chi"
)

var infoMap map[string][]string
var publicSigningKeyPath string
var publicSigningKey string
var clientAssertAud string

var server *service.Server

func init() {
	infoMap = make(map[string][]string)
	publicSigningKeyPath = os.Getenv("SSAS_PUBLIC_SIGNING_KEY_PATH")
	publicSigningKey = os.Getenv("SSAS_PUBLIC_SIGNING_KEY")
	ssas.Logger.Info("public signing key sourced from ", publicSigningKeyPath)
	clientAssertAud = os.Getenv("SSAS_CLIENT_ASSERTION_AUD")
	ssas.Logger.Info("aud value required in client assertion tokens:", clientAssertAud)
}

func Server() *service.Server {
	unsafeMode := os.Getenv("HTTP_ONLY") == "true"
	useMTLS := os.Getenv("PUBLIC_USE_MTLS") == "true"

	signingKey, err := service.ChooseSigningKey(publicSigningKeyPath, publicSigningKey)
	if err != nil {
		msg := fmt.Sprintf("Unable to get public server signing key: %v", err)
		ssas.Logger.Error(msg)
		return nil
	}

	server = service.NewServer("public", ":3003", constants.Version, infoMap, routes(), unsafeMode, useMTLS, signingKey, 20*time.Minute, clientAssertAud)
	if server != nil {
		r, _ := server.ListRoutes()
		infoMap["banner"] = []string{fmt.Sprintf("%s server running on port %s", "public", ":3003")}
		infoMap["routes"] = r
	}
	return server
}

func routes() *chi.Mux {
	router := chi.NewRouter()
	//v1 Routes
	router.Use(service.NewAPILogger(), service.ConnectionClose)
	router.Post("/token", token)
	router.Post("/introspect", introspect)
	router.Post("/authn", VerifyPassword)
	router.With(parseToken, requireMFATokenAuth).Post("/authn/challenge", RequestMultifactorChallenge)
	router.With(parseToken, requireMFATokenAuth).Post("/authn/verify", VerifyMultifactorResponse)
	router.With(parseToken, requireRegTokenAuth, readGroupID).Post("/register", RegisterSystem)
	router.With(parseToken, requireRegTokenAuth, readGroupID).Post("/reset", ResetSecret)

	//v2 Routes
	router.Post("/v2/token", tokenV2)
	router.Post("/v2/token_info", validateAndParseToken)

	return router
}
