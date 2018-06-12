FROM golang
ADD . /home
WORKDIR /home
RUN go build main.go
CMD ./main
