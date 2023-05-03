# Protect the API with the API Firewall

To deploy the API Firewall in front of the Petstore API you have to get a protection token linked to the API in 42Crunch platform. To get this token follow the tutorial (https://docs.42crunch.com/latest/content/tasks/protect_apis.htm
)

When you have the protection token you can execute the command below

```
PROTECTION_TOKEN=<token> docker-compose -p 42Crunch -f protect.yml up fw-petstore-secured
```

It deploys a FW in 'front' of the API, and listen on the port 4241. All requests are redirected to the petstore API after a validation against the OpenAPI specification file.


### Documentation

Step by step deployment with docker compose

https://github.com/42Crunch/resources/blob/master/firewall-deployment/DockerCompose.md
