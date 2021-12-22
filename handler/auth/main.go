package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	log "github.com/sirupsen/logrus"
	"github.com/uris77/auth0"
	"net/http"
	"os"
)

var (
	client = auth0.NewAuth0(60, 518400)
	jwks   = os.Getenv("JWKS_URI")
	aud    = os.Getenv("AUTH_AUDIENCE")
	iss    = os.Getenv("AUTH_ISSUER")
)

func init() {
	if len(jwks) < 1 {
		panic("The required JWKS_URI is missing")
	}
	if len(aud) < 1 {
		panic("The required AUTH_AUDIENCE is missing")
	}
	if len(iss) < 1 {
		panic("The required AUTH_ISSUER is missing")
	}
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func handle(ctx context.Context, evt events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {

	if evt.HTTPMethod == http.MethodOptions {
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: true,
			Context:      map[string]interface{}{"key": "val"},
		}, nil
	}

	log.WithFields(log.Fields{
		"event":   evt,
		"jwks":    jwks,
		"aud":     aud,
		"iss":     iss,
		"client":  client,
		"context": ctx,
	}).Info("Authorizer triggered")

	jwtToken := evt.Headers["authorization"]

	jwt, err := client.Validate(jwks, aud, iss, jwtToken)

	if err != nil {

		r := events.APIGatewayCustomAuthorizerResponse{
			PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
				Version: "2012-10-17",
				Statement: []events.IAMPolicyStatement{
					{
						Action:   []string{"execute-api:Invoke"},
						Effect:   "Deny",
						Resource: []string{evt.MethodArn},
					},
				},
			},
		}

		log.WithFields(log.Fields{
			"jwt":      jwt,
			"err":      err,
			"response": r,
		}).Error("Token Validation Failed")

		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
		}, nil
	}

	resp := events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: jwt.Subject(),
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Allow",
					Resource: []string{evt.MethodArn},
				},
			},
		},
	}

	log.WithFields(log.Fields{
		"jwt":      jwt,
		"response": resp,
	}).Info("Successfully Validated Token")

	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context:      map[string]interface{}{"key": "val"},
	}, nil
}

func main() {
	lambda.Start(handle)
}
