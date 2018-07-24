## Host-Changer for Windows 10
## Prerequisite
- hc.config.yml file 
- Setting environment value HostChangerPath = {PATH_ON_WHICH_CONFIG_FILE_RESIDE}
- Run Cmd as Administrator (Don't forget it)

## Usage
host-changer switch --env [local | dev | test(default) | pre | live ] --target [host(default) | group] --list test1,test2

## Example
### hc.config.yml (same as the file on repository)
```yml
    envRule:
    dev: 
        - 10.1.55.
        - 10.2.56.
    test: 
        - 22.56.11.
        - 23.
    pre: 
        - 192.168.244.
        - 172.30.200.

group:
    group1:
        - ${group2}
        - test.host.com
        - mouse.wiki
        - www.changer.com

    group2:
        - test2.host.com
        - mouse2.wiki
        - www.changer2.com

address:
  test.host.com:
    - 10.1.55.22
    - 23.2.2.2
    - 172.30.200.2
  mouse.wiki:
    - 22.56.11.2
    - 192.168.244.10
  www.changer.com:
    - 192.168.244.3
  test2.host.com:
    - 10.2.56.12
    - 23.99.2.2
  mouse2.wiki:
    - 10.1.55.12
    - 172.30.200.200
  www.changer2.com:
    - 10.2.56.80
    - 172.30.200.90
```

- hc switch --env live --target group --list group1
> Result : 
>    empty content (because 'env' is live)

- hc switch --env local --target group --list group1
> Result :
>    127.0.0.1 test2.host.com
>    127.0.0.1 mouse2.wiki
>    127.0.0.1 www.changer2.com
>    127.0.0.1 test.host.com
>    127.0.0.1 mouse.wiki
>    127.0.0.1 www.changer.com

- hc switch --env test --list mouse.wiki
> Result :
>    22.56.11.2 mouse.wiki 

- hc switch -t group -l group2
> Result :
>    (Empty, because no host has test range ip address)

