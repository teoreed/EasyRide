FROM golang:1.15
LABEL maintainer=670041240
RUN mkdir /app
ADD . /app/
WORKDIR /app


RUN go get github.com/dgrijalva/jwt-go
RUN go get github.com/gorilla/mux


EXPOSE 3333


CMD [ "go", "run", "/app/token.go" ]

# This dockerfile will be run by docker-compose, so does not require a command to build.