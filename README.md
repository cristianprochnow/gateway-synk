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

## Get list of Templates

> `GET` /templates

### GET Params

```
template_id=1&include_content=1
```

* `template_id`: ID do Template desejado, para realizar uma consulta direta
* `include_content = '1'`: para trazer o valor do campo `template.template_content` na listagem.

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
            "template_name": "Marketing Announcement",
            "template_content": "Join our webinar next week on {topic}! #Webinar #{tag}",
            "template_url_import": "",
            "created_at": "25/09/2025 21:19:06"
        }
    ]
}
```

## Create a Template

> `POST` /templates

### Request

```json
{
    "template_name": "template brabo demais",
    "template_content": "template brabo demais",
    "template_url_import": "template brabo demais"
}
```

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "template": {
        "template_id": 3
    }
}
```

## Update a Template

> `PUT` /templates

### Request

```json
{
    "template_id": 1,
    "template_name": "template brabo demais toppp",
    "template_content": "template brabo demais toppp",
    "template_url_import": "template brabo demais toppp"
}
```

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "template": {
        "rows_affected": 1
    }
}
```

## Delete a Template

> `DELETE` /templates

### Request

```json
{
    "template_id": 3
}
```

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "template": {
        "rows_affected": 1
    }
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

## Get list of Integration Profiles

> `GET` /int_profiles

### GET Params

```
int_profile_id=1
```

* `int_profile_id`: ID do Integration Profile desejado, para realizar uma consulta direta

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
			"color_id": 1,
			"color_name": "Primary Blue",
			"color_hex": "007BFF",
			"created_at": "25/09/2025 21:19:06",
			"credentials": [
				{
					"int_credential_id": 1,
					"int_credential_name": "Alice Twitter Account",
					"int_credential_type": "twitter"
				},
				{
					"int_credential_id": 3,
					"int_credential_name": "Alice LinkedIn Account",
					"int_credential_type": "linkedin"
				}
			]
		},
		{
			"int_profile_id": 2,
			"int_profile_name": "Bob Tech Profiles",
			"color_id": 2,
			"color_name": "Success Green",
			"color_hex": "28A745",
			"created_at": "25/09/2025 21:19:06",
			"credentials": [
				{
					"int_credential_id": 2,
					"int_credential_name": "Bob LinkedIn Account",
					"int_credential_type": "linkedin"
				}
			]
		},
		{
			"int_profile_id": 4,
			"int_profile_name": "Integração topzera",
			"color_id": 2,
			"color_name": "Success Green",
			"color_hex": "28A745",
			"created_at": "28/10/2025 00:58:31",
			"credentials": [
				{
					"int_credential_id": 3,
					"int_credential_name": "Alice LinkedIn Account",
					"int_credential_type": "linkedin"
				}
			]
		}
	]
}
```

## Create a Integration Profile

> `POST` /int_profiles

### Request

```json
{
	"int_profile_name": "Novo Perfil de Integração",
	"color_id": 1,
	"credentials": [1, 2]
}
```

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "int_profile": {
        "int_profile_id": 3
    }
}
```

## Update a Integration Profile

> `PUT` /int_profiles

### Request

```json
{
	"int_profile_id": 4,
	"int_profile_name": "Integração topzera",
	"color_id": 2,
	"credentials": [3]
}
```

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "int_profile": {
        "rows_affected": 1
    }
}
```

## Delete a Integration Profile

> `DELETE` /int_profiles

### Request

```json
{
    "int_profile_id": 3
}
```

### Response

```json
{
    "resource": {
        "ok": true,
        "error": ""
    },
    "int_profile": {
        "rows_affected": 1
    }
}
```

## Get list of Integration Credentials for dropdowns

> `GET` /int_credentials/basic

### Response

```json
{
	"resource": {
		"ok": true,
		"error": ""
	},
	"int_credentials": [
		{
			"int_credential_id": 3,
			"int_credential_name": "Alice LinkedIn Account",
			"int_credential_type": "linkedin"
		},
		{
			"int_credential_id": 1,
			"int_credential_name": "Alice Twitter Account",
			"int_credential_type": "twitter"
		},
		{
			"int_credential_id": 2,
			"int_credential_name": "Bob LinkedIn Account",
			"int_credential_type": "linkedin"
		}
	]
}
```

## Get list of Integration Credentials

> `GET` /int_credentials

### GET Params

```
int_credential_id=1&include_config=1
```

* `int_credential_id`: ID do Integration Credential desejado para realizar uma consulta direta
* `include_config`: flag para trazer ou não o conteúdo do campo de `int_credential_config`

### Response

```json
{
	"resource": {
		"ok": true,
		"error": ""
	},
	"int_credentials": [
		{
			"int_credential_id": "1",
			"int_credential_name": "Alice Twitter Account",
			"int_credential_type": "twitter",
			"int_credential_config": "{\"apiKey\": \"key123\", \"apiSecret\": \"secret123\", \"accessToken\": \"token123\"}",
			"created_at": "25/09/2025 21:19:06"
		}
	]
}
```

## Create a Integration Credential

> `POST` /int_credentials

### Request

```json
{
	"int_credential_name": "Linkedinho",
	"int_credential_type": "linkedin",
	"int_credential_config": "{}"
}
```

### Response

```json
{
	"resource": {
		"ok": true,
		"error": ""
	},
	"int_credential": {
		"int_credential_id": 5
	}
}
```

## Update a Integration Credential

> `PUT` /int_credentials

### Request

```json
{
	"int_credential_id": 4,
	"int_credential_name": "Linkedinho",
	"int_credential_type": "twitter",
	"int_credential_config": "{}"
}
```

### Response

```json
{
	"resource": {
		"ok": true,
		"error": ""
	},
	"int_credential": {
		"rows_affected": 1
	}
}
```

## Delete a Integration Credential

> `DELETE` /int_credentials

### Request

```json
{
	"int_credential_id": 5
}
```

### Response

```json
{
	"resource": {
		"ok": true,
		"error": ""
	},
	"int_credential": {
		"rows_affected": 1
	}
}
```