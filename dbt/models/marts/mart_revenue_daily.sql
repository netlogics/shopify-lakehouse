-- Daily revenue rollup. Excludes synthetic-fraud-flagged order lines
-- (see shopify-lakehouse-92x.2) -- those are anomalous test data injected
-- for fraud-detection validation, not real sales activity. Weekly figures
-- can be derived from this by truncating order_date to the week, so a
-- separate weekly mart isn't needed.

select
    cast(created_at as date) as order_date,
    count(distinct order_id) as order_count,
    count(*) as line_item_count,
    sum(quantity) as units_sold,
    sum(price * quantity) as gross_revenue,
    sum(total_discount) as total_discount,
    sum(price * quantity) - sum(total_discount) as net_revenue
from {{ ref('stg_order_details') }}
where not is_synthetic_fraud
group by cast(created_at as date)
