TAG=${TAG:-v0.1.23}
docker build -t miko/dmt .
docker tag miko/dmt miko/dmt:${TAG}

