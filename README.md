# Kodiflow

Kodiflow is a simple tool that is meant to help freelancers and SMEs with filing thier kenyan returns by
automating KRA-Approved invoice generation in a simple manner. With the recent changes in Kenyan regulations,
it is expected that all invoices must be e-Tims accredited, and this tool should help those who do not want
to struggle with manual management of their invoices and the e-Tims platform.

## Features

- TBD

## Tech Stack

This application is built using Go and PostgreSQL.
More information will be updated as we continue to develop

## Setup Instructions

```bash
# Clone the project
git clone git@github.com:stivo-m/api.kodiflow.com.git

# Move into project dir
cd api.kodiflow.com

# Install dependencies
go install


# Setup the env files
cp .env.example .env
# NOTE:
# Update the .env variables to your liking
# The `APP_ENCRYPTION_KEY` must be a base 64 of 32 characters in length.
# You can use `openssl rand -base64 32` to generate one

# Source the env variables
source .env

# Run the server
go run cmd/api/main.go

# Run the worker service for queues (Optional)
go run cmd/worker/main.go
```
