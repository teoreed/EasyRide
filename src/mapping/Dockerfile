FROM golang:1.15
LABEL maintainer=670041240
RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN go get github.com/gorilla/mux
RUN go get googlemaps.github.io/maps

CMD [ "go", "run", "/app/mapping.go" ]

# This dockerfile will be run by docker-compose, so does not require a command to build.