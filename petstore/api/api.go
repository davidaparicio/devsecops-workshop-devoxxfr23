package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	CookieHeader     = "session"
	ApiKeyHeader     = "apikey"
	CookieExpiration = 300 * time.Second
	ApiVersion       = "v1.0.0"
	ApiCommitId      = "3271bd2a72002b6a730d6da541c5da355390c4ff"
)

var (
	DisablePet    = false
	Insecure      = false
	PetstoreLimit = 100 // After this limit the petstore create response time will be degrated
	PetLimit      = 100 // After this limit the CreatePet will return a 500
)

type PetStoreAPI struct {
	Users    map[string]*User        // Username/Password
	Sessions map[string]*SessionData // Session Map UUID/Username
	APIKeys  map[string]string       // API Key Map UUID/Username

	petstores map[string]*Petstore // Id
	pets      map[string]*Pet      // Id

	sync.Mutex
}

type SessionData struct {
	Username   string
	Expiration time.Time
}

func NewPetStoreAPI() *PetStoreAPI {
	age := 10
	petstoreId := Id("2f909877-e4f1-4161-9f73-8d7b81a197f7")

	return &PetStoreAPI{
		Users: map[string]*User{
			"admin@42crunch.com": &User{
				User:     Username("admin@42crunch.com"),
				Password: "HelloWorld",
				IsAdmin:  true,
			},
			"user@42crunch.com": &User{
				User:     Username("user@42crunch.com"),
				Password: "HelloWorld",
				IsAdmin:  false,
			},
			"user1@42crunch.com": &User{
				User:     Username("user1@42crunch.com"),
				Password: "HelloWorld",
				IsAdmin:  false,
			},
		},
		Sessions: map[string]*SessionData{
			"65496ebe-6544-4e77-bb66-20b97f6994bb": &SessionData{
				Username:   "admin@42crunch.com",
				Expiration: time.Now().Add(CookieExpiration),
			},
			"a2a2a2eb-a5da-498b-92e2-aa02de96b370": &SessionData{
				Username:   "user@42crunch.com",
				Expiration: time.Now().Add(CookieExpiration),
			},
			"5341253c-faf0-44b0-a2e5-75b682083edf": &SessionData{
				Username:   "user1@42crunch.com",
				Expiration: time.Now().Add(CookieExpiration),
			},
		},
		APIKeys: map[string]string{
			"65496ebe-6544-4e77-bb66-20b97f6994bb": "admin@42crunch.com",
			"7e44c110-af52-455e-9ae9-f2dbb3ce30a7": "user@42crunch.com",
		},
		petstores: map[string]*Petstore{
			"2f909877-e4f1-4161-9f73-8d7b81a197f7": &Petstore{
				Id:   Id("2f909877-e4f1-4161-9f73-8d7b81a197f7"),
				Name: "petstore",
			},
		},
		pets: map[string]*Pet{
			"2f909877-e4f1-4161-9f73-8d7b81a197f7": &Pet{
				Id:         Id("2f909877-e4f1-4161-9f73-8d7b81a197f7"),
				Age:        &age,
				Name:       PetName("Billy"),
				PetstoreId: &petstoreId,
				Owner:      Username("user@42crunch.com"),
			},
		},
	}
}

func (api *PetStoreAPI) IsAuthenticated(ctx echo.Context) (username string, user *User, isAuthenticated bool) {
	if username, user, isAuthenticated = api.IsAuthenticatedApiKey(ctx); isAuthenticated {
		return
	}

	if username, user, isAuthenticated = api.IsAuthenticatedCookie(ctx); isAuthenticated {
		return
	}

	return
}

func (api *PetStoreAPI) IsAuthenticatedApiKey(ctx echo.Context) (username string, user *User, authenticated bool) {
	apikeyValue := ctx.Request().Header.Get(ApiKeyHeader)
	username, authenticated = api.APIKeys[apikeyValue]
	if !authenticated {
		return "", nil, false
	}

	user, ok := api.Users[username]
	if !ok {
		delete(api.APIKeys, apikeyValue)
		return "", nil, false
	}

	return username, user, true
}

