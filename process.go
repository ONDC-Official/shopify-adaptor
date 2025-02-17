package main

import (
    "log"
)

// processSearch handles the search request, querying Shopify and sending on_search response
func processSearch(req ONDCSearchRequest) {
	city := req.Context.City
	log.Printf("Processing search for city: %s", city)

	// Query Shopify for products using a tag-based search
	products := queryShopify(city)

	if len(products) == 0 {
		log.Printf("No products found for search term: %s", city)
	}

	// Transform the Shopify products into an ONDC catalog format
	catalog := transformToONDCCatalog(products)

	// Pass the entire request structure for context
	if err := sendOnSearch(req.Context.BapURI, catalog, req); err != nil {
		log.Printf("Failed to send on_search for message ID %s: %v", req.Context.MessageID, err)
	}
}

func processSelect(req ONDCSelectRequest) {
    log.Printf("Processing select for transaction: %s", req.Context.TransactionID)

    // Query Shopify for product details
    shopifyProducts := queryShopifyForSelect(req)
    
    // Create ShopifyResponse structure from products
    shopifyResp := ShopifyResponse{
        Data: struct {
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
        }{
            Products: struct {
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
            }{
                Edges: make([]struct {
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
                }, len(shopifyProducts)),
            },
        },
    }

    // Convert shopifyProducts to the expected format
    for i, product := range shopifyProducts {
        shopifyResp.Data.Products.Edges[i].Node.ID = product.ID
        shopifyResp.Data.Products.Edges[i].Node.Title = product.Title
        shopifyResp.Data.Products.Edges[i].Node.Variants.Edges = []struct {
            Node struct {
                Price string `json:"price"`
            } `json:"node"`
        }{
            {
                Node: struct {
                    Price string `json:"price"`
                }{
                    Price: product.Price,
                },
            },
        }
    }

    if len(shopifyProducts) == 0 {
        log.Printf("No products found for selection: %v", req.Message.Order.Items)
        return
    }

    // Transform into ONDC select response format
    selectResponse := transformToONDCSelectResponse(req, shopifyResp)

    // Send on_select response
    if err := sendOnSelect(req.Context.BapURI, selectResponse, req); err != nil {
        log.Printf("Failed to send on_select for transaction %s: %v", req.Context.TransactionID, err)
    }
}
// processInit handles the init request, transforming it into an on_init response.
func processInit(req ONDCInitRequest) {
    log.Printf("Processing init for transaction: %s", req.Context.TransactionID)

    // Transform the init request into an ONDC on_init response.
    initResponse := transformToONDCInitResponse(req)

    // Send on_init response.
    if err := sendOnInit(req.Context.BapURI, initResponse, req); err != nil {
        log.Printf("Failed to send on_init for transaction %s: %v", req.Context.TransactionID, err)
    }
}
