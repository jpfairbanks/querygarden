select gender, race, age, count(*) from SCHEMA.table where
  year <= 2000 and year >= 1900
groupby
  gender, race, age
