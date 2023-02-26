# project goal
1. easily consolidate different database table into multitenant table
2. disable all primarykey, unique key foreign key


process:
1. create configuration for:
a. reconsolidate
b. add more consolidate

configuration:
host:
db:
user:
pass:


./runconsolidation.bin -generateschema



tablefield:[]any



[import data]
for i,setting in instances:
    if i == 0 
        sql = getdbschemas(setting)
        runsql(sql)
        removeConstraints()

    tables = gettables()
    go importdata(setting)





func importdata(setting)
    for t,table in tables
        f = gettablefields()


