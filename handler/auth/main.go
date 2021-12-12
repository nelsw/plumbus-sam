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

var auth0Client auth0.Auth0

func init() {
	// Instantiate an auth0 client with a Cache with the capacity for
	// 60 tokens and a ttl of 24 hours
	auth0Client = auth0.NewAuth0(60, 518400)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
}

func handle(ctx context.Context, evt events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {

	if evt.HTTPMethod == http.MethodOptions {
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: true,
			Context: map[string]interface{}{"key":"val"},
		}, nil
	}

	if len(os.Getenv("JWKS_URI")) < 1 {
		panic("The required JWKS_URI is missing")
	}
	jwkUrl := os.Getenv("JWKS_URI")

	if len(os.Getenv("AUTH_AUDIENCE")) < 1 {
		panic("The required AUTH_AUDIENCE is missing")
	}
	aud := os.Getenv("AUTH_AUDIENCE")

	if len(os.Getenv("AUTH_ISSUER")) < 1 {
		panic("The required AUTH_ISSUER is missing")
	}
	iss := os.Getenv("AUTH_ISSUER")

	log.WithFields(log.Fields{
		"event":       evt,
		"jwkUrl":      jwkUrl,
		"aud":         aud,
		"iss":         iss,
		"auth0Client": auth0Client,
		"context":     ctx,
	}).Info("Authorizer triggered")

	jwtToken := evt.Headers["authorization"]

	jwt, err := auth0Client.Validate(jwkUrl, aud, iss, jwtToken)

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
		Context: map[string]interface{}{"key":"val"},
	}, nil



}

func main() {
	lambda.Start(handle)
}