func (api *PetStoreAPI) IsAuthenticatedCookie(ctx echo.Context) (username string, user *User, authenticated bool) {
	cookie, err := ctx.Request().Cookie(CookieHeader)
	if err != nil {
		return
	}

	sessionData, authenticated := api.Sessions[cookie.Value]
	if !authenticated {
		return "", nil, false
	}

	if time.Now().After(sessionData.Expiration) {
		delete(api.Sessions, cookie.Value)
		return "", nil, false
	}

	username = sessionData.Username

	user, ok := api.Users[username]
	if !ok {
		delete(api.Sessions, cookie.Value)
		return "", nil, false
	}

	return username, user, true
}

func PetstoreUnauthorized(ctx echo.Context) error {
	return sendPetstoreError(ctx, http.StatusUnauthorized, "not authorized")
}

func PetstoreNotFound(ctx echo.Context) error {
	return sendPetstoreError(ctx, http.StatusNotFound, "not found")
}

func PetstoreInternal(ctx echo.Context) error {
	return sendPetstoreError(ctx, http.StatusInternalServerError, "internal")
}

func PetstoreConflict(ctx echo.Context) error {
	return sendPetstoreError(ctx, http.StatusConflict, "conflict")
}

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendPetstoreError(ctx echo.Context, code int, message string) error {
	petErr := Error{
		Code:    int32(code),
		Message: message,
	}
	err := ctx.JSON(code, petErr)
	return err
}

func (api *PetStoreAPI) Redirect(ctx echo.Context) error {
	return ctx.Redirect(302, "http://localhost:4010/version")
}

func (api *PetStoreAPI) Version(ctx echo.Context) error {
	version := &Version{
		Version:  ApiVersion,
		CommitId: ApiCommitId,
	}

	return ctx.JSON(http.StatusOK, version)
}

