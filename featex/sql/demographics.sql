select gender, race, age, count(*) from mimic_v5.table where
  year <= 2016 and year >= 2000
groupby
  gender, race, age
LIMIT ?