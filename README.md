c-jobs
======================

# Technologies
List of technologies that I chose to work:
* golang - performance, low memory consumption, fast, fun, opensource, easy to deploy, clean and much more =)
* elasticsearch - the problem was text searching, so elasticsearch was a good option
* docker/docker-compose - container, helps to guarantees the environment creation and isolates the development


# Build
```sh
$ 
```

# Run
```sh
$ 
```

# Stop
```sh
$ 
```

# Logs
```sh
$ 
```

# API
- [Add jobs](#add-jobs)
- [Search jobs](#search-jobs)

## Error handling
if something went wrong on request, the application should return http code different from 2xx and on body the [Error response](#error-response)

## Add jobs
index jobs

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
$ 
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
$ 
```

# Schema
## Jobs Request

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
