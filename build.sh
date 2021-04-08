TAG=${TAG:-v0.1.7}
docker build -t miko/dmt .
docker tag miko/dmt miko/dmt:${TAG}

