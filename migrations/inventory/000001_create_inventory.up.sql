CREATE TABLE IF NOT EXISTS inventory.inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_variant_id UUID NOT NULL,
    warehouse VARCHAR(100) NOT NULL DEFAULT 'main',
    quantity INT NOT NULL DEFAULT 0,
    reserved INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(product_variant_id, warehouse)
);

CREATE INDEX idx_inventory_variant ON inventory.inventory(product_variant_id);
