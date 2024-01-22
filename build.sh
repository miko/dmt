TAG=${TAG:-v0.2.27}
docker build --build-arg TAG=${TAG} -t miko/dmt .
docker tag miko/dmt miko/dmt:${TAG}
echo docker push miko/dmt:${TAG}
echo docker push miko/dmt:latest
docker push miko/dmt:${TAG}
docker push miko/dmt:latest

