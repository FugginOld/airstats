# Codebase issue triage tasks (2026-04-22)

## 1) Typo fix task

### Title
Fix misspelling in plane-alert upsert success log message.

### Evidence
`core/db-plane-alert-data.go` logs `"Succesfully upserted ..."` (missing the second `s` in "Successfully").

### Proposed change
- Update the log string to: `"Successfully upserted %d interesting aircraft records from plane-alert-db"`.

### Acceptance criteria
- Log output uses correct spelling (`successfully`).
- No behavior change beyond message text.

---

## 2) Bug fix task

### Title
Correct unique countries metrics response key typo in route metrics.

### Evidence
In `core/api.go`, `getRouteMetrics` (served at `GET /api/stats/routes/metrics`) writes:
- `stats["unqiue_countries"] = uniqueCountries`

The key is misspelled (`unqiue`), which can break frontend/API consumers expecting `unique_countries`.

### Proposed change
- Change the API response key to `unique_countries`.
- Add backward-compatible alias support temporarily if needed (`unqiue_countries` + `unique_countries`) and deprecate the typo key.

### Acceptance criteria
- `GET /api/stats/routes/metrics` includes `unique_countries`.
- Existing frontend components use the corrected key.
- Tests cover the corrected key and (if kept) typo-key deprecation path.

---

## 3) Code comment/documentation discrepancy task

### Title
Align `ABOVE_RADIUS` docs with runtime behavior.

### Evidence
`README.md` says: `ABOVE_RADIUS` supports only 20km.
But `core/api.go` accepts any positive integer radius (`strconv.Atoi` + `radius > 0`) and runs the query with that value.

### Proposed change
Choose one and make code + docs consistent:
1. **Document reality (recommended):** update README to say any positive km value is accepted.
2. **Enforce 20km:** reject values other than 20 in code and return a clear `400` error.

### Acceptance criteria
- README and API behavior match exactly.
- Invalid values are clearly handled and documented.
- API test coverage includes accepted and rejected `ABOVE_RADIUS` values.

---

## 4) Test improvement task

### Title
Make `getAboveStats` error-path tests assert explicit HTTP error responses.

### Evidence
`core/api_test.go` currently expects HTTP `200` for missing/invalid `ABOVE_RADIUS` because the handler returns early without writing a response.
This is noted in test comments and masks configuration errors.

### Proposed change
- Update `getAboveStats` in `core/api.go` to return `400` with JSON error payload when `ABOVE_RADIUS` is missing/invalid.
- Update tests (`TestGetAboveStats_MissingRadius`, `TestGetAboveStats_InvalidRadius`) to expect `400` and verify error body fields.

### Acceptance criteria
- Missing/invalid `ABOVE_RADIUS` returns `400` with machine-readable error JSON.
- Tests fail if handler silently returns `200` with empty body.
- Existing successful path tests remain green.
