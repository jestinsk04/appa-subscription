package shopify

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"go.uber.org/zap"
)

const (
	customNamespace = "custom"
	orderDogDataKey = "dog_data"
)

// Repository defines methods to interact with Shopify API
type Repository interface {
	GetDogData(ctx context.Context, gid string) (*Pets, error)
	GetVariantByID(ctx context.Context, gid string) (*Variant, error)
	CreateOrder(ctx context.Context, input any) (*OrderCreateResponse, error)
}

// Repository is a Shopify API repository
type repository struct {
	gql    *GraphQLClient
	Logger *zap.Logger
}

// NewRepository creates a new Shopify API repository
func NewRepository(
	shopDomain, apiVersion, adminToken string, logger *zap.Logger,
) Repository {
	return &repository{
		gql:    NewGraphQLClient(shopDomain, apiVersion, adminToken, logger),
		Logger: logger,
	}
}

// GetDogData retrieves dog data from the Shopify API.
func (r *repository) GetDogData(
	ctx context.Context, gid string,
) (*Pets, error) {
	if !strings.Contains(gid, orderKind) {
		gid = GID(orderKind, gid)
	}
	vars := map[string]any{
		"id":        gid,
		"namespace": customNamespace,
		"key":       orderDogDataKey,
	}

	var resp GetOrderMetafieldResponse
	if err := r.gql.Do(ctx, getOrderMetafield, vars, &resp); err != nil {
		r.Logger.Error("failed to get order metafield", zap.Error(err), zap.Any("vars", vars))
		return nil, err
	}

	if resp.Order.Metafield == nil {
		r.Logger.Error("customer dog data metafield not found", zap.Any("vars", vars))
		return nil, errors.New("customer dog data metafield not found")
	}

	var pets Pets
	if err := json.Unmarshal(resp.Order.Metafield.JsonValue, &pets); err != nil {
		r.Logger.Error("failed to unmarshal order dog data", zap.Error(err), zap.Any("jsonValue", string(resp.Order.Metafield.JsonValue)))
		return nil, err
	}

	return &pets, nil
}

// GetVariantByID retrieves a product variant by its ID.
func (r *repository) GetVariantByID(
	ctx context.Context, gid string,
) (*Variant, error) {
	if !strings.Contains(gid, "ProductVariant") {
		gid = GID("ProductVariant", gid)
	}
	vars := map[string]any{
		"id": gid,
	}

	var resp GetVariantByIDResponse
	if err := r.gql.Do(ctx, getVariantByID, vars, &resp); err != nil {
		return nil, err
	}

	return resp.ProductVariant, nil
}

// CreateOrder creates a order
func (r *repository) CreateOrder(
	ctx context.Context, input any,
) (*OrderCreateResponse, error) {
	var resp CreateOrderResponse
	if err := r.gql.Do(ctx, createOrder, input, &resp); err != nil {
		return nil, err
	}

	if resp.UserErrors != nil {
		r.Logger.Error("failed to create draft order", zap.Any("errors", resp.UserErrors))
		return nil, errors.New("failed to create draft order")
	}

	return &resp.OrderCreate.Order, nil
}
