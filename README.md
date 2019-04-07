# abuse-mesh-go
Abuse mesh daemon and cli written in Go

## How to install

`todo`

## How to run

`todo`

## Contribution guide

`todo`

### Debugging grpc

evans is a universal grpc client which can be useful to debug a grpc server

for admin api `evans internal/adminapi/adminapi.proto --path vendor --path . --port 181 --package adminapi --service admininterface`

for abuse mesh protocol `evans abuse-mesh.proto --path vendor/github.com/abuse-mesh/abuse-mesh-protocol --port 180 --package abusemesh --service AbuseMesh`

## Wish list

- graceful config reloads
- prometheus endpoint
- debug endpoint (pprof, trace)
- flexible logging (stdout, file, syslog and network options)
- web interface (at some point)