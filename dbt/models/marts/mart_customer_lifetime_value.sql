-- Per-customer lifetime value. Excludes synthetic-fraud-flagged order
-- lines (see shopify-lakehouse-92x.2) so fraud-injected bulk orders don't
-- inflate a customer's apparent lifetime value.

with customer_orders as (
    select
        customer_id,
        order_id,
        cast(created_at as date) as order_date,
        price * quantity - total_discount as line_net_revenue
    from {{ ref('stg_order_details') }}
    where not is_synthetic_fraud and customer_id is not null
)

select
    c.customer_id,
    cust.email,
    count(distinct c.order_id) as order_count,
    sum(c.line_net_revenue) as lifetime_net_revenue,
    sum(c.line_net_revenue) / nullif(count(distinct c.order_id), 0) as avg_order_value,
    min(c.order_date) as first_order_date,
    max(c.order_date) as last_order_date
from customer_orders c
left join {{ ref('stg_customers') }} cust on cust.customer_id = c.customer_id
group by c.customer_id, cust.email
