c-jobs
======================

# Technologies
List of technologies that I chose to work:
* golang - performance, low memory consumption, fast, fun, opensource, easy to deploy, clean and much more =)
* elasticsearch - the problem was text searching, so elasticsearch was a good option
* goji - request multiplexer
* docker/docker-compose - container, helps to guarantees the environment creation and isolates the development
* govendor - simple go tool for vendor control

# Improvements needed
* process asynchronous jobs received on 'Add jobs', send in bulks using more than one goroutine
* authentication on 'Add jobs'
* 'Delete jobs' service
* custom configuration for elasticsearch docker
* configure docker to be able to use golang elastic client's sniff (https://github.com/olivere/elastic/wiki/Docker)
* tests for http server
* ...

# Build
run tests and compile

```sh
$ ./build.sh linux
```
```sh
$ ./build.sh darwin
```

## Tests
```sh
$ docker run -v "$(pwd)":/gopath/src/github.com/bvieira/c-jobs -e "GOPATH=/gopath" -w /gopath/src/github.com/bvieira/c-jobs golang:latest sh -c "./test-coverage.sh"
```

## Environment Variables
list all variables available

```sh
$ docker run -v "$(pwd)":/gopath/src/github.com/bvieira/c-jobs -e "GOPATH=/gopath" -w /gopath/src/github.com/bvieira/c-jobs/jobsserver/ golang:latest sh -c "go install; /gopath/bin/jobsserver -env"
```

# Run
```sh
$ ./start.sh
```

# Stop
```sh
$ docker-compose stop
```

# Logs
```sh
$ docker-compose logs -f
```

# API
- [Add jobs](#add-jobs)
- [Search jobs](#search-jobs)

## Error handling
if something went wrong on request, the application should return http code different from 2xx and on body the [Error response](#error-response)

| code   | description           |
|-------------------|-----------------------|
| `JOB0000`         | unknown  error |
| `JOB1001`         | invalid request error |
| `JOB1002`         | not found error |
| `JOB1003`         | parser error  |
| `JOB2001`         | elastic search connect error  |
| `JOB2002`         | elastic search access error  |


## Add jobs
index jobs on repository, create if ID do not exists, updates otherwise

### Request:
`POST` /jobs

#### Body:
- [Jobs Request](#jobs-request)


### Response:
| code   | description           | body content |
|-------------------|-----------------------|-------|
| 204             | success  |  |
| 400             | invalid request  | [Error response](#error-response) |
| 500             | error accessing elasticsearch  | [Error response](#error-response) |

### Example:
```sh
$ curl -v -H "Content-Type: application/json" -X POST localhost:8080/jobs -d '{"docs":[{"title":"Analista de TI","description":"<li> Conhecimento aprofundado em Linux Server (IPTables, proxy, mail, samba) e Windows Server(MS-AD, WTS, compartilhamentos).</li>","salario":3200.5,"cidade":["Joinville"],"cidadeFormated":["Joinville - SC (1)"]}]}'
> POST /jobs HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.43.0
> Accept: */*
> Content-Type: application/json
> Content-Length: 264
>
< HTTP/1.1 204 No Content
< Date: Thu, 26 Jan 2017 02:02:39 GMT
<
```

```sh
$ curl -v -H "Content-Type: application/json" -X POST localhost:8080/jobs -d @vagas.json
> POST /jobs HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.43.0
> Accept: */*
> Content-Type: application/json
> Content-Length: 686934
> Expect: 100-continue
>
< HTTP/1.1 100 Continue
< HTTP/1.1 204 No Content
< Date: Thu, 26 Jan 2017 02:08:43 GMT
<
```


## Search jobs
search jobs according with query and sort options

### Request:
`GET` /jobs?content=:content&city=:city&sort=:sort

obs: either content and city are not required, but at least one should be defined

| param   |          required | description           |
|-------------------|-------|-----------------------|
| `:content`          | no |  `content` for searching on 'title' and 'description'. '*' wildcard can be used on the right of the content. if content contains space, the result must contain each word. For exact search, use '"' (double quote), for more info, look on https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl-simple-query-string-query.html|
| `:city`             | no |  `city` for searching on 'cidade'. use the same rules defined on `content` param |
| `:sort`             | no |  `sort` for sorting, use 'asc' or 'desc' for order. default: desc  |


### Response:
| code   | description           | body content |
|-------------------|-----------------------|-------|
| 200             | success  | [Job Result Response](#jobs-search-response) |
| 400             | invalid request  | [Error response](#error-response) |
| 500             | error accessing elasticsearch  | [Error response](#error-response) |

### Example:
```sh
$ curl -v "http://localhost:8080/jobs?content=analista&sort=asc"
> GET /jobs?content=analista&sort=asc HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< Date: Thu, 26 Jan 2017 02:04:51 GMT
< Content-Length: 256
<
[{"title":"Analista de TI","description":"<li> Conhecimento aprofundado em Linux Server (IPTables, proxy, mail, samba) e Windows Server(MS-AD, WTS, compartilhamentos).</li>","salario":3200.5,"cidade":["Joinville"],"cidadeFormated":["Joinville - SC (1)"]}] 
```

# Schema
## Jobs Request

Job ID: composition of 'title', 'salario' and 'cidade'. Each field is normalized, any accent or symbol is removed.

| header   | value           |
|-------------------|-----------------------|
| `Content-Type`             | application/json  |

	{
        "docs": [
            {
                "title": string,
                "description": string,
                "salario": floating-point number,
                "cidade": string[],
                "cidadeFormated": string[]
            }
        ]
    }
	
eg.

	{
        "docs": [{
            "title": "Analista de Suporte de TI",
            "description": "<li> Prestar atendimento remoto e presencial a clientes. Atuar com suporte de TI.</li><li> Conhecimento aprofundado em Linux Server (IPTables, proxy, mail, samba) e Windows Server(MS-AD, WTS, compartilhamentos).</li>",
            "salario": 3200,
            "cidade": [
                "Joinville"
            ],
            "cidadeFormated": [
                "Joinville - SC (1)"
            ]
        }]
    }


## Jobs Search Response

| header   | value           |
|-------------------|-----------------------|
| `Content-Type`             | application/json  |

	[
		{
			"title": string,
			"description": string,
			"salario": floating-point number,
			"cidade": string[],
			"cidadeFormated": string[]
		}
	]
	
eg.

	[
        {
            "title": "Estagio de Auxiliar Fiscal",
            "description": "<li> Deverá estar cursando: Ensino Superior ou Técnico em Contabilidade.</li><li> Auxiliar nas rotinas do departamento, tais como arquivamento de documentações, lançamento de dados nos sistemas, identificação de pastas e caixas, abrir malotes de documentos.</li>",
            "salario": 1000,
            "cidade": [
                "Blumenau"
            ],
            "cidadeFormated": [
                "Blumenau - SC (1)"
            ]
        }
    ]


## Error response

| header   | value           |
|-------------------|-----------------------|
| `Content-Type`             | application/json  |

	{
        "error": string,
		"message": string
	}
	
eg.

	{
        "error":"JOB1001",
        "message":"could not parse body content, error: EOF"
	}
