TAG=${TAG:-v0.1.21}
docker build -t miko/dmt .
docker tag miko/dmt miko/dmt:${TAG}

