FROM golang
ADD . /home
WORKDIR /home
RUN go get gopkg.in/yaml.v2
RUN go get github.com/dgrijalva/jwt-go
RUN go build main.go
CMD ./main
