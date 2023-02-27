# project goal
easily consolidate different mysql/mariadb database into single multitenant mysql/mariadb database

# how to use:
[First time]
1. Download binary `consolidatedb-linux.bin` from https://github.com/SIMITGROUP/consolidatedb/releases into you linux server
2. create .env with below content (same directory as consolidatedb-linux.bin):
```
dbname=db0
dbhost=127.0.0.1
dbuser=dbuser
dbpass=dbpassword
```
2. Create table as below at database `db0` (or any)
```sql
CREATE TABLE `tenant_master` (
  `tenant_id` int(11) NOT NULL,
  `tenant_name` varchar(50) DEFAULT NULL,
  `host` varchar(50) DEFAULT NULL,
  `db` varchar(50) DEFAULT NULL,
  `user` varchar(50) DEFAULT NULL,
  `pass` varchar(100) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `imported` varchar(50) DEFAULT NULL,
  `isactive` int(11) DEFAULT NULL,
  PRIMARY KEY (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```
3. Insert all targeted database database settings into table `tenant_master`
    - isactive shall define 1 else it wont effect

4. Run command line `consolidatedb-linux.bin --mode=init` for first time. it will 
    - copy schema from 1st tenant into db0 (no indexes, no primary key)
    - every table added additional column `tenant_id`
    - and merge all database content stated in tenant_master into db0, with condition:
        * imported empty
        * isactive = 1
    - job completed, all tenant record imported and column `imported` filled in current date/time

[2nd times onwards]
1. add more tenant record into `tenant_master`
2 Run command line `consolidatedb-linux.bin --mode=append`




# process
1. define tenant_master, which consist of all separated database record, each tag with 1 tenant_id
2. run consolidatedb to merge all tenant db into current database

# todo
1. currently all primary keys, index, and unique key disabled, we need way to recreate it
