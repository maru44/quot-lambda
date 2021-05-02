FROM golang:1.16.3-alpine
RUN apk update apk add git
#    && go get -u github.com/aws/aws-lambda-go/lambda \
#    && go get -u github.com/aws/aws-lambda-go \
#    && go get -u github.com/aws/aws-sdk-go

ENV ROOT=/go/src/app
WORKDIR ${ROOT}

COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

EXPOSE 8080

CMD ["go", "run", "main.go"]
