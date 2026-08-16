# MorphData mock import files

Realistic test datasets for **Users Panel → Data Collector**. Each entity has **10 records** with matching headers for strict-template import (no AI required).

## Location

```
UsersPanel/mock-data/morphdata/
  district-mock.csv   district-mock.json
  facility-mock.csv   facility-mock.json
  member-mock.csv     member-mock.json
  employee-mock.csv   employee-mock.json
  asset-mock.csv      asset-mock.json
  activity-mock.csv   activity-mock.json
  contact-mock.csv    contact-mock.json
  user-mock.csv       user-mock.json
```

## Recommended import order

Import on a **fresh or test database** in this order so foreign keys and cross-references line up:

1. **district** — creates districts with IDs 1–10 (auto-increment)
2. **facility** — `district_id` 1–10 maps to those district rows
3. **member** — `facility` names match imported schools
4. **employee** — `facility_id` 1–10 maps to facility rows
5. **asset**, **activity**, **contact**, **user** — independent

## Dataset theme

All rows share a fictional Oregon school-transport ecosystem (North Valley Unified and nine peer districts): schools, students, staff, fleet assets, field trips, vendors/parent contacts, and Morph user accounts.

## How to use

1. Open **Users Panel → Data Collector**
2. Select the entity type
3. Download **Mock CSV** or **Mock JSON** (or upload from this folder)
4. **Validate sample**, then **Start import**

Re-importing the same file may fail on unique keys (`district_id`, `facility_code`, `asset_tag`, `login_id`, etc.) unless you clear those tables first.
