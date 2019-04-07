# Admin api

To generate the admin api stubs run the following command from the root directory `docker run --rm -v $(pwd):$(pwd) -w $(pwd) znly/protoc -I . -I vendor pkg/adminapi/adminapi.proto --go_out=plugins=grpc:.`

WARNING: make sure you have first initialized the project by running `govendor sync` since the admin api protobuf file has a dependency.

NOTE: govendor sometimes doesn't want to update the abuse-mesh-protocol dependency because it contains no go code. If this happens the solution is to clear the govendor cache and then retrying to fetch the new version.