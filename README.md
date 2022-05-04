
<h1 align="center">
    <br>
    b(uild c)url
    <br>
    <br>
</h1>

burl builds curl commands. that's about it.

<img src=".github/demo.gif">

<br />

## install

```
go install github.com/philippta/burl@latest
```

## usage

1. run burl
2. fill in request parameters
3. hit enter
4. enjoy the curl command

```
$ burl
curl https://www.example.com/endpoint
  -X POST
  -H Content-Type: application/json
  -d {"foo":"bar"}

<ctrl-h> header | <ctrl-d> data | <ctrl-x> remove | <enter> build

$ curl https://www.example.com/endpoint -X POST -H 'Content-Type: application/json' -d '{"foo":"bar"}'
```

## license

[MIT](/LICENSE)
