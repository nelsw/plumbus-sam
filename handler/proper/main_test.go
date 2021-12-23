package main

import (
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"testing"
)

const body = `{
  "type": "dashboard",
  "scheduled_plan": {
    "scheduled_plan_id": 2324,
    "title": "Acquisition",
    "type": "LookMLDashboard",
    "url": null,
    "query_id": null,
    "query": null,
    "filters_differ_from_look": null,
    "download_url": null
  },
  "attachment": {
    "mimetype": "application/zip;base64",
    "extension": "zip",
    "data": "UEsDBBQAAAgIAHIElFNVsBCbpwAAABEBAAAlAAAAZGFzaGJvYXJkLWFjcXVp\nc2l0aW9uL2ltcHJlc3Npb25zLmNzdl2Nyw6CMBBF93xF0/UNdiht7ZKlO4Px\nAwCLNkZQi/r7FhFiXM2cO4+Dork9fPCD7zsUB7a5XO8uhEhhxH3nh/+sdG3E\nEyurwWFbHd3Tu1fAbl5ZIrZi3zAhtFXj6r4/g5OB0RZKS/4BRQpCEAelgsBz\nkJGwMh+nkpCL2ElKjU4ytDV4BinXfKnxSkCTjgRap9okmByGYMe3k8PEuf11\n5FBKzA5pZscbUEsDBBQAAAgIAHIElFPAWrADqwAAAAABAAAlAAAAZGFzaGJv\nYXJkLWFjcXVpc2l0aW9uL3BlcmZvcm1hbmNlLmNzdm2NuwqDQBBFe79CZO2G\nQfeppUjKgJhHr2YMSxIluCbk77NR0tncOQP3cKHonrOdrLPjAMUlpLLa/+5p\nsG59quZKYe3hQNPkWwuXxxrOlt5Na+/WfWA3OftoHF3Cml40zBSk0DcdteN4\nA5ZizpeUwDJMFTCuckwSSJDrGLRGZWKImJCQCIFaRQGHvv0raxoUxrMwqLkX\nhfCiROl9JlHJALZ3MszN5o4v5FHwBVBLAwQUAAAICAByBJRT7zwCC0wAAABc\nAAAAHwAAAGRhc2hib2FyZC1hY3F1aXNpdGlvbi9jaGFydC5jc3bTCQ3xVQjO\nLy1KTtVJS0xOTcrPz9ZJS+LSUXBJLEnVcS0uycwFMlIUglLLUvNKsYhwGeoY\nGRgZ6hoa6Rpa6hib6BkYe+uY6JmacAEAUEsBAj8DFAAACAgAcgSUU1WwEJun\nAAAAEQEAACUAAAAAAAAAAQAAAKSBAAAAAGRhc2hib2FyZC1hY3F1aXNpdGlv\nbi9pbXByZXNzaW9ucy5jc3ZQSwECPwMUAAAICAByBJRTwFqwA6sAAAAAAQAA\nJQAAAAAAAAABAAAApIHqAAAAZGFzaGJvYXJkLWFjcXVpc2l0aW9uL3BlcmZv\ncm1hbmNlLmNzdlBLAQI/AxQAAAgIAHIElFPvPAILTAAAAFwAAAAfAAAAAAAA\nAAEAAACkgdgBAABkYXNoYm9hcmQtYWNxdWlzaXRpb24vY2hhcnQuY3N2UEsF\nBgAAAAADAAMA8wAAAGECAAAAAA==\n"
  },
  "data": null,
  "form_params": {}
}`

func TestHandle(t *testing.T) {

	request := events.APIGatewayV2HTTPRequest{
		Body: body,
	}

	response, err := handle(request)
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fail()
	}

}
