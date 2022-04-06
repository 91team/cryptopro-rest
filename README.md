## Build

`docker build -t registry.91.team/cryptopro/rest:latest .`

## Run

```
docker run --rm -it -p "3001:3000" -v "$(pwd)/keys:/var/opt/cprocsp/keys/root" \
    -e "KEY_PASSWORD=12345678" \
    -e "KEY_THUMBPRINT=05067bc2dafbc8fdd0e729567bbdfb140096281f" \
    -e "LICENSE_KEY=40406-A0000-0219M-Q778D-1Y222" \
    -e "API_KEY=c042ee2fa0f5bd5a3bceeae6f5cd8de066d6d9b9fd7" \
    registry.91.team/cryptopro/rest:latest
```

- KEY_THUMBPRINT can be obtained from `certmgr -list` command output("SHA1 Hash" row)

## Push to registry

`docker push registry.91.team/cryptopro/rest:latest`

## Usage

`curl -H 'Authorization: Bearer c042ee2fa0f5bd5a3bceeae6f5cd8de066d6d9b9fd7' -d 'data for signing' -X POST http://localhost:3001/api/sign`

## Kubernetes deployment

`kubectl create ns csp`
`kubectl apply -f secret.yaml -n csp`
`kubectl apply -f deployment.yaml -n csp`
