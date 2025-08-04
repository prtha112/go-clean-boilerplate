# Database Schema

## Entity Relationship Diagram

```
┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
│    products     │       │   order_items   │       │     orders      │
├─────────────────┤       ├─────────────────┤       ├─────────────────┤
│ id (UUID) PK    │◄──────┤ product_id (FK) │       │ id (UUID) PK    │
│ name            │       │ order_id (FK)   ├──────►│ customer_name   │
│ description     │       │ quantity        │       │ customer_email  │
│ price           │       │ price           │       │ status          │
│ stock           │       │ id (UUID) PK    │       │ total_amount    │
│ created_at      │       │ created_at      │       │ created_at      │
│ updated_at      │       │ updated_at      │       │ updated_at      │
└─────────────────┘       └─────────────────┘       └─────────────────┘
```

## Table Relationships

### 1. Products Table
- **Primary Key**: `id` (UUID)
- **Purpose**: Stores product information and inventory
- **Key Fields**: name, description, price, stock

### 2. Orders Table
- **Primary Key**: `id` (UUID)
- **Purpose**: Stores order header information
- **Key Fields**: customer_name, customer_email, status, total_amount

### 3. Order Items Table (Relation Table)
- **Primary Key**: `id` (UUID)
- **Foreign Keys**: 
  - `product_id` → `products.id` (RESTRICT on delete)
  - `order_id` → `orders.id` (CASCADE on delete)
- **Purpose**: Many-to-many relationship between orders and products
- **Key Fields**: quantity, price (snapshot of product price at order time)
- **Constraints**: 
  - UNIQUE(order_id, product_id) - prevents duplicate items in same order
  - quantity > 0
  - price >= 0

## Relationship Types

### Orders ↔ Products (Many-to-Many)
- One order can contain multiple products
- One product can be in multiple orders
- Implemented via `order_items` junction table

### Foreign Key Constraints

1. **order_items.product_id → products.id**
   - **ON DELETE RESTRICT**: Cannot delete a product if it's referenced in any order
   - **Purpose**: Maintains data integrity and order history

2. **order_items.order_id → orders.id**
   - **ON DELETE CASCADE**: Deleting an order automatically removes all its items
   - **Purpose**: Maintains referential integrity

## Indexes

### Products Table
- `idx_products_name` - Fast product name searches
- `idx_products_price` - Price-based queries

### Orders Table
- `idx_orders_customer_email` - Customer order lookup
- `idx_orders_status` - Status-based filtering
- `idx_orders_created_at` - Date-based queries

### Order Items Table
- `idx_order_items_order_id` - Fast order item lookup
- `idx_order_items_product_id` - Product-based queries

## Business Rules Enforced by Schema

1. **Stock Management**: Products track available inventory
2. **Order Integrity**: Orders cannot exist without valid customer information
3. **Item Validation**: Order items must have positive quantity and non-negative price
4. **Status Control**: Orders have predefined status values
5. **Price History**: Order items store price at time of purchase (not current product price)
6. **Referential Integrity**: Cannot delete products that are referenced in orders
7. **Unique Items**: Same product cannot appear twice in the same order (must update quantity instead)

## Data Flow

1. **Product Creation**: Products are created with initial stock
2. **Order Creation**: 
   - Order header created in `orders` table
   - Order items created in `order_items` table with current product prices
   - Product stock decremented automatically
3. **Order Updates**: Status changes tracked in `orders` table
4. **Order Cancellation**: Stock restored to products when order is cancelled