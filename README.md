# Shopify ONDC Adapter

A Go-based adapter that connects Shopify stores to the ONDC (Open Network for Digital Commerce) network. This adapter enables Shopify merchants to list their products on ONDC by translating between Shopify's GraphQL API and ONDC's protocol.

## Prerequisites
- Go 1.19 or higher
- A Shopify store with API access
- Shopify Admin API access token
- Netcat (`nc`) for testing callbacks

## Configuration

### Environment Variables
The adapter requires the following environment variables:

```sh
SHOPIFY_URL=https://your-store.myshopify.com
SHOPIFY_ACCESS_TOKEN=your_access_token
```

Set these in your environment or they will default to test values if not set.

## Installation

Clone the repository:

```sh
git clone [repository-url]
cd shopify-ondc-adapter
```

Install dependencies:

```sh
go mod download
go mod tidy
```

## Running the Server

Start the server:

```sh
go run .
```

The server will start on port `9090`.

## Testing

### 1. Start the Callback Listener
Open a terminal and start a netcat listener to simulate the ONDC callback endpoint:

```sh
nc -l -p 9091
```

### 2. Send a Test Search Request

In another terminal, send a test search request to `localhost:9090`:

```sh
curl -X POST http://localhost:9090/search \
  -H "Content-Type: application/json" \
  -d '{
    "context": {
      "domain": "ONDC:RET10",
      "action": "search",
      "country": "IND",
      "city": "std:080",
      "core_version": "1.2.0",
      "bap_id": "buyerNP.com",
      "bap_uri": "http://localhost:9091",
      "transaction_id": "T1",
      "message_id": "M1",
      "timestamp": "2025-02-08T04:32:05.303Z",
      "ttl": "PT30S"
    },
    "message": {
      "intent": {
        "category": {
          "id": "Foodgrains"
        },
        "fulfillment": {
          "type": "Delivery"
        },
        "payment": {
          "@ondc/org/buyer_app_finder_fee_type": "percent",
          "@ondc/org/buyer_app_finder_fee_amount": "3"
        },
        "tags": [
          {
            "code": "bap_terms",
            "list": [
              {
                "code": "static_terms",
                "value": ""
              },
              {
                "code": "static_terms_new",
                "value": "https://github.com/ONDC-Official/NP-Static-Terms/buyerNP_BNP/1.0/tc.pdf"
              },
              {
                "code": "effective_date",
                "value": "2023-10-01T00:00:00.000Z"
              }
            ]
          }
        ]
      }
    }
  }'
```

###  Verify the Responses
You should see:
- An immediate ACK response from the `/search` endpoint.
- Server logs showing the Shopify GraphQL query.
- The callback response in your netcat (`nc`) terminal, similar to:

```json
{
  "context": {
    "domain": "ONDC:RET10",
    "country": "IND",
    "city": "std:080",
    "action": "on_search",
    "core_version": "1.2.0",
    "bap_id": "buyerNP.com",
    "bap_uri": "http://localhost:9091",
    "bpp_id": "sellerNP.com",
    "bpp_uri": "http://localhost:9090",
    "transaction_id": "T1",
    "message_id": "M1",
    "timestamp": "2025-02-08T04:32:05.303Z"
  },
  "message": {
    "catalog": {
      "bpp/fulfillments": [
        {
          "id": "1",
          "type": "Delivery"
        },
        {
          "id": "2",
          "type": "Self-Pickup"
        }
      ],
      "bpp/descriptor": {
        "name": "Seller NP",
        "symbol": "https://sellerNP.com/images/np.png"
      },
      "bpp/providers": [
        {
          "id": "P1",
          "items": [
            {
              "id": "I1",
              "descriptor": {
                "name": "Example T-Shirt"
              },
              "price": {
                "currency": "INR",
                "value": "25.00"
              }
            }
          ]
        }
      ]
    }
  }
}
```

