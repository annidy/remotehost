# routehost

## build
```
go build .
cp remotehost /usr/local/bin/
```

## sudo without password
```
sudo visudo
```

add
```
username ALL= NOPASSWD: /usr/local/bin/remotehost
```