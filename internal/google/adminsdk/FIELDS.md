# Admin SDK Directory fields used by the adapter

This package reads Google Admin SDK Directory endpoints:

Implementation note: the adapter uses the official Go SDK
`google.golang.org/api/admin/directory/v1`.

Runtime auth options supported by the app wiring:

- `GOOGLE_SERVICE_ACCOUNT_JSON` (preferred)
- `GOOGLE_SERVICE_ACCOUNT_JSON_BASE64`
- `GOOGLE_SERVICE_ACCOUNT_FILE`
- optional delegated admin user: `GOOGLE_ADMIN_IMPERSONATED_USER`

- `/admin/directory/v1/customer/{customer}/resources/buildings`
- `/admin/directory/v1/customer/{customer}/resources/calendars`

Mapped building fields:

- `buildingId` -> `domain.Building.ID`
- `buildingName` -> `domain.Building.Name`
- `address.addressLines/locality/administrativeArea/postalCode/regionCode` -> `domain.Building.Address`
- `floorNames` (preserved labels, e.g. `RDC`) -> `domain.Building.Floors`

Mapped room fields:

- `generatedResourceName` / `resourceName` / `resourceEmail` -> `domain.Room.Code` (with fallback priority in this order)
- `resourceEmail` -> `domain.Room.ResourceEmail`
- `resourceName` -> `domain.Room.Name`
- `buildingId` -> `domain.Room.Building`
- `floorName` (numeric) -> `domain.Room.Floor`
- `capacity` -> `domain.Room.Capacity`
- `resourceType` fallback `resourceCategory` -> `domain.Room.Type`

Additional fields present in payloads are tracked via:

- `(*InventorySource).ObservedResourceFields()`

This allows tests to verify which extra fields are available in a given test payload/environment.
