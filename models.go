package main

// ONDC Structures

// ONDCSearchRequest represents the search request payload from ONDC.
type ONDCSearchRequest struct {
	Context struct {
		Domain        string `json:"domain"`
		Action        string `json:"action"`
		Country       string `json:"country"`
		City          string `json:"city"`
		CoreVersion   string `json:"core_version"`
		BapID         string `json:"bap_id"`
		BapURI        string `json:"bap_uri"`
		TransactionID string `json:"transaction_id"`
		MessageID     string `json:"message_id"`
		Timestamp     string `json:"timestamp"`
		TTL           string `json:"ttl"`
	} `json:"context"`
	Message struct {
		Intent struct {
			Category struct {
				ID string `json:"id"`
			} `json:"category"`
			Fulfillment struct {
				// fulfillment section is no longer used for city code as per instructions
				Type string `json:"type"`
			} `json:"fulfillment"`
			Payment struct {
				FinderFeeType   string `json:"@ondc/org/buyer_app_finder_fee_type"`
				FinderFeeAmount string `json:"@ondc/org/buyer_app_finder_fee_amount"`
			} `json:"payment"`
			Tags []struct {
				Code string `json:"code"`
				List []struct {
					Code  string `json:"code"`
					Value string `json:"value"`
				} `json:"list"`
			} `json:"tags"`
		} `json:"intent"`
	} `json:"message"`
}

// ONDCCatalog represents the catalog structure to be sent in on_search.
type ONDCCatalog struct {
	Fulfillments []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"bpp/fulfillments"`
	Descriptor struct {
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
	} `json:"bpp/descriptor"`
	Providers []struct {
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
	} `json:"bpp/providers"`
}

// Shopify Structures

// shopifyProduct represents a product returned from Shopify.
type shopifyProduct struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Price string `json:"price"`
}

type ONDCContext struct {
    Domain        string `json:"domain"`
    Action        string `json:"action"`
    CoreVersion   string `json:"core_version"`
    BapID        string `json:"bap_id"`
    BapURI       string `json:"bap_uri"`
    BppID        string `json:"bpp_id"`
    BppURI       string `json:"bpp_uri"`
    TransactionID string `json:"transaction_id"`
    MessageID    string `json:"message_id"`
    City         string `json:"city"`
    Country      string `json:"country"`
    Timestamp    string `json:"timestamp"`
}

type ShopifyResponse struct {
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
}

type ONDCTag struct {
    Code string        `json:"code"`
    List []ONDCTagItem `json:"list"`
}
type ONDCTagItem struct {
    Code  string `json:"code"`
    Value string `json:"value"`
}

type ONDCSelectRequest struct {
    Context ONDCContext `json:"context"`
    Message struct {
        Order struct {
            Items []struct {
                ID            string `json:"id"`
                ParentItemID string `json:"parent_item_id"`
                LocationID   string `json:"location_id"`
                Quantity     struct {
                    Count int `json:"count"`
                } `json:"quantity"`
                Tags []ONDCTag `json:"tags"`
            } `json:"items"`
            Offers []struct {
                ID   string    `json:"id"`
                Tags []ONDCTag `json:"tags"`
            } `json:"offers"`
            Fulfillments []struct {
                End struct {
                    Location struct {
                        GPS     string `json:"gps"`
                        Address struct {
                            AreaCode string `json:"area_code"`
                        } `json:"address"`
                    } `json:"location"`
                } `json:"end"`
            } `json:"fulfillments"`
            Payment struct {
                Type string `json:"type"`
            } `json:"payment"`
        } `json:"order"`
    } `json:"message"`
}

type ONDCSelectResponse struct {
    Context ONDCContext `json:"context"`
    Message struct {
        Order struct {
            Provider struct {
                ID string `json:"id"`
            } `json:"provider"`
            Items []struct {
                ID            string `json:"id"`
                FulfillmentID string `json:"fulfillment_id"`
                Quantity     struct {
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
            } `json:"items"`
            Fulfillments []struct {
                ID            string `json:"id"`
                Type         string `json:"type"`
                ProviderName string `json:"@ondc/org/provider_name"`
                Tracking     bool   `json:"tracking"`
                Category     string `json:"@ondc/org/category"`
                TAT          string `json:"@ondc/org/TAT"`
                State       struct {
                    Descriptor struct {
                        Code string `json:"code"`
                    } `json:"descriptor"`
                } `json:"state"`
                Price struct {
                    Currency string `json:"currency"`
                    Value    string `json:"value"`
                } `json:"price"`
            } `json:"fulfillments"`
            Quote struct {
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
            } `json:"quote"`
        } `json:"order"`
    } `json:"message"`
}


// ONDCInitRequest represents the /init request payload.
type ONDCInitRequest struct {
    Context ONDCContext `json:"context"`
    Message struct {
        Order struct {
            Provider struct {
                ID        string `json:"id"`
                Locations []struct {
                    ID string `json:"id"`
                } `json:"locations"`
            } `json:"provider"`
            Items []struct {
                ID       string `json:"id"`
                Quantity struct {
                    Count int `json:"count"`
                } `json:"quantity"`
            } `json:"items"`
            Payment struct {
                Type string `json:"type"`
            } `json:"payment"`
            Fulfillments []struct {
                ID   string `json:"id"`
                Type string `json:"type"`
            } `json:"fulfillments"`
            Terms struct {
                StaticTerms   string `json:"static_terms"`
                EffectiveDate string `json:"effective_date"`
            } `json:"terms"`
        } `json:"order"`
    } `json:"message"`
}

// ONDCInitResponse represents the /on_init response payload.
type ONDCInitResponse struct {
    Context ONDCContext `json:"context"`
    Message struct {
        Order struct {
            Provider struct {
                ID        string `json:"id"`
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
                ID           string `json:"id"`
                Type         string `json:"type"`
                ProviderName string `json:"@ondc/org/provider_name"`
                Tracking     bool   `json:"tracking"`
                Category     string `json:"@ondc/org/category"`
                TAT          string `json:"@ondc/org/TAT"`
                Price        struct {
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
    } `json:"message"`
}

