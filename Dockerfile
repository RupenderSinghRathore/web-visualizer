FROM debian:stable-slim 
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* 

COPY ./bin/web-visualizer /bin/web-visualizer

ENTRYPOINT [ "/bin/web-visualizer" ]
