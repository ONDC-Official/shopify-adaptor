package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// queryShopify constructs and sends a GraphQL query to Shopify to retrieve
// products that have tags matching the search criteria.
// If the context city equals "std:080", it searches for products tagged with "Delhi" OR "080".
func queryShopify(city string) []shopifyProduct {
	shopifyURL := os.Getenv("SHOPIFY_URL")
	accessToken := os.Getenv("SHOPIFY_ACCESS_TOKEN")

	log.Printf("Starting Shopify query with URL: %s", shopifyURL)

	if shopifyURL == "" || accessToken == "" {
		log.Printf("Error: Missing Shopify configuration - URL: %s, Access Token: %s",
			shopifyURL,
			func() string {
				if accessToken != "" {
					return "**REDACTED**"
				}
				return "EMPTY"
			}())
		return nil
	}

	fullShopifyURL := shopifyURL + "/admin/api/2025-01/graphql.json"

	// If city is "std:080", use a combined query tag.
	queryTag := city
	if city == "std:080" {
		queryTag = "Delhi OR 080"
	}

	// Construct the GraphQL query.
	gqlQuery := fmt.Sprintf(`{
		products(first: 10, query: "tag:%s") {
			edges {
				node {
					id
					title
					variants(first: 1) {
						edges {
							node {
								price
							}
						}
					}
				}
			}
		}
	}`, queryTag)

	log.Printf("Executing GraphQL query: %s", gqlQuery)

	reqBody := struct {
		Query string `json:"query"`
	}{Query: gqlQuery}

	body, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Error marshaling GraphQL request body: %v", err)
		return nil
	}

	httpReq, err := http.NewRequest("POST", fullShopifyURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Error creating Shopify request: %v", err)
		return nil
	}
	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("X-Shopify-Access-Token", accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("Shopify request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	// Read and log the raw response.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil
	}
	log.Printf("Raw Shopify response: %s", string(respBody))

	var result struct {
		Data struct {
			Products struct {
				Edges []struct {
					Node struct {
						ID       string `json:"id"`
						Title    string `json:"title"`
						Variants struct {
							Edges []struct {
								Node struct {
									Price string `json:"price"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"variants"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"products"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Printf("Error decoding Shopify GraphQL response: %v", err)
		return nil
	}

	if len(result.Errors) > 0 {
		for _, gqlErr := range result.Errors {
			log.Printf("GraphQL error: %s", gqlErr.Message)
		}
		return nil
	}

	var products []shopifyProduct
	for _, edge := range result.Data.Products.Edges {
		node := edge.Node
		price := ""
		if len(node.Variants.Edges) > 0 {
			price = node.Variants.Edges[0].Node.Price
		}
		products = append(products, shopifyProduct{
			ID:    node.ID,
			Title: node.Title,
			Price: price,
		})
	}

	log.Printf("Successfully found %d products", len(products))
	return products
}

func queryShopifyForSelect(req ONDCSelectRequest) []shopifyProduct {
    log.Printf("Querying Shopify for select operation...")
    return queryShopify(req.Context.City) // Reusing existing query logic
}