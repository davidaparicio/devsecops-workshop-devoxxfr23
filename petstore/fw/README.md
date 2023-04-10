# Protect the API with the API Firewall

To deploy the API Firewall in front of the Petstore API, run the following command : 

```

PROTECTION_TOKEN=<token> docker-compose -p 42Crunch -f protect.yml up fw-petstore-secured

```



### Documentation

Create a protection token : 

https://docs.42crunch.com/latest/content/tasks/protect_apis.htm

Step by step deployment with docker compose : 

https://github.com/42Crunch/resources/blob/master/firewall-deployment/DockerCompose.md