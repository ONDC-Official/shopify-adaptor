package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"net/http"
	"os"
	"time"
)

// transformToONDCCatalog converts Shopify products into the ONDC catalog format.
func transformToONDCCatalog(products []shopifyProduct) ONDCCatalog {
	catalog := ONDCCatalog{
		Fulfillments: []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		}{
			{ID: "1", Type: "Delivery"},
			{ID: "2", Type: "Self-Pickup"},
			{ID: "3", Type: "Delivery and Self-Pickup"},
		},
		Descriptor: struct {
			Name      string   `json:"name"`
			Symbol    string   `json:"symbol"`
			ShortDesc string   `json:"short_desc"`
			LongDesc  string   `json:"long_desc"`
			Images    []string `json:"images"`
			Tags      []struct {
				Code string `json:"code"`
				List []struct {
					Code  string `json:"code"`
					Value string `json:"value"`
				} `json:"list"`
			} `json:"tags"`
		}{
			Name:      "Seller NP",
			Symbol:    "https://sellerNP.com/images/np.png",
			ShortDesc: "Seller Marketplace",
			LongDesc:  "Seller Marketplace",
			Images:    []string{"https://sellerNP.com/images/np.png"},
			Tags: []struct {
				Code string `json:"code"`
				List []struct {
					Code  string `json:"code"`
					Value string `json:"value"`
				} `json:"list"`
			}{
				{
					Code: "bpp_terms",
					List: []struct {
						Code  string `json:"code"`
						Value string `json:"value"`
					}{
						{Code: "np_type", Value: "MSN"},
						{Code: "accept_bap_terms", Value: "Y"},
						{Code: "collect_payment", Value: "Y"},
					},
				},
			},
		},
		Providers: []struct {
			ID    string `json:"id"`
			Items []struct {
				ID         string `json:"id"`
				Descriptor struct {
					Name string `json:"name"`
				} `json:"descriptor"`
				Price struct {
					Currency string `json:"currency"`
					Value    string `json:"value"`
				} `json:"price"`
			} `json:"items"`
		}{
			{
				ID:    "P1",
				Items: []struct {
					ID         string `json:"id"`
					Descriptor struct {
						Name string `json:"name"`
					} `json:"descriptor"`
					Price struct {
						Currency string `json:"currency"`
						Value    string `json:"value"`
					} `json:"price"`
				}{},
			},
		},
	}

	// Create items from Shopify products.
	for i, p := range products {
		itemID := fmt.Sprintf("I%d", i+1)
		item := struct {
			ID         string `json:"id"`
			Descriptor struct {
				Name string `json:"name"`
			} `json:"descriptor"`
			Price struct {
				Currency string `json:"currency"`
				Value    string `json:"value"`
			} `json:"price"`
		}{
			ID: itemID,
			Descriptor: struct {
				Name string `json:"name"`
			}{Name: p.Title},
			Price: struct {
				Currency string `json:"currency"`
				Value    string `json:"value"`
			}{
				Currency: "INR",
				Value:    p.Price,
			},
		}
		catalog.Providers[0].Items = append(catalog.Providers[0].Items, item)
	}

	return catalog
}

// sendOnSearch sends the ONDC on_search response to the BAP's on_search endpoint.
func sendOnSearch(bapURI string, catalog ONDCCatalog, req ONDCSearchRequest) error {
	response := struct {
		Context struct {
			Domain        string `json:"domain"`
			Country       string `json:"country"`
			City          string `json:"city"`
			Action        string `json:"action"`
			CoreVersion   string `json:"core_version"`
			BapID         string `json:"bap_id"`
			BapURI        string `json:"bap_uri"`
			BppID         string `json:"bpp_id"`
			BppURI        string `json:"bpp_uri"`
			TransactionID string `json:"transaction_id"`
			MessageID     string `json:"message_id"`
			Timestamp     string `json:"timestamp"`
		} `json:"context"`
		Message struct {
			Catalog ONDCCatalog `json:"catalog"`
		} `json:"message"`
	}{}

	// Populate the context using the incoming request and environment variables.
	response.Context.Domain = req.Context.Domain
	response.Context.Country = req.Context.Country
	response.Context.City = req.Context.City
	response.Context.Action = "on_search"
	response.Context.CoreVersion = "1.2.0"
	response.Context.BapID = req.Context.BapID
	response.Context.BapURI = req.Context.BapURI
	response.Context.BppID = os.Getenv("BPP_ID")
	response.Context.BppURI = os.Getenv("BPP_URI")
	response.Context.TransactionID = req.Context.TransactionID
	response.Context.MessageID = req.Context.MessageID
	response.Context.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	response.Message.Catalog = catalog

	// Format the JSON payload with indentation.
	prettyPayload, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal on_search response: %v", err)
	}

	log.Printf("Sending on_search to %s with payload:\n%s", bapURI+"/on_search", string(prettyPayload))

	resp, err := http.Post(bapURI+"/on_search", "application/json", bytes.NewBuffer(prettyPayload))
	if err != nil {
		return fmt.Errorf("failed to send on_search: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("on_search failed with status: %s", resp.Status)
	}

	log.Printf("Successfully sent on_search response for message ID: %s", req.Context.MessageID)
	return nil
}


