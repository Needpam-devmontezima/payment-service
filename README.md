# Payment Service API

A secure backend API for handling payment processing with Stripe integration, built with Go and deployed on AWS ECS Fargate.

## Features

- 💳 Process payments via Stripe
- 🚀 Containerized with Docker
- ☁️ Deployed on AWS ECS with Load Balancer
- 📊 Health check monitoring

## Tech Stack

- **Language**: Go 1.24.3
- **Framework**: none
- **Database**: PostgreSQL (AWS RDS)
- **Infrastructure**: 
  - AWS ECS Fargate
  - Application Load Balancer
  - AWS RDS
- **Payment Processing**: First integration -> Stripe API

## Prerequisites

- Go 1.24+
- Docker 28.0.4+
- AWS CLI (for deployment)
- Stripe account

## Getting Started

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/payment-service.git
   cd payment-service

   docker build -f deploy/docker/Dockerfile.api -t payment-service .

   docker compose up -d