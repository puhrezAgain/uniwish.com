# Unilist.com

A sort of universal wishlist service. Inspired by the need to aggregate and track items across websites, this application creates an explorable repository of price-monitored products the user wants to be updated on.

The original domain to be considered is fashion items because this is regularly wishlisted product category due to it's seasonality, the variable availability of goods depending on demand, and price volatity which is informed by all kinds of factors. 

## Purpose

Use this product idea to develop a modern backend system building a REST API in Go, queue based workers and Postgres for application usage. 

## Architecture
```
┌───────────────────────────┐
│     Chrome Extension      │
│  (Add item to wishlist)   │
│                           │
│  - product URL            │
│  - shop identifier        │
│  - user auth token        │
└─────────────┬─────────────┘
              │ HTTP / gRPC
              ▼
┌───────────────────────────┐
│       SKU Processor       │
│   (API / Backend Service) │
│                           │
│  - normalize URL          │
│  - extract SKU / key      │
│  - check product exists   │
└─────────────┬─────────────┘
              │
      ┌───────┴────────┐
      │ product exists?│
      └───────┬────────┘
          YES │              NO
              │               │
              ▼               ▼
┌───────────────────┐   ┌────────────────────┐
│ Wishlist Database │   │  SKU Scraper Queue  │
│                   │   │  (DB / Redis / SQS) │
│ - users            │   │ - product_url      │
│ - products         │   │ - shop             │
│ - prices           │   │ - priority         │
│ - wishlists        │   └─────────┬──────────┘
└─────────┬─────────┘             │
          │                        ▼
          │              ┌────────────────────┐
          │              │    SKU Scraper     │
          │              │  (Workers / Jobs)  │
          │              │                    │
          │              │ - fetch page       │
          │              │ - parse name       │
          │              │ - parse image      │
          │              │ - parse price      │
          │              │ - validate         │
          │              └─────────┬──────────┘
          │                        │
          └───────────────◄────────┘
                   (upsert product + price)

              ▼
┌───────────────────────────┐
│        Dashboard          │
│     (Web Application)     │
│                           │
│ - user login              │
│ - wishlist view           │
│ - price history           │
│ - price drop alerts       │
└───────────────────────────┘
```