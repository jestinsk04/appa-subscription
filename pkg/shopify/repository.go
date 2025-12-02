package shopify

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	customNamespace  = "custom"
	orderDogDataKey  = "dog_data"
	orderUserDataKey = "user_data"
)

// Repository defines methods to interact with Shopify API
type Repository interface {
	GetDogData(ctx context.Context, gid string) (*Pets, error)
	GetVariantByID(ctx context.Context, gid string) (*Variant, error)
	CreateOrder(ctx context.Context, input any) (*OrderCreateResponse, error)
	GetUserData(ctx context.Context, gid string) (*User, error)
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
	intent := 0
	var resp GetOrderMetafieldResponse
	for intent < 10 {
		if err := r.gql.Do(ctx, getOrderMetafield, vars, &resp); err != nil {
			r.Logger.Error("failed to get order metafield", zap.Error(err), zap.Any("vars", vars))
			return nil, err
		}

		if resp.Order.Metafield != nil {
			break
		}
		intent++
		time.Sleep(1 * time.Second)
	}

	if resp.Order.Metafield == nil {
		r.Logger.Error("order dog data metafield not found", zap.Any("vars", vars))
		return nil, errors.New("order dog data metafield not found")
	}

	var pets Pets
	if err := json.Unmarshal(resp.Order.Metafield.JsonValue, &pets); err != nil {
		r.Logger.Error("failed to unmarshal order dog data", zap.Error(err), zap.Any("jsonValue", string(resp.Order.Metafield.JsonValue)))
		return nil, err
	}

	return &pets, nil
}

// GetUserData retrieves user data from the Shopify API.
func (r *repository) GetUserData(
	ctx context.Context, gid string,
) (*User, error) {
	if !strings.Contains(gid, orderKind) {
		gid = GID(orderKind, gid)
	}
	vars := map[string]any{
		"id":        gid,
		"namespace": customNamespace,
		"key":       orderUserDataKey,
	}
	intent := 0
	var resp GetOrderMetafieldResponse
	for intent < 10 {
		if err := r.gql.Do(ctx, getOrderMetafield, vars, &resp); err != nil {
			r.Logger.Error("failed to get order metafield", zap.Error(err), zap.Any("vars", vars))
			return nil, err
		}

		if resp.Order.Metafield != nil {
			break
		}
		intent++
		time.Sleep(1 * time.Second)
	}

	if resp.Order.Metafield == nil {
		r.Logger.Error("order dog data metafield not found", zap.Any("vars", vars))
		return nil, errors.New("order dog data metafield not found")
	}

	var user User
	if err := json.Unmarshal(resp.Order.Metafield.JsonValue, &user); err != nil {
		r.Logger.Error("failed to unmarshal order dog data", zap.Error(err), zap.Any("jsonValue", string(resp.Order.Metafield.JsonValue)))
		return nil, err
	}

	return &user, nil
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
