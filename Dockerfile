FROM golang
ADD . /home
WORKDIR /home
RUN go get github.com/dgrijalva/jwt-go
RUN go build main.go
CMD ./main
