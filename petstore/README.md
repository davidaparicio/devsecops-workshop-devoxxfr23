# Petstore API 

## Building and running the API 

```
cd ./api
go build && ./api
```

## Vizualize OpenAPI

```
docker run --rm -p 8081:8080 -e SWAGGER_JSON=/petstore/petstore.json -v $(pwd):/petstore swaggerapi/swagger-ui
```


## API Issues

The Petstore API is a testing API for the scan which has the following issues in insecure mode (--insecure) : 

## Authorization 
- CreateUser does not verify that the user is an admin
- CreatePetstore does not verify that the user is an admin
- TransferPet does not verify that the user calling the method is the owner of the Pet 
- ReadPet does not verify that the user calling the method is the owner of the Pet 

## Performance
- The response time of CreatePetstore increase linearly after a limit (100 by default)
- The CreatePet returns a 500 http status code after a limit (100 by default) 

## OpenAPI

- CreateUser isAdmin is not required (default "false")


To detect more OpenAPI issues it is easier to use the petstoreInvalid.json file which change the contract to not follow the code. The issues present in the file which are not respected by the API are as follows : 
- Username max length is 99 instead of 100
- Petname max length is  9 instead of 10
- Delete Pet response 200 require a required header
- Delete Pet response 400 require a required header