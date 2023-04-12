build:
	echo "Build petstore.gen files"
	oapi-codegen -package main -generate server ../petstore.json > petstore.gen.server.go
	oapi-codegen -package main -generate types ../petstore.json > petstore.gen.types.go
	oapi-codegen -package main -generate spec ../petstore.json > petstore.gen.spec.go
	go build