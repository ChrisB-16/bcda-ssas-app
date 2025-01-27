package public

import (
	"fmt"
	"github.com/CMSgov/bcda-ssas-app/ssas"
	"github.com/CMSgov/bcda-ssas-app/ssas/cfg"
	"github.com/CMSgov/bcda-ssas-app/ssas/service"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

var selfRegistrationTokenDuration time.Duration

func init() {
	minutes := cfg.GetEnvInt("SSAS_MFA_TOKEN_TIMEOUT_MINUTES", 60)
	selfRegistrationTokenDuration = time.Duration(int64(time.Minute) * int64(minutes))
}

// MintMFAToken generates a tokenstring for MFA endpoints
func MintMFAToken(oktaID string) (*jwt.Token, string, error) {
	claims := service.CommonClaims{
		TokenType: "MFAToken",
		OktaID:    oktaID,
	}

	if err := checkTokenClaims(&claims); err != nil {
		return nil, "", err
	}

	return server.MintTokenWithDuration(&claims, selfRegistrationTokenDuration)
}

// MintRegistrationToken generates a tokenstring for system self-registration endpoints
func MintRegistrationToken(oktaID string, groupIDs []string) (*jwt.Token, string, error) {
	claims := service.CommonClaims{
		TokenType: "RegistrationToken",
		OktaID:    oktaID,
		GroupIDs:  groupIDs,
	}

	if err := checkTokenClaims(&claims); err != nil {
		return nil, "", err
	}

	return server.MintTokenWithDuration(&claims, selfRegistrationTokenDuration)
}

// MintAccessToken generates a tokenstring that expires in server.tokenTTL time
func MintAccessToken(systemID, clientID string, data string, systemXData string) (*jwt.Token, string, error) {
	claims := service.CommonClaims{
		TokenType:   "AccessToken",
		SystemID:    systemID,
		ClientID:    clientID,
		Data:        data,
		SystemXData: systemXData,
	}

	if err := checkTokenClaims(&claims); err != nil {
		return nil, "", err
	}

	return server.MintToken(&claims)
}

func empty(arr []string) bool {
	empty := true
	for _, item := range arr {
		if item != "" {
			empty = false
			break
		}
	}
	return empty
}

func tokenValidity(tokenString string, requiredTokenType string) error {
	tknEvent := ssas.Event{Op: "tokenValidity"}
	ssas.OperationStarted(tknEvent)
	t, err := server.VerifyToken(tokenString)
	if err != nil {
		tknEvent.Help = err.Error()
		ssas.OperationFailed(tknEvent)
		return err
	}

	c := t.Claims.(*service.CommonClaims)

	err = checkAllClaims(c, requiredTokenType)
	if err != nil {
		tknEvent.Help = err.Error()
		ssas.OperationFailed(tknEvent)
		return err
	}

	err = c.Valid()
	if err != nil {
		tknEvent.Help = err.Error()
		ssas.OperationFailed(tknEvent)
		return err
	}

	if service.TokenBlacklist.IsTokenBlacklisted(c.Id) {
		err = fmt.Errorf("token has been revoked")
		tknEvent.Help = err.Error()
		ssas.OperationFailed(tknEvent)
		return err
	}

	ssas.OperationSucceeded(tknEvent)
	return nil
}

func checkAllClaims(claims *service.CommonClaims, requiredTokenType string) error {
	if err := server.CheckRequiredClaims(claims, requiredTokenType); err != nil {
		return err
	}

	if err := checkTokenClaims(claims); err != nil {
		return err
	}
	return nil
}

func checkTokenClaims(claims *service.CommonClaims) error {
	switch claims.TokenType {
	case "MFAToken":
		if claims.OktaID == "" {
			return fmt.Errorf("MFA token must have OktaID claim")
		}
	case "RegistrationToken":
		if empty(claims.GroupIDs) {
			return fmt.Errorf("registration token must have GroupIDs claim")
		}
	case "AccessToken":
		if claims.Data == "" {
			return fmt.Errorf("access token must have Data claim")
		}
	default:
		return fmt.Errorf("missing token type claim")
	}
	return nil
}
