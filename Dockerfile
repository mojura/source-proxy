FROM golang:1.20
WORKDIR /home/deploy
COPY ./include/env.toml ./include/env.toml
COPY ./routes ./routes 
COPY ./config.toml ./config.toml
COPY ./apikeys.json ./apikeys.json
COPY ./resources.json ./resources.json
RUN go install github.com/mojura/source-proxy@v0.2.4
EXPOSE 80
EXPOSE 443
ENTRYPOINT ["source-proxy"]