### 3. Send a Test Select Request
``` 
curl -X POST http://localhost:9090/select \
  -H "Content-Type: application/json" \
  -d '{
    "context": {
      "domain": "ONDC:RET11",
      "action": "select",
      "core_version": "1.2.0",
      "bap_id": "buyerNP.com",
      "bap_uri": "http://localhost:9091",
      "bpp_id": "sellerNP.com",
      "bpp_uri": "http://localhost:9090",
      "transaction_id": "T2",
      "message_id": "M2",
      "city": "std:080",
      "country": "IND",
      "timestamp": "2025-02-08T05:00:00.000Z",
      "ttl": "PT30S"
    },
    "message": {
      "order": {
        "provider": {
          "id": "P1"
        },
        "items": [
          {
            "id": "DI1",
            "parent_item_id": "BaseItem1",
            "location_id": "L1",
            "quantity": {
              "count": 1
            },
            "tags": [
              {
                "code": "type",
                "list": [
                  {
                    "code": "type",
                    "value": "item"
                  }
                ]
              }
            ]
          },
          {
            "id": "C1",
            "parent_item_id": "DI1",
            "location_id": "L1",
            "quantity": {
              "count": 1
            },
            "tags": [
              {
                "code": "type",
                "list": [
                  {
                    "code": "type",
                    "value": "customization"
                  }
                ]
              }
            ]
          }
        ],
        "offers": [
          {
            "id": "BUY2GET3",
            "tags": [
              {
                "code": "selection",
                "list": [
                  {
                    "code": "apply",
                    "value": "yes"
                  }
                ]
              }
            ]
          }
        ],
        "fulfillments": [
          {
            "end": {
              "location": {
                "gps": "12.453544,77.928379",
                "address": {
                  "area_code": "560001"
                }
              }
            },
            "price": {
              "currency": "INR",
              "value": "30.00"
            }
          }
        ],
        "payment": {
          "type": "ON-FULFILLMENT"
        }
      }
    }
  }'
```

### Verify on select Response

``` 
{
  "context": {
    "domain": "ONDC:RET11",
    "action": "on_select",
    "core_version": "1.2.0",
    "bap_id": "buyerNP.com",
    "bap_uri": "http://localhost:9091",
    "bpp_id": "sellerNP.com",
    "bpp_uri": "http://localhost:9090",
    "transaction_id": "T2",
    "message_id": "M2",
    "city": "std:080",
    "country": "IND",
    "timestamp": "2025-02-13T13:09:29Z"
  },
  "message": {
    "order": {
      "provider": {
        "id": "P1"
      },
      "items": [
        {
          "id": "gid://shopify/Product/15045213487475",
          "fulfillment_id": "F1617",
          "quantity": {
            "available": 5,
            "maximum": 3
          },
          "price": {
            "currency": "INR",
            "value": "25.00"
          },
          "breakup": [
            {
              "title": "Base Item - Example T-Shirt",
              "price": {
                "currency": "INR",
                "value": "25.00"
              }
            }
          ]
        },
        {
          "id": "gid://shopify/Product/15045213520243",
          "fulfillment_id": "F1617",
          "quantity": {
            "available": 5,
            "maximum": 3
          },
          "price": {
            "currency": "INR",
            "value": "49.99"
          },
          "breakup": [
            {
              "title": "Base Item - Example Pants",
              "price": {
                "currency": "INR",
                "value": "49.99"
              }
            }
          ]
        },
        {
          "id": "gid://shopify/Product/15046962577779",
          "fulfillment_id": "F1617",
          "quantity": {
            "available": 5,
            "maximum": 3
          },
          "price": {
            "currency": "INR",
            "value": "90.00"
          },
          "breakup": [
            {
              "title": "Base Item - Clavin klen t shirt",
              "price": {
                "currency": "INR",
                "value": "90.00"
              }
            }
          ]
        }
      ],
      "fulfillments": [
        {
          "id": "F1617",
          "type": "Delivery",
          "@ondc/org/provider_name": "LSP Delivery",
          "tracking": false,
          "@ondc/org/category": "Immediate Delivery",
          "@ondc/org/TAT": "PT60M",
          "state": {
            "descriptor": {
              "code": "Serviceable"
            }
          },
          "price": {
            "currency": "INR",
            "value": "30.00"
          }
        }
      ],
      "quote": {
        "price": {
          "currency": "INR",
          "value": "194.99"
        },
        "breakup": [
          {
            "title": "Product Total",
            "price": {
              "currency": "INR",
              "value": "164.99"
            }
          },
          {
            "title": "Delivery Charge",
            "price": {
              "currency": "INR",
              "value": "30.00"
            }
          }
        ]
      }
    }
  }
}
```

