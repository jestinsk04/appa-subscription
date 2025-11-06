package shopify

// GetOrderByID is the GraphQL query for retrieving an order by its ID
const getOrderByIDQuery = `
	query orderByIDQuery($id: ID!) {
		order(id: $id) {
			id
			name
			createdAt
			displayFinancialStatus
			displayFulfillmentStatus
			totalPriceSet {
				shopMoney {
					amount
					currencyCode
				}
			}
			subtotalPriceSet {
				shopMoney {
					amount
					currencyCode
				}
			}
			lineItems(first: 5) {
				edges {
					node {
						name
						quantity
						sku
						variant {
							id
							title
							selectedOptions {
								name
								value
							}
						}
					}
				}
			}
			customer {
				displayName
				id
				metafield(namespace: "customer_fields", key: "parent_id") {
					value
					key
				}
			}
		}
	}
`

const getOrderByName = `
query orderByName($query: String!, $first: Int!) {
	orders(first: $first, query: $query) {
		nodes {
			id
			name
			closedAt
			displayFinancialStatus
			displayFulfillmentStatus
			totalRefundedSet {
				shopMoney {
					amount
					currencyCode
				}
			}
			subtotalPriceSet {
				shopMoney {
					amount
					currencyCode
				}
			}
			lineItems(first: 5) {
				edges {
					node {
						name
						quantity
						sku
						variant {
							id
							title
							selectedOptions {
								name
								value
							}
						}
					}
				}
			}
			customer {
				displayName
				id
				metafield(namespace: "customer_fields", key: "parent_id") {
					value
					key
				}
			}
		}
		pageInfo {
		hasNextPage
		startCursor
		hasPreviousPage
		endCursor
		}
	}
}`

const getCustomerMetafield = `
query($id: ID!, $namespace: String!, $key: String!) {
  customer(id: $id) {
    metafield(namespace: $namespace, key: $key) { 
      key
      value
      jsonValue
    }
  }
}`

const getOrderMetafield = `
query($id: ID!, $namespace: String!, $key: String!) {
  order(id: $id) {
    metafield(namespace: $namespace, key: $key) { 
      key
      value
      jsonValue
    }
  }
}`

const getVariantByID = `
query GetVariantByID($id: ID!) {
  productVariant(id: $id) {
    id
	title
    selectedOptions {
      name
      value
    }
  }
}`

const createOrder = `
mutation orderCreate($order: OrderCreateOrderInput!, $options: OrderCreateOptionsInput) {
  orderCreate(order: $order, options: $options) {
    order {
      id
      name
      statusPageUrl
	  TotalPriceSet {
		shopMoney {
			amount
			currencyCode
		}
	  }
    }
    userErrors {
      field
      message
    }
  }
}
`
