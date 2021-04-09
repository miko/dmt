TAG=${TAG:-v0.1.16}
docker build -t miko/dmt .
docker tag miko/dmt miko/dmt:${TAG}

