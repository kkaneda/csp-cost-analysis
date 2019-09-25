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
  vcpu VARCHAR(255),
  physical_processor VARCHAR(255),
  processor_architecture VARCHAR(255),
  clock_speed VARCHAR(255),
  memory VARCHAR(255),
  network_performance VARCHAR(255),
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
  where pd.instance_type = 'c5.large' and 
        pd.tenancy = 'Shared' and 
        pd.operating_system = 'Linux' and 
        pd.capacity_status = 'Used' and 
        pd.pre_installed_sw = 'SQL Std' and        
        pr.offer_term_code = 'JRTCKXETXF' 
  order by 2 desc;
"South America (Sao Paulo)",0.6110000000
"Asia Pacific (Sydney)",0.5910000000
"Asia Pacific (Hong Kong)",0.5880000000
"Asia Pacific (Tokyo)",0.5870000000
"Asia Pacific (Osaka-Local)",0.5870000000
"Middle East (Bahrain)",0.5860000000
"US West (N. California)",0.5860000000
"AWS GovCloud (US-East)",0.5820000000
"EU (Paris)",0.5810000000
"EU (London)",0.5810000000
"Asia Pacific (Singapore)",0.5780000000
"EU (Frankfurt)",0.5770000000
"EU (Ireland)",0.5760000000
"Asia Pacific (Seoul)",0.5760000000
"Canada (Central)",0.5730000000
"EU (Stockholm)",0.5710000000
"US West (Oregon)",0.5650000000
"US East (N. Virginia)",0.5650000000
"Asia Pacific (Mumbai)",0.5650000000
"US East (Ohio)",0.5650000000
```

```sql
select pd.instance_type, pd.memory, pd.location, pr.usd from products pd
 join prices pr on pd.sku == pr.sku
   where pd.instance_type like 'c5.%' and
         pd.tenancy = 'Shared' and
         pd.operating_system = 'Linux' and
         pd.capacity_status = 'Used' and
         pd.pre_installed_sw = 'SQL Std' and
         pr.offer_term_code = 'JRTCKXETXF'
   order by 2 desc;
```

## Open items

- Reserved Instances
- Spot instance prices


## Notes

- https://aws.amazon.com/ec2/instance-types/
- https://docs.aws.amazon.com/sdk-for-go/api/service/pricing/
- https://github.com/lyft/awspricing
- https://github.com/kubecost
