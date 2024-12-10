# Loan Engine Service

A state machine-based loan processing engine service that handles loan lifecycles with robust transaction management and monitoring capabilities.

## Features

### Core Features
- **State Machine**: Implements loan processing workflow with state transitions and validations
- **State Movement Tracker**: Tracks and logs all state transitions in the loan lifecycle
- **Optimistic Locking**: Prevents concurrent modifications to loan data
- **Atomic Transaction**: Ensures data consistency across operations
- **Basic Authentication**: Secure API access control
- **Database Migration**: Structured database schema management
- **Metric Monitoring**: Integration with Prometheus for service monitoring
- **Email Notifications**: SendGrid integration for automated notifications
- **PDF Generation**: Document generation using GoPDF
- **File Hosting**: File storage integration with file.io

### Technical Scope Notes
The following features are considered out of scope or have specific assumptions:
- **Proof File Handling**: Assumes Client provides valid URLs for approval and disbursement processes
- **Rate & ROI Calculations**: Assumes calculations are performed at loan proposal stage
- **Master Data Management**: Does not handle master data for borrowers, investors, and employees (uses identifiers only)

## Prerequisites

- Go 1.x
- PostgreSQL
- Make (for running Makefile commands)
- SendGrid API Key
- Prometheus (for monitoring)

## Getting Started

### Installation

1. Clone the repository
```bash
git clone https://github.com/yoratyo/loan-engine.git
cd loan-engine
```

2. Install dependencies
```bash
go mod tidy
```

3. Set up environment variables
```bash
cp .env.example .env
# Edit .env with your configuration
```

### Database Setup

Use the provided docker compose to create database container:
```bash
make setup
```

Use the provided migration scripts to set up your database:
```bash
make migrate
```

### Running the Service

1. Start the service:
```bash
make run
```

2. To run unit test:
```bash
make test
```

## API Documentation

The service exposes RESTful endpoints for loan management. Detailed API documentation can be found on file postman collection on this repo.

### Authentication

All endpoints require basic authentication. Include the following header in your requests:
```
Authorization: Basic <base64-encoded-credentials>
```

## Monitoring

The service exposes Prometheus metrics at `/metrics` endpoint. Key metrics include:
- State transition latency
- API request counts
- Error rates
- Transaction processing time
