GOARCH=amd64 GOOS=linux go build -v -o ./bin/go-retwis
docker build --tag=victornguyenm/go-retwis:latest .

echo "DOCKER_HUB_ACCESS_TOKEN:" $DOCKER_HUB_ACCESS_TOKEN

docker login -u victornguyenm -p $DOCKER_HUB_ACCESS_TOKEN
docker push victornguyenm/go-retwis:latest