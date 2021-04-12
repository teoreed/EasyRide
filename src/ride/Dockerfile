FROM golang:1.15
LABEL maintainer=670041240
RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN go get github.com/gorilla/mux


EXPOSE 4444

CMD [ "go", "run", "/app/ride.go"  ]

# This dockerfile will be run by docker-compose, so does not require a command to build.