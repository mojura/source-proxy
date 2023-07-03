FROM golang:1.20
WORKDIR /home/deploy
COPY ./include/autocert.toml ./include/autocert.toml
COPY ./include/env.toml ./include/env.toml
COPY ./routes ./routes 
COPY ./config.toml ./config.toml
COPY ./groups.json ./groups.json
RUN go install github.com/mojura/source-proxy@v0.1.2
EXPOSE 80
EXPOSE 443
ENTRYPOINT ["source-proxy"]