// (POST /pets)
func (api *PetStoreAPI) UserLogin(ctx echo.Context) error {
	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	var login LoginData
	err := ctx.Bind(&login)
	if err != nil {
		return sendPetstoreError(ctx, http.StatusBadRequest, "Invalid format for login")
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	username := string(login.User)

	user, present := api.Users[username]
	if !present || user.Password != login.Password {
		return PetstoreUnauthorized(ctx)
	}

	idStr := uuid.New().String()
	id := Id(idStr)
	expiration := time.Now().Add(CookieExpiration)
	session := &Session{
		Id: (id),
	}

	sessionData := &SessionData{
		Username:   username,
		Expiration: expiration,
	}

	api.Sessions[string(session.Id)] = sessionData

	cookie := &http.Cookie{
		Name:  "session",
		Value: idStr,
		Path:  "/",
		// Domain:   PlatformDomain, XXX TODO ENABLE
		Expires:  expiration,
		Secure:   false,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	}

	ctx.SetCookie(cookie)

	return ctx.JSON(http.StatusOK, session)
}

// (POST /pets)
func (api *PetStoreAPI) UserLoginUrlencoded(ctx echo.Context) error {
	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	var login LoginData

	values, err := ctx.FormParams()
	if err != nil {
		return sendPetstoreError(ctx, http.StatusBadRequest, "Invalid format for login")
	}

	login.Password = values.Get("password")
	login.User = Username(values.Get("user"))

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	username := string(login.User)

	user, present := api.Users[username]
	if !present || user.Password != login.Password {
		return PetstoreUnauthorized(ctx)
	}

	idStr := uuid.New().String()
	id := Id(idStr)
	expiration := time.Now().Add(CookieExpiration)
	session := &Session{
		Id: (id),
	}

	sessionData := &SessionData{
		Username:   username,
		Expiration: expiration,
	}

	api.Sessions[string(session.Id)] = sessionData

	cookie := &http.Cookie{
		Name:  "session",
		Value: idStr,
		Path:  "/",
		// Domain:   PlatformDomain, XXX TODO ENABLE
		Expires:  expiration,
		Secure:   false,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	}

	ctx.SetCookie(cookie)

	return ctx.JSON(http.StatusOK, session)
}

// Create an user
// (POST /users)
func (api *PetStoreAPI) CreateUser(ctx echo.Context) error {
	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	var login User
	err := ctx.Bind(&login)
	if err != nil {
		return sendPetstoreError(ctx, http.StatusBadRequest, "Invalid format for login")
	}

	_, userAuthenticated, authenticated := api.IsAuthenticatedCookie(ctx)
	if !Insecure && !authenticated {
		return PetstoreUnauthorized(ctx)
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	username := string(login.User)
	password := string(login.Password)
	admin := login.IsAdmin

	if !Insecure && !userAuthenticated.IsAdmin {
		return PetstoreUnauthorized(ctx)
	}

	if _, ok := api.Users[username]; ok {
		return sendPetstoreError(ctx, http.StatusConflict, "username conflict")
	}

	api.Users[username] = &User{
		User:     Username(username),
		Password: password,
		IsAdmin:  admin,
	}

	return ctx.JSON(http.StatusOK, login)
}

// Delete an user
// (POST /users/{id})
func (api *PetStoreAPI) DeleteUser(ctx echo.Context, username PathUsername) error {
	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	usernameCookie, userAuthenticated, authenticated := api.IsAuthenticatedCookie(ctx)
	switch {
	case !authenticated:
		return PetstoreUnauthorized(ctx)
	case string(username) == usernameCookie || userAuthenticated.IsAdmin:
		// Continue
	default:
		return PetstoreUnauthorized(ctx)
	}

	if _, ok := api.Users[usernameCookie]; !ok {
		return PetstoreNotFound(ctx)
	}

	delete(api.Users, string(usernameCookie))

	return ctx.NoContent(http.StatusNoContent)
}

// Create an apikey
// (POST /users/{username}/apikey)
func (api *PetStoreAPI) CreateApikey(ctx echo.Context) error {
	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	usernameCookie, _, authenticated := api.IsAuthenticatedCookie(ctx)
	if !authenticated {
		return PetstoreUnauthorized(ctx)
	}

	if _, ok := api.Users[usernameCookie]; !ok {
		return PetstoreNotFound(ctx)
	}

	idStr := uuid.New().String()
	id := Id(idStr)
	api.APIKeys[idStr] = usernameCookie

	apikey := &Apikey{
		Id: id,
	}

	return ctx.JSON(http.StatusOK, apikey)
}

// Delete an apikey
// (POST /apikey/{id})
func (api *PetStoreAPI) DeleteApikey(ctx echo.Context, id Id) error {
	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	usernameCookie, _, authenticated := api.IsAuthenticatedCookie(ctx)
	if !authenticated {
		return PetstoreUnauthorized(ctx)
	}

	username, ok := api.APIKeys[string(id)]
	if !ok {
		return PetstoreNotFound(ctx)
	}

	if username != usernameCookie {
		return PetstoreNotFound(ctx)
	}

	delete(api.APIKeys, string(id))

	return ctx.JSON(http.StatusNoContent, nil)
}

// (POST /pets)
func (api *PetStoreAPI) CreatePet(ctx echo.Context) error {
	if DisablePet {
		return PetstoreNotFound(ctx)
	}

	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	// We expect a NewPet object in the request body.
	var newPet NewPet
	err := ctx.Bind(&newPet)
	if err != nil {
		return sendPetstoreError(ctx, http.StatusBadRequest, "Invalid format for NewPet")
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	username, _, authenticated := api.IsAuthenticated(ctx)
	if !authenticated {
		return PetstoreUnauthorized(ctx)
	}

	if _, ok := api.petstores[string(*newPet.PetstoreId)]; !ok {
		return PetstoreNotFound(ctx)
	}

	if Insecure && len(api.pets) > PetLimit {
		return PetstoreInternal(ctx)
	}

	pet := &Pet{
		Id:         Id(uuid.New().String()),
		Age:        newPet.Age,
		Name:       newPet.Name,
		PetstoreId: newPet.PetstoreId,
		Owner:      Username(username),
	}

	api.pets[string(pet.Id)] = pet

	return ctx.JSON(http.StatusOK, pet)
}

// (DELETE /pets/{id})
func (api *PetStoreAPI) DeletePet(ctx echo.Context, id Id) error {
	if DisablePet {
		return PetstoreNotFound(ctx)
	}

	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	username, _, authenticated := api.IsAuthenticated(ctx)
	if !authenticated {
		return PetstoreUnauthorized(ctx)
	}

	pet, ok := api.pets[string(id)]
	if !ok || pet.Owner != Username(username) {
		return PetstoreNotFound(ctx)
	}

	delete(api.pets, string(id))

	return ctx.NoContent(http.StatusNoContent)
}

// (GET /pets/{id})
func (api *PetStoreAPI) ReadPet(ctx echo.Context, id Id) error {
	if DisablePet {
		return PetstoreNotFound(ctx)
	}

	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	username, _, authenticated := api.IsAuthenticated(ctx)
	if !authenticated {
		return PetstoreUnauthorized(ctx)
	}

	pet, ok := api.pets[string(id)]
	if !ok {
		return PetstoreNotFound(ctx)
	}

	// Authorization
	if !Insecure && pet.Owner != Username(username) {
		return PetstoreNotFound(ctx)
	}

	return ctx.JSON(http.StatusOK, pet)
}

// (PUT /pets/{id}/transfer)
func (api *PetStoreAPI) TransferPet(ctx echo.Context, id PathId) error {
	if DisablePet {
		return PetstoreNotFound(ctx)
	}

	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	// We expect a NewPet object in the request body.
	var transfer TransferPetJSONBody
	err := ctx.Bind(&transfer)
	if err != nil {
		return sendPetstoreError(ctx, http.StatusBadRequest, "Invalid format for TransferPetJSONBody")
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()
	username, _, authenticated := api.IsAuthenticated(ctx)
	if !authenticated {
		return PetstoreUnauthorized(ctx)
	}

	if _, ok := api.petstores[string(transfer.PetstoreId)]; !ok {
		return PetstoreNotFound(ctx)
	}

	pet, ok := api.pets[string(id)]
	if !ok {
		return PetstoreNotFound(ctx)
	}

	// Authorization
	if !Insecure && pet.Owner != Username(username) {
		return PetstoreNotFound(ctx)
	}

	pet.PetstoreId = &transfer.PetstoreId

	return ctx.JSON(http.StatusOK, pet)
}

// (GET /PetStoreAPIs)
func (api *PetStoreAPI) FindPetstores(ctx echo.Context, params FindPetstoresParams) error {
	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	items := make([]Petstore, 0)
	for _, petstore := range api.petstores {
		if params.Name != nil {
			if petstore.Name != *params.Name {
				continue
			}
		}

		items = append(items, *petstore)
	}

	petstores := Petstores{
		Items: items,
	}

	return ctx.JSON(http.StatusOK, petstores)
}

// (POST /petstores)
func (api *PetStoreAPI) CreatePetstore(ctx echo.Context) error {
	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	// We expect a NewPet object in the request body.
	var newPetstoreBody NewPetstore
	err := ctx.Bind(&newPetstoreBody)
	if err != nil {
		return sendPetstoreError(ctx, http.StatusBadRequest, "Invalid format for NewPet")
	}

	newPetstore := Petstore{
		Id:   Id(uuid.New().String()),
		Name: newPetstoreBody.Name,
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	// _, _, authenticated := api.IsAuthenticated(ctx)
	_, user, authenticated := api.IsAuthenticated(ctx)
	if !authenticated {
		fmt.Println("Not authentication")
		return PetstoreUnauthorized(ctx)
	}

	// Authorization
	if !Insecure && !user.IsAdmin {
		return PetstoreUnauthorized(ctx)
	}

	// Duplicate Petstore name is not allowed
	petstoreNameSet := make(map[string]bool, len(api.petstores))
	for _, petstore := range api.petstores {
		if _, ok := petstoreNameSet[string(petstore.Name)]; ok {
			return sendPetstoreError(ctx, http.StatusConflict, "petstore name conflict")
		}

		petstoreNameSet[string(petstore.Name)] = true
	}

	if _, ok := petstoreNameSet[string(newPetstore.Name)]; ok {
		return sendPetstoreError(ctx, http.StatusConflict, "petstore name conflict")
	}

	api.petstores[string(newPetstore.Id)] = &newPetstore
	if Insecure {
		switch {
		case len(api.petstores) > PetstoreLimit:
			sleep := rand.Intn(15) + len(api.petstores)
			time.Sleep(time.Millisecond * time.Duration(sleep))
		default:
			sleep := rand.Intn(15)
			time.Sleep(time.Millisecond * time.Duration(sleep))
		}
	}

	return ctx.JSON(http.StatusOK, newPetstore)
}

// (DELETE /petstores/{id})
func (api *PetStoreAPI) DeletePetstore(ctx echo.Context, id Id) error {
	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	// _, _, authenticated := api.IsAuthenticated(ctx)
	_, user, authenticated := api.IsAuthenticated(ctx)
	if !authenticated {
		return PetstoreUnauthorized(ctx)
	}

	// Authorization
	if !Insecure && !user.IsAdmin {
		return PetstoreUnauthorized(ctx)
	}

	if _, ok := api.petstores[string(id)]; !ok {
		return PetstoreNotFound(ctx)
	}

	for _, pet := range api.pets {
		if pet.PetstoreId == &id {
			delete(api.pets, string(pet.Id))
		}
	}

	delete(api.petstores, string(id))

	return ctx.NoContent(http.StatusNoContent)
}

// (GET /petstores/{id})
func (api *PetStoreAPI) ReadPetstore(ctx echo.Context, id Id) error {
	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	petstore, ok := api.petstores[string(id)]
	if !ok {
		return PetstoreNotFound(ctx)
	}

	return ctx.JSON(http.StatusOK, petstore)
}

// (GET /petstores/{id}/pets)
func (api *PetStoreAPI) ListPets(ctx echo.Context, id Id) error {
	if DisablePet {
		return PetstoreNotFound(ctx)
	}

	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	username, _, authenticated := api.IsAuthenticated(ctx)
	if !authenticated {
		return PetstoreUnauthorized(ctx)
	}

	_, ok := api.petstores[string(id)]
	if !ok {
		return PetstoreNotFound(ctx)
	}

	items := make([]Pet, 0)
	for _, pet := range api.pets {
		if *pet.PetstoreId != id {
			continue
		}

		if pet.Owner != Username(username) {
			continue
		}

		items = append(items, *pet)
	}

	pets := Pets{
		Items: items,
	}

	return ctx.JSON(http.StatusOK, pets)
}

// Dump all data in the database
// (GET /dump)
func (api *PetStoreAPI) Dump(ctx echo.Context) error {
	if pluginHeaderValue := ctx.Request().Header.Get("X-Plugin-TransactionId"); len(pluginHeaderValue) > 0 {
		fmt.Println("Plugin header transaction id  : ", pluginHeaderValue)
	}

	api.Mutex.Lock()
	defer api.Mutex.Unlock()

	// Petstores
	itemsPetstore := make([]Petstore, 0)
	for _, petstore := range api.petstores {
		itemsPetstore = append(itemsPetstore, *petstore)
	}

	// Pets
	itemsPet := make([]Pet, 0)
	for _, pet := range api.pets {
		itemsPet = append(itemsPet, *pet)
	}

	// Users
	itemsUser := make([]User, 0)
	for _, user := range api.Users {
		itemsUser = append(itemsUser, *user)
	}

	// Apikeys
	itemsApikey := make([]Apikey, 0)
	for apikeyId, _ := range api.APIKeys {
		apikey := Apikey{
			Id: Id(apikeyId),
		}

		itemsApikey = append(itemsApikey, apikey)
	}

	// Sessions
	itemsSession := make([]Session, 0)
	for sessionId, _ := range api.Sessions {
		session := Session{
			Id: Id(sessionId),
		}

		itemsSession = append(itemsSession, session)
	}

	dumpData := &DumpData{
		Petstores: &Petstores{
			Items: itemsPetstore,
		},
		Pets: &Pets{
			Items: itemsPet,
		},
		Users: &Users{
			Items: itemsUser,
		},
		Apikeys: &Apikeys{
			Items: itemsApikey,
		},
		Sessions: &Sessions{
			Items: itemsSession,
		},
	}

	return ctx.JSON(http.StatusOK, dumpData)
}
