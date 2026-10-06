# Go Hiring Challenge

This repository contains a Go application for managing products and their prices, including functionalities for CRUD operations and seeding the database with initial data.

## Project Structure

1. **cmd/**: Contains the main application and seed command entry points.

   - `server/main.go`: The main application entry point, serves the REST API.
   - `seed/main.go`: Command to seed the database with initial product data.

2. **app/**: Contains the application logic.
3. **sql/**: Contains a very simple database migration scripts setup.
4. **models/**: Contains the data models and repositories used in the application.
5. `.env`: Environment variables file for configuration.

## Setup Code Repository

1. Create a github/bitbucket/gitlab repository and push all this code as-is.
2. Create a new branch, and provide a pull-request against the main branch with your changes. Instructions to follow.

## Application Setup

- Ensure you have Go installed on your machine.
- Ensure you have Docker installed on your machine.
- Important makefile targets:
  - `make tidy`: will install all dependencies.
  - `make docker-up`: will start the required infrastructure services via docker containers.
  - `make seed`: ⚠️ Will destroy and re-create the database tables.
  - `make test`: Will run the tests.
  - `make run`: Will start the application.
  - `make docker-down`: Will stop the docker containers.

Follow up for the assignemnt here: [ASSIGNMENT.md](ASSIGNMENT.md)

## API

- `GET /catalog?offset=0&limit=10&category=SHOES&price_max=20` - paginated product list (`products`, `total_number`). `limit` is clamped to [1, 100]; products are ordered by id.
- `GET /catalog/{code}` - product details with category and variants (a variant without its own price inherits the product's). `404` if not found.
- `GET /categories` - list all categories.
- `POST /categories` with `{"code": "HATS", "name": "Hats"}` - creates a category (`201`; `400` on invalid body; `409` if the code already exists).

## Design Decisions

- **Consumer-defined interfaces**: each handler package (`app/catalog`, `app/categories`) declares the repository interface it needs, instead of depending on the concrete `models.*Repository` structs. Repositories stay decoupled from the HTTP layer.
- **Repositories are "dumb"**: `models` only fetches/persists data. Business rules (e.g. a variant inheriting the product's price when its own is null) live in the model layer as methods (`Variant.EffectivePrice`), not in handlers or repositories.
- **Sentinel domain errors**: repositories translate infrastructure errors (`gorm.ErrRecordNotFound`, Postgres unique-violation `23505`) into domain-level sentinels (`models.ErrProductNotFound`, `models.ErrCategoryAlreadyExists`), so `gorm`/`pq` stay implementation details hidden from handlers.
- **Pagination/filtering**: `GET /catalog` takes `offset`/`limit` (default 0/10, clamped to [1, 100]) and optional `category`/`price_max` filters, applied via query params since they're optional/collection-level, not resource identity.
- **Category responses never expose the internal numeric `ID`**: only `code` and `name`, per the assignment ("ID: internal use only").
- **Common JSON response helpers**: `app/api.OKResponse`/`ErrorResponse` centralize response formatting (`Content-Type`, status code, JSON body) and are used by all handlers.
- **Mocks generated with `mockery`** (`make mocks`, config in `.mockery.yml`) instead of hand-written test doubles, to keep interface/mock drift from becoming a maintenance burden.
- **Models encapsulate business logic**: methods on models (e.g., `Variant.EffectivePrice`) handle domain-specific rules, keeping handlers and repositories simple and focused on their respective responsibilities.
