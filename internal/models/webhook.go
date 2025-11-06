package models

type PriceSet struct {
	ShopMoney        ShopMoney        `json:"shop_money"`
	PresentmentMoney PresentmentMoney `json:"presentment_money"`
}

type ShopMoney struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type PresentmentMoney struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type LineItem struct {
	ProductID          int                  `json:"product_id"`
	CurrentQuantity    int                  `json:"current_quantity"`
	Name               string               `json:"name"`
	PriceSet           PriceSet             `json:"price_set"`
	SKU                string               `json:"sku"`
	Title              string               `json:"title"`
	VariantTitle       string               `json:"variant_title"`
	VariantID          int                  `json:"variant_id"`
	PreTaxPriceSet     PriceSet             `json:"pre_tax_price_set"`
	TaxLines           []TaxLine            `json:"tax_lines"`
	DiscountAllocation []DiscountAllocation `json:"discount_allocations"`
}

type DiscountAllocation struct {
	AmountSet PriceSet `json:"amount_set"`
}

type TaxLine struct {
	Rate  float64 `json:"rate"`
	Title string  `json:"title"`
}

type Address struct {
	Company     string `json:"company"`
	Address1    string `json:"address1"`
	Address2    string `json:"address2"`
	City        string `json:"city"`
	Province    string `json:"province"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Zip         string `json:"zip"`
	Phone       string `json:"phone"`
}

type Customer struct {
	ID             int     `json:"id"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Email          string  `json:"email"`
	Phone          string  `json:"phone"`
	DefaultAddress Address `json:"default_address"`
}

type Webhook struct {
	ID                       int            `json:"id"`
	AdminGraphqlAPIID        string         `json:"admin_graphql_api_id"`
	FinancialStatus          string         `json:"financial_status"`
	FulfillmentStatus        string         `json:"fulfillment_status"`
	ContactEmail             string         `json:"contact_email"`
	CreatedAt                string         `json:"created_at"`
	Currency                 string         `json:"currency"`
	CurrentShippingPriceSet  PriceSet       `json:"current_shipping_price_set"`
	CurrentSubtotalPriceSet  PriceSet       `json:"current_subtotal_price_set"`
	CurrentTotalPriceSet     PriceSet       `json:"current_total_price_set"`
	CurrentTotalTaxSet       PriceSet       `json:"current_total_tax_set"`
	CurrentTotalDiscountsSet PriceSet       `json:"current_total_discounts_set"`
	Name                     string         `json:"name"`
	PaymentGatewayNames      []string       `json:"payment_gateway_names"`
	SubTotalPriceSet         PriceSet       `json:"sub_total_price_set"`
	Tags                     string         `json:"tags"`
	LineItems                []LineItem     `json:"line_items"`
	Customer                 Customer       `json:"customer"`
	DefaultAddress           Address        `json:"default_address"`
	ShippingLines            []ShippingLine `json:"shipping_lines"`
	TaxLines                 []TaxLine      `json:"tax_lines"`
	DiscountCodes            any            `json:"discount_codes"`
	// ShippingAddress          Address        `json:"shipping_address"`
}

type ShippingLine struct {
	PriceSet           PriceSet `json:"price_set"`
	DiscountedPriceSet PriceSet `json:"discounted_price_set"`
}

type Discount struct {
	Amount string `json:"amount"`
	Code   string `json:"code"`
	Type   string `json:"type"`
}

type Variant struct {
	ID           string `json:"id"`
	ProductID    string `json:"product_id"`
	PetAge       string `json:"pet_age"`
	PetCondition string `json:"pet_condition"`
	PetSize      string `json:"pet_size"`
}
