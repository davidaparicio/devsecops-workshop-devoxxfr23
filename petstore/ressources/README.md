# Postman Collection

Documentation :  https://learning.postman.com/docs/collections/running-collections/working-with-data-files/

# Postman Environment 

Managing environment : https://learning.postman.com/docs/sending-requests/managing-environments/

Please update the hosts file with the following data to reflect the environment on your machine : 

```
127.0.0.1    pixi-secured.42crunch.test
127.0.0.1    pixi-open.42crunch.test
```


### Petstore Insecured - API

| Variable | Value                 |
|----------|-----------------------|
| scheme   | http                  |
| hostname | petstore:4010         |

### Petstore Secured - FW

| Variable | Value                 |
|----------|-----------------------|
| scheme   | https                 |
| hostname | petstore-secured:4241 |