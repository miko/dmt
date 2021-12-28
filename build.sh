TAG=${TAG:-v0.2.14}
docker build --build-arg TAG=${TAG} -t miko/dmt .
docker tag miko/dmt miko/dmt:${TAG}
echo docker push miko/dmt:${TAG}
echo docker push miko/dmt:latest

