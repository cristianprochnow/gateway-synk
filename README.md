# Gateway
Service to receive all requests and redirect logic to other pieces from system.

# Getting Started

First step to run project, is to run docker command to build database. This database will need config from [SQL file from setup](./_setup/schema.sql) folder.

```shell
docker run --name synk_db -d -ti -p 3307:3306 -e MYSQL_ROOT_PASSWORD=password -v "path\to\db:/var/lib/mysql" mysql:8.0 --sql-mode="TRADITIONAL" --bind-address="0.0.0.0" --default_time_zone="-03:00"
```

So next step is to create a `.env` file in project root and change example values to your config. You can use `example.env` file from `_setup` folder as template.

And then, run `docker compose up -d` into project root to start project.

## Tests

The easy way to run tests is just run `docker compose up -d` command to start project with variables. So, enter in `synk_gateway` with `docker exec` and run `go test ./tests -v`.

# Routes

## Get info about app

> `GET` /about

### Response

```json
{
	"ok": true,
	"error": "",
	"info": {
		"server_port": "8080",
		"app_port": "8083",
		"db_working": true
	},
	"list": null
}
```

## Get list of Posts

> `GET` /post

### GET Params

```
post_id=1&include_content=1
```

* `post_id`: ID do Post desejado, para realizar uma consulta direta
* `include_content = '1'`: para trazer o valor do campo `post.post_content` na listagem.

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "posts": [
        {
            "post_id": 1,
            "post_name": "Post name atualizado",
            "template_name": "Marketing Announcement",
            "int_profile_name": "Alice Marketing Profiles",
            "created_at": "25/09/2025 21:20:37",
            "status": "pending",
            "post_content": "",
            "template_id": 1,
            "int_profile_id": 1
        },
        {
            "post_id": 2,
            "post_name": "Version 2.5 Release",
            "template_name": "Tech Update Post",
            "int_profile_name": "Bob Tech Profiles",
            "created_at": "25/09/2025 21:20:37",
            "status": "failed",
            "post_content": "",
            "template_id": 2,
            "int_profile_id": 2
        }
    ]
}
```

## Create a Post

> `POST` /post

### Request

```json
{
	"post_name": "Post name show",
	"post_content": "conteúdo show",
	"template_id": 1,
	"int_profile_id": 2
}
```

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "post": {
        "post_id": 3
    }
}
```

## Update a Post

> `PUT` /post

### Request

```json
{
    "post_id": 1,
    "post_name": "Post name atualizado",
    "post_content": "conteúdo atualizado",
    "template_id": 1,
    "int_profile_id": 1
}
```

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "post": {
        "rows_affected": 1
    }
}
```

## Delete a Post

> `DELETE` /post

### Request

```json
{
    "post_id": 3
}
```

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "post": {
        "rows_affected": 1
    }
}
```

## Get list of Templates for dropdowns

> `GET` /templates/basic

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "templates": [
        {
            "template_id": 1,
            "template_name": "Marketing Announcement"
        },
        {
            "template_id": 2,
            "template_name": "Tech Update Post"
        }
    ]
}
```

## Get list of Integration Profiles for dropdowns

> `GET` /int_profiles/basic

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "int_profiles": [
        {
            "int_profile_id": 1,
            "int_profile_name": "Alice Marketing Profiles",
            "color_name": "Primary Blue",
            "color_hex": "007BFF"
        },
        {
            "int_profile_id": 2,
            "int_profile_name": "Bob Tech Profiles",
            "color_name": "Success Green",
            "color_hex": "28A745"
        }
    ]
}
```