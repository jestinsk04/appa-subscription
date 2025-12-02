package shopify

import (
	"encoding/json"
)

type gqlRequest struct {
	Query     string `json:"query"`
	Variables any    `json:"variables"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// emun shopify kind
const (
	orderKind          = "Order"
	CustomerKind       = "Customer"
	ProductVariantKind = "ProductVariant"
)

// GetOrderByIDResponse constructs a global ID for Shopify entities
type GetOrderByIDResponse struct {
	Order *Order `json:"order"`
}

// GetOrderByQueryResponse represents the response for querying multiple orders
type GetOrderByQueryResponse struct {
	Orders OrdersNodes `json:"orders"`
}

// OrdersNodes represents a list of order nodes
type OrdersNodes struct {
	Nodes []Order `json:"nodes"`
}

// OrdersByPage represents a paginated list of orders
type OrdersByPage struct {
	Order
	PageInfo PageInfo `json:"pageInfo"`
}

// PageInfo represents pagination information for a list of orders
type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	StartCursor     string `json:"startCursor"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	EndCursor       string `json:"endCursor"`
}

// Order represents a Shopify order
type Order struct {
	ID                       string        `json:"id"`
	Name                     string        `json:"name"`
	CreatedAt                string        `json:"createdAt"`
	DisplayFinancialStatus   string        `json:"displayFinancialStatus"`
	DisplayFulfillmentStatus string        `json:"displayFulfillmentStatus"`
	TotalPriceSet            ShopMoney     `json:"totalPriceSet"`
	SubtotalPriceSet         ShopMoney     `json:"subtotalPriceSet"`
	LineItems                LineItemsEdge `json:"lineItems"`
	Customer                 Customer      `json:"customer"`
}

// LineItemsEdge represents the edge of line items in an order
type LineItemsEdge struct {
	Edges []LineItemsNode `json:"edges"`
}

// LineItemsNode represents a node in the line items edge
type LineItemsNode struct {
	Node LineItem `json:"node"`
}

// LineItem represents a line item in an order
type LineItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	SKU      string  `json:"sku"`
	Variant  Variant `json:"variant"`
}

// Variant represents a product variant
type Variant struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	SelectedOptions []SelectedOption `json:"selectedOptions"`
}

// SelectedOption represents a selected option for a product variant
type SelectedOption struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ShopMoney represents the total or subtotal price set of an order
type ShopMoney struct {
	ShopMoney ShopMoneyProps `json:"shopMoney"`
}

// ShopMoneyProps represents an amount of money in a specific currency
type ShopMoneyProps struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currencyCode"`
}

// Customer represents a Shopify customer
type Customer struct {
	ID          string     `json:"id"`
	DisplayName string     `json:"displayName"`
	Metafield   *Metafield `json:"metafield"`
}

// Metafield represents a Shopify metafield
type Metafield struct {
	Value     string          `json:"value"`
	Key       string          `json:"key"`
	JsonValue json.RawMessage `json:"jsonValue,omitempty"`
}

// QueryOrderFilter represents the filters for querying orders
type QueryOrderFilter struct {
	Name string
}

type GetUserByIDResponse struct {
	Customer *Customer `json:"customer"`
}

type GetCustomerMetafieldResponse struct {
	Customer struct {
		Metafield *Metafield `json:"metafield"`
	} `json:"customer"`
}

type GetOrderMetafieldResponse struct {
	Order struct {
		Metafield *Metafield `json:"metafield"`
	} `json:"order"`
}

// Pets represents a list of pet objects
type Pets struct {
	Pets []Pet `json:"pets"`
}

// Pet represents a single pet's information
type Pet struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	Gender           string `json:"gender"`
	Neutered         string `json:"neutered"`
	Birthday         string `json:"birthday"`
	Breed            string `json:"breed"`
	ProductVariantID string `json:"product_variant_id"`
}

type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	DocType   string `json:"docType"`
	DocNumber string `json:"docNumber"`
	City      string `json:"city"`
	State     string `json:"state"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
}

type GetVariantByIDResponse struct {
	ProductVariant *Variant `json:"productVariant"`
}

type CreateOrderResponse struct {
	OrderCreate struct {
		Order OrderCreateResponse `json:"order"`
	} `json:"orderCreate"`
	UserErrors []UserErrors `json:"userErrors"`
}

type OrderCreateResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	StatusPageURL string    `json:"statusPageUrl"`
	TotalPriceSet ShopMoney `json:"totalPriceSet"`
}

type MarkOrderAsPaidResponse struct {
	Order      Order       `json:"order"`
	UserErrors *UserErrors `json:"userErrors"`
}

type UserErrors struct {
	Message string `json:"message"`
}

type CreateOrderInShopifyRequest struct {
	Order CreateOrderInput `json:"order"`
}

type CreateOrderInput struct {
	CustomerID      string                 `json:"customerId"`
	Email           string                 `json:"email"`
	Tags            []string               `json:"tags"`
	LineItems       []LineItemsNodeRequest `json:"lineItems"`
	Note            string                 `json:"note,omitempty"`
	FinancialStatus string                 `json:"financialStatus,omitempty"`
}

type LineItemsNodeRequest struct {
	VariantID string `json:"variantId"`
	Quantity  int    `json:"quantity"`
}

type ShippingAddress struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Phone     string `json:"phone"`
	Address1  string `json:"address1"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Country   string `json:"country"`
	Zip       string `json:"zip"`
}
