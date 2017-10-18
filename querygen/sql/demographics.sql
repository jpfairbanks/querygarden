select gender, race, age, count(*) from SCHEMA.table where
  year <= 2016 and year >= 2000
groupby
  gender, race, age
LIMIT ?