TAG=${TAG:-v0.2.1}
docker build -t miko/dmt .
docker tag miko/dmt miko/dmt:${TAG}

