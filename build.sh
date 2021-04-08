TAG=${TAG:-v0.1.6}
docker build -t miko/dmt .
docker tag miko/dmt miko/dmt:${TAG}

