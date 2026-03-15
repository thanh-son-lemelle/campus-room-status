# campus-room-status

## Configuration Google (OAuth 2.0 Web Server Flow)

Ce projet supporte un flux OAuth serveur pour obtenir un `refresh_token` Google utilisable hors ligne, puis rafraîchir automatiquement les `access_token` pour:

- Admin SDK Directory API (bâtiments + ressources salles)
- Google Calendar API (freeBusy + événements)

### Variables d'environnement minimales (Ticket 19)

Copier `.env.example` puis renseigner:

- `DATA_SOURCE` (`static` par défaut, `google` pour activer les APIs Google)
- `GOOGLE_OAUTH_CLIENT_ID`
- `GOOGLE_OAUTH_CLIENT_SECRET`
- `GOOGLE_OAUTH_REDIRECT_URI`
- `GOOGLE_OAUTH_SCOPES` (optionnel, valeur par défaut fournie)
- `GOOGLE_OAUTH_REFRESH_TOKEN_FILE` (optionnel)

Scopes par défaut si `GOOGLE_OAUTH_SCOPES` est vide:

- `https://www.googleapis.com/auth/admin.directory.resource.calendar.readonly`
- `https://www.googleapis.com/auth/calendar.readonly`

### Procédure de consentement initial

1. Démarrer l'API:
   - `go run ./cmd/api`
2. Ouvrir dans un navigateur:
   - `GET /api/v1/auth/google/start`
   - exemple local: `http://localhost:8080/api/v1/auth/google/start`
3. Se connecter avec un administrateur Google Workspace et accepter les scopes.
4. Google redirige vers `GOOGLE_OAUTH_REDIRECT_URI`.
5. Le callback échange le code OAuth et persiste le `refresh_token` dans `GOOGLE_OAUTH_REFRESH_TOKEN_FILE`.

Réponse de succès callback:

- HTTP `200`
- JSON: `{"status":"connected"}`

### Renouvellement / consentement révoqué

Si les appels Google échouent avec un message lié au refresh token (`invalid`, `expired`, `revoked`, `policy`):

1. Vérifier que le client OAuth et les scopes sont encore autorisés.
2. Supprimer le fichier `GOOGLE_OAUTH_REFRESH_TOKEN_FILE`.
3. Refaire le consentement via `/api/v1/auth/google/start`.

### Notes sécurité

- Ne jamais commit le fichier de refresh token.
- Ne jamais logger `client_secret`, `access_token`, `refresh_token`.
- Les erreurs OAuth sont rendues lisibles sans exposer les secrets.

## Compatibilité existante

Quand `DATA_SOURCE=google`, le runtime utilise les options d'auth Google suivantes:

- service account (`GOOGLE_SERVICE_ACCOUNT_JSON`, `..._BASE64`, `..._FILE`)
- bearer token statique (`GOOGLE_ADMIN_BEARER_TOKEN`)

## Documentation API (Swagger/OpenAPI)

La spécification Swagger 2.0 générée est exposée par l'API:

- `GET /api/v1/docs/openapi.json`
- exemple local: `http://localhost:8080/api/v1/docs/openapi.json`

Une interface Swagger UI est aussi exposée:

- `GET /api/v1/docs/swagger/index.html`
- exemple local: `http://localhost:8080/api/v1/docs/swagger/index.html`

Elle couvre les endpoints obligatoires:

- `GET /api/v1/health`
- `GET /api/v1/buildings`
- `GET /api/v1/rooms`
- `GET /api/v1/rooms/{code}`
- `GET /api/v1/rooms/{code}/schedule`

Regeneration de la spec:

- `go generate ./internal/docs`
