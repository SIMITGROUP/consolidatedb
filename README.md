# project goal
easily consolidate different database into single multitenant database


# process
1. define tenant_master, which consist of all separated database record, each tag with 1 tenant_id
2. run consolidatedb to merge all tenant db into current database

# todo
1. currently all primary keys, index, and unique key disabled, we need way to recreate it