### 4 Send init request

```
curl -X POST http://localhost:9090/init \
-H "Content-Type: application/json"   -d '{
    "context": {
      "domain": "ONDC:RET11",
      "action": "init",
      "core_version": "1.2.0",
      "bap_id": "buyerNP.com",
      "bap_uri": "http://localhost:9091",
      "bpp_id": "sellerNP.com",
      "bpp_uri": "http://localhost:9090",
      "transaction_id": "T1",
      "message_id": "M1",
      "city": "std:080",
      "country": "IND",
      "timestamp": "2025-02-08T05:00:00.000Z",
      "ttl": "PT30S"
    },
    "message": {
      "order": {
        "provider": {
          "id": "P1",
          "locations": [
            {"id": "L1"}
          ]
        },
        "items": [
          {
            "id": "I1",
            "quantity": {"count": 1}
          }
        ],
        "payment": {
          "type": "ON-FULFILLMENT"
        },
        "fulfillments": [
          {
            "id": "F1",
            "type": "Delivery"
          }
        ],
        "terms": {
          "static_terms": "https://github.com/ONDC-Official/NP-Static-Terms/buyerNP_BNP/1.0/tc.pdf",
          "effective_date": "2023-10-01T00:00:00.000Z"
        }
      }
    }
  }'

```

### Verify on_init

```
{
  "context": {
    "domain": "ONDC:RET11",
    "action": "on_init",
    "core_version": "1.2.0",
    "bap_id": "buyerNP.com",
    "bap_uri": "http://localhost:9091",
    "bpp_id": "sellerNP.com",
    "bpp_uri": "http://localhost:9090",
    "transaction_id": "T1",
    "message_id": "M1",
    "city": "std:080",
    "country": "IND",
    "timestamp": "2025-02-17T05:37:40.356Z"
  },
  "message": {
    "order": {
      "provider": {
        "id": "P1",
        "locations": [
          {
            "id": "L1"
          }
        ]
      },
      "items": [
        {
          "id": "I1",
          "quantity": {
            "available": 10,
            "maximum": 5
          },
          "price": {
            "currency": "INR",
            "value": "269.00"
          }
        }
      ],
      "payment": {
        "type": "ON-FULFILLMENT",
        "status": "Pending"
      },
      "fulfillments": [
        {
          "id": "F1",
          "type": "Delivery",
          "@ondc/org/provider_name": "Seller NP Name",
          "tracking": false,
          "@ondc/org/category": "Immediate Delivery",
          "@ondc/org/TAT": "PT60M",
          "price": {
            "currency": "INR",
            "value": "30.00"
          },
          "state": {
            "descriptor": {
              "code": "Serviceable"
            }
          }
        }
      ],
      "terms": {
        "static_terms": "https://github.com/ONDC-Official/NP-Static-Terms/buyerNP_BNP/1.0/tc.pdf",
        "effective_date": "2023-10-01T00:00:00.000Z"
      }
    }
  }
}


```

## Troubleshooting

### Common Issues

#### 1. Empty Product List
- Verify your Shopify store has products.
- Check if the search term matches product titles.
- Check server logs for Shopify GraphQL query responses.

#### 2. Connection Refused on Callback
- Ensure `nc` (netcat) is running on port 9091.
- Verify `bap_uri` in your request matches the netcat port.

#### 3. Authentication Errors
- Ensure `SHOPIFY_ACCESS_TOKEN` is correct.
- Verify the token has the necessary permissions.

## Debugging
The server provides detailed logs for debugging:
- Incoming search requests
- GraphQL queries sent to Shopify
- Raw responses from Shopify
- Callback attempts and responses

Monitor these logs in the terminal where you run `go run .`

## API Reference

### Search Endpoint

**POST** `/search`

#### Request Body:
```json
{
  "context": {
    "domain": "ONDC:RET10",
    "bap_uri": "string",
    "message_id": "string"
  },
  "message": {
    "intent": {
      "item": {
        "descriptor": {
          "name": "string"
        }
      }
    }
  }
}
```

#### Response:
```json
{
  "message": {
    "ack": {
      "status": "ACK"
    }
  }
}
```

## Contributing
Please submit issues and pull requests for any improvements you'd like to make.



