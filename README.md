# remotehost

## build
```
go build .
go install .
```

## example
demos
```
sudo remotehost -r -i 30 -u https://gitee.com/if-the-wind/github-hosts/raw/main/hosts -n 'GitHub Host'
```

cron job
```
*/30 * * * * remotehost -v -u https://gitee.com/if-the-wind/github-hosts/raw/main/hosts -n 'GitHub Host'
```
