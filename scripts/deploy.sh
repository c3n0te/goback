docker build -f ./services/gateway/Containerfile --rm -t gateway:v0.1 .
docker run -d --rm --name gateway gateway:v0.1
