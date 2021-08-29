TAG=${TAG:-v0.2.12}
docker build --build-arg TAG=${TAG} -t miko/dmt .
docker tag miko/dmt miko/dmt:${TAG}

