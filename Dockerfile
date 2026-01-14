FROM ubuntu:latest
LABEL authors="ima"

ENTRYPOINT ["top", "-b"]