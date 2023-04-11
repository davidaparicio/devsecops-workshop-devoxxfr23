# Workshop Devoxx Petstore API

The Petstore API implement the petstore.json specification with some conformance and security issues. You will be able to find them and to protect the API with 42Crunch tools (Audit/Scan/Firewall). There are more documentations about the Petstore API in the directory 'ressources'


## Prerequisites

- Golang 1.18 or higher (https://go.dev/doc/install)
- An IDE
    - VSCode : (https://code.visualstudio.com/download)
    - IntelliJ : (https://www.jetbrains.com/help/idea/installation-guide.html)
- Docker (https://docs.docker.com/engine/install/)
- Docker-Compose 
- Postman (Recommended to test the API)

## Building and running the API

To build and run the API execute the following commands : 

```
cd ./api
go build
./api
```

When running the last command you should see the following : 

```
   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v4.8.0
High performance, minimalist Go web framework
https://echo.labstack.com
____________________________________O/_______
                                    O\
⇨ http server started on [::]:4010

```

The API is up and is listening on the port 4010. To test that everything is fine you can execute the command and get the response below: 

```
~/devsecops-workshop-devoxxfr23/petstore/api$ curl 'http://petstore:4010/version'
{"commitId":"3271bd2a72002b6a730d6da541c5da355390c4ff","version":"v1.0.0"}

```


## Vizualize the specification (OpenAPI)

There are 3 differents ways to visualize the specification file.

### Binary

Download the binary in the following link :

https://github.com/swagger-api/swagger-ui/releases/latest

### UI 

Open the browser, copy the following link and drag and drop the file './petstore.json': 

https://github.com/swagger-api/swagger-ui/releases/latest


### Docker 

Execute the following command to have an instance of Swagger UI listening on port 8080

```
docker run --rm -p 8080:8080 -e SWAGGER_JSON=/petstore/petstore.json -v $(pwd):/petstore swaggerapi/swagger-ui
```

## 42Crunch Tools 

To be able to execute 42Crunch tools you must have an account in the platform. You can register here : (https://platform.42crunch.com/).

## Static Analysis (Audit)

There are two ways to run the Audit: 

- Using the 42Crunch UI Platform
    - Go to the https://platform.42crunch.com/ page and login
    - Create a collection in the API Collection page
    - Import an API inside the collection
    - Read the Audit report in the 'Security Audit Report' tab
    - Update the specification and re-run the audit in the 'Security editor tag'
    - More documentation : https://42crunch.com/tutorial-api-security-audit/
- Using the 42Crunch IDE Extension 
    - Open your IDE VSCode/IntelliJ
    - Install the 42Crunch extension
        - VSCode (https://marketplace.visualstudio.com/items?itemName=42Crunch.vscode-openapi)
        - IntelliJ (https://plugins.jetbrains.com/plugin/14837-openapi-swagger-editor)
    - Create an IDE token in the page https://platform.42crunch.com/
    - Setup the extension in the IDE 
        - VSCode : Ctrl + Shift + P : Update platform credentials
    - Open an OpenAPI Specification file and click on the top right 42Crunch logo

## Dynamic Analysis (Scan)

There are two ways to run the scan: 

- Using the 42Crunch UI Platform (Old scan engine) 
    - Go to the API page created in the previous section
    - Generate a configuration in the 'On Premises Scan Report' tabs
        - Setup the authentication to reflect API credentials (Hardcoded sessions/tokens are defined in the ./api/api.go file at line 67-84)
    - Execute the following command with the token set
    - Read the Scan report in the Scan report page 
    ```
    docker pull 42crunch/scand-agent:latest
    docker run -e SCAN_TOKEN=<token> 42crunch/scand-agent:latest
    ```

- Using the 42Crunch IDE Extension (New scan engine) 
   - Open VSCode/IntelliJ with the 42Crunch extension
   - Hover an operation and click on 'Scan', it opens a card in a new tab
   - Update the body and the authentication (Hardcoded sessions/tokens are defined in the ./api/api.go file at line 67-84)
   - Click on 'Scan'


## Protection the API (Firewall)

To protect the API, follow the readme file located in the 'fw' directory


## API Issues

The API has some implementation issues describes below. The authorization issues can be enable/disable by using the flag --insecure 

```

./api # Secure mode
./api --insecure # Insecure mode

```

The Petstore API is a testing API for the scan which has the following issues in insecure mode (--insecure) : 

### Authorization 
- CreateUser does not verify that the user is an admin
- CreatePetstore does not verify that the user is an admin
- TransferPet does not verify that the user calling the method is the owner of the Pet 
- ReadPet does not verify that the user calling the method is the owner of the Pet 

### Performance
- The response time of CreatePetstore increase linearly after a limit (100 by default)
- The CreatePet returns a 500 http status code after a limit (100 by default) 

### OpenAPI

- In the request of the operation 'CreateUser' the property 'isAdmin' is not required (default "false")
- In the response of the operation 'Version' the property 'version' is a string

To detect more OpenAPI issues it is easier to use the petstoreInvalid.json (in './scan/petstoreInvalid.json') file which change the contract to not follow the implementation. The issues present in the file which are not respected by the API are as follows : 
- Username max length is 99 instead of 100
- Petname max length is  9 instead of 10
- Delete Pet response 200 require a required header
- Delete Pet response 400 require a required header
