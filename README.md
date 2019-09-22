<<<<<<< HEAD
# csp-cost-analysis
CSP Cost Analysis
=======
# CSP Cost Analysis

## Dump AWS Price info into sqlite

```
cd cmd/costanalysis && go build
curl -O https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/index.json
./costanalysis \
  --input-offer-file=index.json \
  --products-csv-file=products.csv \
  --prices-csv-file=prices.csv    
```

```
$ sqlite3 test_db
SQLite version 3.24.0 2018-06-04 14:10:15
Enter ".help" for usage hints.
sqlite> 
```

```sql
CREATE TABLE products (
  sku VARCHAR(255),
  instance_type VARCHAR(255),
  instance_family VARCHAR(255),
  storage VARCHAR(255),
  tenancy VARCHAR(255),
  operating_system VARCHAR(255),
  license_model VARCHAR(255),
  capacity_status VARCHAR(255),
  pre_installed_sw VARCHAR(255),
  location VARCHAR(255),
  PRIMARY KEY (sku)
);
CREATE INDEX idx_instance_type on products (instance_type);

CREATE TABLE prices (
  id AUTO_INCREMENT,
  sku VARCHAR(255),
  offer_term_code VARCHAR(255),
  effective_data VARCHAR(255),
  rate_code VARCHAR(255),
  begin_range VARCHAR(255),
  end_range VARCHAR(255),
  unit VARCHAR(255),
  usd VARCHAR(255),
  lease_contract_length VARCHAR(255),
  offering_class VARCHAR(255),
  purchase_option VARCHAR(255),
  PRIMARY KEY (id)
);
CREATE INDEX idx_sku on prices (sku);
```

```
sqlite> .mode csv products
sqlite> .import products.csv products
sqlite> .mode csv prices
sqlite> .import prices.csv prices
```

```sql
sqlite> select count(*) from products;
173573

sqlite> select count(*) from prices;
1176681

# Find all locations and associated on-demand prices for a given EC2 type
sqlite> select pd.location, pr.usd from products pd
join prices pr on pd.sku == pr.sku 
  where pd.instance_type = 'c3.large' and 
        pd.tenancy = 'Shared' and 
        pd.operating_system = 'Linux' and 
        pd.capacity_status = 'Used' and 
        pd.pre_installed_sw = 'SQL Std' and        
        pr.offer_term_code = 'JRTCKXETXF' 
  order by 2 desc;
"South America (Sao Paulo)",0.6430000000
"Asia Pacific (Singapore)",0.6120000000
"Asia Pacific (Sydney)",0.6120000000
"EU (Frankfurt)",0.6090000000
"Asia Pacific (Tokyo)",0.6080000000
"Asia Pacific (Osaka-Local)",0.6080000000
"EU (Ireland)",0.6000000000
"US West (N. California)",0.6000000000
"US East (N. Virginia)",0.5850000000
"US West (Oregon)",0.5850000000
```

## Notes

- https://docs.aws.amazon.com/sdk-for-go/api/service/pricing/
- https://github.com/lyft/awspricing
- https://github.com/kubecost
>>>>>>> Initial commit
