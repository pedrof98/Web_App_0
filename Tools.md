# Toolkit

For this project I will try to make first an iteration using Python and afterwards the same project using Go.

I will start with the tools I will use for the Python version and then Go.

# Python version

## Web Framework

### FastAPI

* Modern, async-first, and built on top of Starlette and Pydantic.
* Auto-generates OpenAPI/Swagger docs (very helpful for your project).
* Excellent performance (for Python) and developer experience.

## Database

### PostgreSQL

* Rich Feature Set: JSONB support, window functions, CTEs, geospatial (PostGIS), time-series extensions (Timesclae).
* Robust & Scalable
* Widely supported by Python ORMs (SQLAlchemy, Django ORM, etc)


## ORM (Object-Relational Mapping)

### SQLAlchemy

* The standard for Python when using relational databases.
* Can be used for both Flask and FastAPI.
* It offers:
    - Powerful query building (Core or ORM style)
    - Good support for relationships
    - Excellent documentation


### Migrations: Alembic

* Alembic is closely tied to SQLAlchemy.
* It can be used to define models in Python (ORM) or manually write migration scripts.
* Commands like ```alembic revision --autogenerate``` create migrations from model changes, and ```alembig upgrade head```applies them to the DB.

### Typical Alembic Workflow

1. Install: ```pip install alembic```
2. Initialize: ```alembic init migrations```
3. Configure ```alembic.ini``` with your DB URL (e.g., postgresql://user:pass@db:5432/dbname)
4. Autogenerate a migration from your SQLAlchemy models.
5. Upgrade or downgrade your database version as needed.

## The complete stack

### FastAPI + SQLAlchemy + Alembic

* A common combination of this stack might look like:
    - ```uvicorn```as the server.
    - ```FastAPI``` for the routing and OpenAPI generation
    - ```SQLAlchemy``` for the data models and queries
    - ```Alembic``` for the schema migrations
    - ```pytest``` for testing
    - Docker Compose to spin up both the API container and PostgreSQL container


# Go version

## Web Framework

### Gin

* Simple router, straighforward context handling, good performance.
* Easy to integrate with middlewares (auth, logging, etc).


## ORM (Object-Relational Mapping)

### GORM 

* Common choice, even though some devs prefer writing raw SQL or use a lightweight query builder due to the verbosity of GORM (i'm using it anyways since the purpose of the project is to learn, not optimize- for now)
* It can handle relationships, migrations, hooks, etc.


## Migrations

### Goose

### Typical Workflow with Goose

1. Install Goose: ```go install github.com/pressly/goose/v3/cmd/goose@latest```
2. Initialize a migration folder: ```goose -dir migrations create init sql```
3. Write your ```up```and ```down``` statements in ```.sql``` files
4. Run migrations: ```goose -dir migrations postgres "user=... dbname=..." up

### Note on API documentation using Go

* You can manually write OpenAPI specs or use tools like swag (github.com/swaggo/swag) to auto-generate docs from your Gin handlers.