// transformToONDCSelectResponse converts the Shopify response into an ONDC select response.
func transformToONDCSelectResponse(req ONDCSelectRequest, shopifyResp ShopifyResponse) ONDCSelectResponse {
	response := ONDCSelectResponse{
		Context: ONDCContext{
			Domain:        "ONDC:RET11",
			Action:        "on_select",
			CoreVersion:   "1.2.0",
			BapID:         req.Context.BapID,
			BapURI:        req.Context.BapURI,
			BppID:         req.Context.BppID,
			BppURI:        req.Context.BppURI,
			TransactionID: req.Context.TransactionID,
			MessageID:     req.Context.MessageID,
			City:          req.Context.City,
			Country:       req.Context.Country,
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Set provider ID.
	response.Message.Order.Provider.ID = "P1"

	// Transform items.
	var totalValue float64
	for _, edge := range shopifyResp.Data.Products.Edges {
		price := edge.Node.Variants.Edges[0].Node.Price
		priceFloat, _ := strconv.ParseFloat(price, 64)
		totalValue += priceFloat

		item := struct {
			ID            string `json:"id"`
			FulfillmentID string `json:"fulfillment_id"`
			Quantity      struct {
				Available int `json:"available"`
				Maximum   int `json:"maximum"`
			} `json:"quantity"`
			Price struct {
				Currency string `json:"currency"`
				Value    string `json:"value"`
			} `json:"price"`
			Breakup []struct {
				Title string `json:"title"`
				Price struct {
					Currency string `json:"currency"`
					Value    string `json:"value"`
				} `json:"price"`
			} `json:"breakup"`
		}{
			ID:            edge.Node.ID,
			FulfillmentID: "F1617",
			Quantity: struct {
				Available int `json:"available"`
				Maximum   int `json:"maximum"`
			}{
				Available: 5,
				Maximum:   3,
			},
			Price: struct {
				Currency string `json:"currency"`
				Value    string `json:"value"`
			}{
				Currency: "INR",
				Value:    price,
			},
			Breakup: []struct {
				Title string `json:"title"`
				Price struct {
					Currency string `json:"currency"`
					Value    string `json:"value"`
				} `json:"price"`
			}{
				{
					Title: fmt.Sprintf("Base Item - %s", edge.Node.Title),
					Price: struct {
						Currency string `json:"currency"`
						Value    string `json:"value"`
					}{
						Currency: "INR",
						Value:    price,
					},
				},
			},
		}
		response.Message.Order.Items = append(response.Message.Order.Items, item)
	}

	// Get delivery fee from configuration.
	deliveryFeeStr := os.Getenv("DELIVERY_FEE")
	if deliveryFeeStr == "" {
		deliveryFeeStr = "30.00"
	}
	deliveryCharge, err := strconv.ParseFloat(deliveryFeeStr, 64)
	if err != nil {
		deliveryCharge = 30.0
	}
	totalValue += deliveryCharge

	// Add fulfillment (delivery charge).
	fulfillment := struct {
		ID            string `json:"id"`
		Type          string `json:"type"`
		ProviderName  string `json:"@ondc/org/provider_name"`
		Tracking      bool   `json:"tracking"`
		Category      string `json:"@ondc/org/category"`
		TAT           string `json:"@ondc/org/TAT"`
		State         struct {
			Descriptor struct {
				Code string `json:"code"`
			} `json:"descriptor"`
		} `json:"state"`
		Price struct {
			Currency string `json:"currency"`
			Value    string `json:"value"`
		} `json:"price"`
	}{
		ID:           "F1617",
		Type:         "Delivery",
		ProviderName: "LSP Delivery",
		Tracking:     false,
		Category:     "Immediate Delivery",
		TAT:          "PT60M",
		State: struct {
			Descriptor struct {
				Code string `json:"code"`
			} `json:"descriptor"`
		}{
			Descriptor: struct {
				Code string `json:"code"`
			}{
				Code: "Serviceable",
			},
		},
		Price: struct {
			Currency string `json:"currency"`
			Value    string `json:"value"`
		}{
			Currency: "INR",
			Value:    fmt.Sprintf("%.2f", deliveryCharge),
		},
	}
	response.Message.Order.Fulfillments = append(response.Message.Order.Fulfillments, fulfillment)

	// Add quote.
	response.Message.Order.Quote = struct {
		Price struct {
			Currency string `json:"currency"`
			Value    string `json:"value"`
		} `json:"price"`
		Breakup []struct {
			Title string `json:"title"`
			Price struct {
				Currency string `json:"currency"`
				Value    string `json:"value"`
			} `json:"price"`
		} `json:"breakup"`
	}{
		Price: struct {
			Currency string `json:"currency"`
			Value    string `json:"value"`
		}{
			Currency: "INR",
			Value:    fmt.Sprintf("%.2f", totalValue),
		},
		Breakup: []struct {
			Title string `json:"title"`
			Price struct {
				Currency string `json:"currency"`
				Value    string `json:"value"`
			} `json:"price"`
		}{
			{
				Title: "Product Total",
				Price: struct {
					Currency string `json:"currency"`
					Value    string `json:"value"`
				}{
					Currency: "INR",
					Value:    fmt.Sprintf("%.2f", totalValue-deliveryCharge),
				},
			},
			{
				Title: "Delivery Charge",
				Price: struct {
					Currency string `json:"currency"`
					Value    string `json:"value"`
				}{
					Currency: "INR",
					Value:    fmt.Sprintf("%.2f", deliveryCharge),
				},
			},
		},
	}

	return response
}


func sendOnSelect(bapURI string, response ONDCSelectResponse, req ONDCSelectRequest) error {
    log.Printf("Sending on_select response for transaction %s", req.Context.TransactionID)

    jsonResponse, err := json.MarshalIndent(response, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal on_select response: %v", err)
    }

    log.Printf("on_select Response: %s", string(jsonResponse))

    resp, err := http.Post(bapURI+"/on_select", "application/json", bytes.NewBuffer(jsonResponse))
    if err != nil {
        return fmt.Errorf("failed to send on_select: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("on_select failed with status: %s", resp.Status)
    }

    log.Printf("Successfully sent on_select response for transaction ID: %s", req.Context.TransactionID)
    return nil
}

// transformToONDCInitResponse converts the /init request into an /on_init response.
func transformToONDCInitResponse(req ONDCInitRequest) ONDCInitResponse {
    return ONDCInitResponse{
        Context: ONDCContext{
            Domain:        req.Context.Domain,
            Action:        "on_init",
            CoreVersion:   req.Context.CoreVersion,
            BapID:         req.Context.BapID,
            BapURI:        req.Context.BapURI,
            BppID:         req.Context.BppID,
            BppURI:        req.Context.BppURI,
            TransactionID: req.Context.TransactionID,
            MessageID:     req.Context.MessageID,
            City:          req.Context.City,
            Country:       req.Context.Country,
            Timestamp:     time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
        },
        Message: struct {
            Order struct {
                Provider struct {
                    ID        string   `json:"id"`
                    Locations []struct {
                        ID string `json:"id"`
                    } `json:"locations"`
                } `json:"provider"`
                Items []struct {
                    ID       string `json:"id"`
                    Quantity struct {
                        Available int `json:"available"`
                        Maximum   int `json:"maximum"`
                    } `json:"quantity"`
                    Price struct {
                        Currency string `json:"currency"`
                        Value    string `json:"value"`
                    } `json:"price"`
                } `json:"items"`
                Payment struct {
                    Type   string `json:"type"`
                    Status string `json:"status"`
                } `json:"payment"`
                Fulfillments []struct {
                    ID            string `json:"id"`
                    Type          string `json:"type"`
                    ProviderName  string `json:"@ondc/org/provider_name"`
                    Tracking      bool   `json:"tracking"`
                    Category      string `json:"@ondc/org/category"`
                    TAT           string `json:"@ondc/org/TAT"`
                    Price         struct {
                        Currency string `json:"currency"`
                        Value    string `json:"value"`
                    } `json:"price"`
                    State struct {
                        Descriptor struct {
                            Code string `json:"code"`
                        } `json:"descriptor"`
                    } `json:"state"`
                } `json:"fulfillments"`
                Terms struct {
                    StaticTerms   string `json:"static_terms"`
                    EffectiveDate string `json:"effective_date"`
                } `json:"terms"`
            } `json:"order"`
        }{
            Order: struct {
                Provider struct {
                    ID        string   `json:"id"`
                    Locations []struct {
                        ID string `json:"id"`
                    } `json:"locations"`
                } `json:"provider"`
                Items []struct {
                    ID       string `json:"id"`
                    Quantity struct {
                        Available int `json:"available"`
                        Maximum   int `json:"maximum"`
                    } `json:"quantity"`
                    Price struct {
                        Currency string `json:"currency"`
                        Value    string `json:"value"`
                    } `json:"price"`
                } `json:"items"`
                Payment struct {
                    Type   string `json:"type"`
                    Status string `json:"status"`
                } `json:"payment"`
                Fulfillments []struct {
                    ID            string `json:"id"`
                    Type          string `json:"type"`
                    ProviderName  string `json:"@ondc/org/provider_name"`
                    Tracking      bool   `json:"tracking"`
                    Category      string `json:"@ondc/org/category"`
                    TAT           string `json:"@ondc/org/TAT"`
                    Price         struct {
                        Currency string `json:"currency"`
                        Value    string `json:"value"`
                    } `json:"price"`
                    State struct {
                        Descriptor struct {
                            Code string `json:"code"`
                        } `json:"descriptor"`
                    } `json:"state"`
                } `json:"fulfillments"`
                Terms struct {
                    StaticTerms   string `json:"static_terms"`
                    EffectiveDate string `json:"effective_date"`
                } `json:"terms"`
            }{
                Provider: struct {
                    ID        string   `json:"id"`
                    Locations []struct {
                        ID string `json:"id"`
                    } `json:"locations"`
                }{
                    ID: "P1",
                    Locations: []struct {
                        ID string `json:"id"`
                    }{
                        {ID: "L1"},
                    },
                },
                Items: []struct {
                    ID       string `json:"id"`
                    Quantity struct {
                        Available int `json:"available"`
                        Maximum   int `json:"maximum"`
                    } `json:"quantity"`
                    Price struct {
                        Currency string `json:"currency"`
                        Value    string `json:"value"`
                    } `json:"price"`
                }{
                    {
                        ID: "I1",
                        Quantity: struct {
                            Available int `json:"available"`
                            Maximum   int `json:"maximum"`
                        }{
                            Available: 10,
                            Maximum:   5,
                        },
                        Price: struct {
                            Currency string `json:"currency"`
                            Value    string `json:"value"`
                        }{
                            Currency: "INR",
                            Value:    "269.00",
                        },
                    },
                },
                Payment: struct {
                    Type   string `json:"type"`
                    Status string `json:"status"`
                }{
                    Type:   "ON-FULFILLMENT",
                    Status: "Pending",
                },
                Fulfillments: []struct {
                    ID            string `json:"id"`
                    Type          string `json:"type"`
                    ProviderName  string `json:"@ondc/org/provider_name"`
                    Tracking      bool   `json:"tracking"`
                    Category      string `json:"@ondc/org/category"`
                    TAT           string `json:"@ondc/org/TAT"`
                    Price         struct {
                        Currency string `json:"currency"`
                        Value    string `json:"value"`
                    } `json:"price"`
                    State struct {
                        Descriptor struct {
                            Code string `json:"code"`
                        } `json:"descriptor"`
                    } `json:"state"`
                }{
                    {
                        ID:           "F1",
                        Type:         "Delivery",
                        ProviderName: "Seller NP Name",
                        Tracking:     false,
                        Category:     "Immediate Delivery",
                        TAT:          "PT60M",
                        Price: struct {
                            Currency string `json:"currency"`
                            Value    string `json:"value"`
                        }{
                            Currency: "INR",
                            Value:    "30.00",
                        },
                        State: struct {
                            Descriptor struct {
                                Code string `json:"code"`
                            } `json:"descriptor"`
                        }{
                            Descriptor: struct {
                                Code string `json:"code"`
                            }{
                                Code: "Serviceable",
                            },
                        },
                    },
                },
                Terms: struct {
                    StaticTerms   string `json:"static_terms"`
                    EffectiveDate string `json:"effective_date"`
                }{
                    StaticTerms:   "https://github.com/ONDC-Official/NP-Static-Terms/buyerNP_BNP/1.0/tc.pdf",
                    EffectiveDate: "2023-10-01T00:00:00.000Z",
                },
            },
        },
    }
}

// sendOnInit sends the ONDC on_init response to the BAP's on_init endpoint.
func sendOnInit(bapURI string, response ONDCInitResponse, req ONDCInitRequest) error {
    log.Printf("Sending on_init response for transaction %s", req.Context.TransactionID)

    jsonResponse, err := json.MarshalIndent(response, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal on_init response: %v", err)
    }

    log.Printf("on_init Response: %s", string(jsonResponse))

    resp, err := http.Post(bapURI+"/on_init", "application/json", bytes.NewBuffer(jsonResponse))
    if err != nil {
        return fmt.Errorf("failed to send on_init: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("on_init failed with status: %s", resp.Status)
    }

    log.Printf("Successfully sent on_init response for transaction ID: %s", req.Context.TransactionID)
    return nil
}
