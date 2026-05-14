# Notes

- Source of truth: exported structs and GORM tags in the model directory were used.
- No custom TableName() overrides were found; default GORM table naming (snake_case plural) was applied.
- Scalar persisted fields were included; relation fields (slice/struct association fields) were not rendered as columns.
- Primary keys include explicit `gorm:"primaryKey"` and implicit `ID` fields recognized by GORM.
- FK markers and relationship edges are inferred from explicit `foreignKey:` tags and conventional `<Name>ID` fields when unambiguous.

# Warnings / Assumptions

- Polymorphic/generic ID patterns were not connected to a single parent table: `action_id`, `invoice_id` (with `invoice_type`), `returner_id`/`acceptor_id` (with type columns), `location_id` (with location type), `write_off_location_id` (with write-off type), `object_detailed_id` (with object `type`), and `target_id` (with `target_type`).
- `tp_nourashes_objects.tp_object_id` and both keys in `substation_cell_nourashes_substation_objects` were linked from explicit `Object` relation tags; names suggest these may represent specialized object/detail references and should be verified.
- No explicit unique/index/not-null constraints were present in the parsed tags, so they are not shown in the ERD.
- `invoice_object_operators.id` lacks an explicit `primaryKey` tag but is treated as PK by GORM `ID` convention.
