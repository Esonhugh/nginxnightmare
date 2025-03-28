

build:
	GOOS=linux GOARCH=amd64 go build -o ingressnightmare ./
	upx ingressnightmare
