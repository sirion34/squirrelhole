FROM golang:latest 

RUN mkdir /app 

ADD . /app/ 

WORKDIR /app 

RUN go build -o squirrelhole . 

CMD ["/app/squirrelhole"]