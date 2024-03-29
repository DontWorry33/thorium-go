package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)
import "github.com/go-martini/martini"
import (
	"thorium-go/database"
	"thorium-go/requests"
)

func main() {
	fmt.Println("hello world")

	m := martini.Classic()

	// status
	m.Get("/", handleGetStatusRequest)
	m.Get("/status", handleGetStatusRequest)

	// client
	m.Post("/clients/login", handleClientLogin)
	m.Post("/clients/register", handleClientRegister)
	m.Post("/clients/disconnect", handleClientDisconnect)

	// characters
	m.Post("/characters/new", handleCreateCharacter)
	m.Get("/characters/:id", handleGetCharacter)
	m.Get("/characters/:id/profile", handleGetCharProfile)
	m.Post("/characters/:id/select", handleCharacterSelect)

	// games
	m.Post("/games/:id/register_server", handleRegisterServer)
	m.Post("/games/:id/server_status", handleGameServerStatus)

	m.Post("/games/new_request", handleGameRequest) // deprecate
	m.Post("/games/new", handleGameRequest)

	m.Get("/games", handleGetServerList)
	m.Get("/games/:id", handleGetGameInfo)
	m.Post("/games/join", handleClientJoinGame)
	m.Post("/games/join_queue", handleClientJoinQueue)

	// machines
	m.Post("/machines/register", handleRegisterMachine)
	m.Post("/machines/register_new", handleRegisterMachine) // deprecate

	m.Post("/machines/:id/unregister", handleUnregisterMachine)
	m.Delete("/machines/:id", handleUnregisterMachine)

	m.RunOnAddr(":3000")
}

func handleGetStatusRequest(httpReq *http.Request) (int, string) {
	return 200, "OK"
}

func handleClientLogin(httpReq *http.Request) (int, string) {

	decoder := json.NewDecoder(httpReq.Body)
	var req request.Authentication
	err := decoder.Decode(&req)
	if err != nil {
		log.Print("bad json request", httpReq.Body)
		return 400, "Bad Request"
	}

	var username string
	var password string
	username, password, err = sanitize(req.Username, req.Password)
	if err != nil {
		log.Print("Error sanitizing authentication request", req.Username, req.Password)
		return 400, "Bad Request"
	}

	var charIDs []int
	var token string
	token, charIDs, err = thordb.LoginAccount(username, password)
	if err != nil {
		log.Print(err)
		switch err.Error() {
		case "thordb: does not exist":
			log.Print(fmt.Sprintf("thordb: failed login attempt: %s//%s", username, password))
			return 400, "Bad Request"
		case "thordb: invalid password":
			log.Print(fmt.Sprintf("thordb: failed login attempt: %s//%s", username, password))
			return 400, "Bad Request"
		case "thordb: already logged in":
			log.Printf("thordb: failed login attempt (already logged in): %s//%s", username, password)
			return 400, "Bad Request"
		default:
			return 500, "Internal Server Error"
		}
	}
	log.Print("List of character ID's: ", charIDs)

	var resp request.LoginResponse
	resp.UserToken = token
	resp.CharacterIDs = charIDs
	var jsonBytes []byte
	jsonBytes, err = json.Marshal(&resp)
	if err != nil {
		log.Print(err)
		return 500, "Internal Server Error"
	}
	return 200, string(jsonBytes)
}

func handleClientRegister(httpReq *http.Request) (int, string) {
	//using authentication struct for now because i haven't added the token yet
	var req request.Authentication
	decoder := json.NewDecoder(httpReq.Body)
	err := decoder.Decode(&req)
	if err != nil {
		fmt.Println("error decoding register account request (authentication)")
		return 500, "Internal Server Error"
	}

	var username string
	var password string

	username, password, err = sanitize(req.Username, req.Password)
	if err != nil {
		log.Print("Error sanitizing authentication request", req.Username, req.Password)
		return 400, "Bad Request"
	}

	uid, err := thordb.RegisterAccount(username, password)
	if err != nil {
		log.Print(err)
		switch err.Error() {
		case "thordb: already in use":
			return 400, "Bad Request"
		default:
			return 500, "Internal Server Error"
		}
	}

	log.Print("thordb: registered player (", uid, ")")
	return 200, fmt.Sprintf("UserId%s", strconv.Itoa(uid))
}

func handleClientDisconnect(httpReq *http.Request) (int, string) {

	// getting session key might become more complicated later (add a request struct)
	bytes, err := ioutil.ReadAll(httpReq.Body)
	if err != nil {
		return 400, "Bad Request"
	}

	accountSession := string(bytes)
	err = thordb.Disconnect(accountSession)
	if err != nil {
		log.Print("thordb couldnt disconnect, something went wrong")
		log.Print(err)
		return 400, "Bad Request"
	}

	return 200, "OK"
}

func handleCreateCharacter(httpReq *http.Request) (int, string) {
	var req request.CreateCharacter
	decoder := json.NewDecoder(httpReq.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Print("character create req json decoding error %s", httpReq.Body)
		return 400, "Bad Request"
	}

	character := thordb.NewCharacterData()
	character.Name = req.Name

	var characterSession *thordb.CharacterSession
	characterSession, err = thordb.CreateCharacter(req.AccountToken, character)
	if err != nil {
		log.Print(err)
		switch err.Error() {
		case "thordb: already in use":
			return 400, "Bad Request"
		case "token contains an invalid number of segments":
			return 400, "Bad Request"
		default:
			return 500, "Internal Server Error"
		}
	}
	var jsonBytes []byte
	jsonBytes, err = json.Marshal(characterSession.CharacterData)
	if err != nil {
		log.Print(err)
		return 500, "Internal Server Error"
	}
	log.Printf("thordb: created character %d - %s\ncharacter details:\n%s", characterSession.ID, characterSession.Name, string(jsonBytes))
	return 200, characterSession.Token
}

func handleGetCharacter(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleGetCharProfile(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleClientJoinGame(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleClientJoinQueue(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleGameServerStatus(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleGetServerList(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleGetGameInfo(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleRegisterMachine(httpReq *http.Request) (int, string) {

	decoder := json.NewDecoder(httpReq.Body)
	var req request.RegisterMachine
	err := decoder.Decode(&req)
	if err != nil {
		logerr("Error decoding machine register request", err)
		return 500, "Internal Server Error"
	}

	if req.Port == 0 {
		fmt.Println("No Port Given")
		return 500, "No Port Given"
	} else {
		fmt.Println("register port = ", req.Port)
	}

	machineIp := strings.Split(httpReq.RemoteAddr, ":")[0]

	var machineId int
	machineId, err = thordb.RegisterMachine(machineIp, req.Port)
	if err != nil {
		logerr("error registering machine", err)
		return 500, "Internal Server Error"
	}
	fmt.Println("machine registered, ip=", machineIp)
	return 200, strconv.Itoa(machineId)
}

func handleUnregisterMachine(params martini.Params) (int, string) {

	machineId, err := strconv.Atoi(params["id"])
	if err != nil {
		logerr(fmt.Sprint("unable to convert request parameter, id=", params["id"]), err)
		return 400, "Bad Request"
	}

	success, err := thordb.UnregisterMachine(machineId)
	if err != nil || success {
		logerr("unable to remove machine registry", err)
	}

	fmt.Println("machine unregistered, id=", machineId)
	return 200, "OK"
}

func handleGameRequest(httpReq *http.Request) (int, string) {
	fmt.Println("[ThoriumNET] master-server.handleGameRequest")

	decoder := json.NewDecoder(httpReq.Body)
	var req request.NewGame
	err := decoder.Decode(&req)
	if err != nil {
		logerr("unable to decode body data", err)
		return 500, "Internal Server Error"
	}

	if req.Map == "" {
		fmt.Println("No Map Name Given")
		return 400, "Missing Parameters"
	}

	var gameId int
	gameId, err = thordb.RegisterNewGame(req.Map, req.MaxPlayers)
	if err != nil {
		fmt.Println("[ThoriumNET] unable to insert new game record")
		fmt.Println(err)
		return 500, "Internal Server Error"
	}

	fmt.Println("[ThoriumNET] new game, id=", strconv.Itoa(gameId))
	return 200, "OK"
}

func handleRegisterServer(httpReq *http.Request, params martini.Params) (int, string) {
	decoder := json.NewDecoder(httpReq.Body)
	var req request.RegisterGame
	err := decoder.Decode(&req)
	if err != nil {
		logerr("Error decoding machine register request", err)
		return 500, "Internal Server Error"
	}

	if req.Port == 0 {
		fmt.Println("No Port Given")
		return 400, "Missing Parameters"
	}

	var gameId int
	gameId, err = strconv.Atoi(params["id"])
	if err != nil {
		logerr(fmt.Sprintf("unable to convert parameter id=%s to integer", params["id"]), err)
		return 400, "Bad Request"
	}

	fmt.Println("[ThoriumNET] master-server.handleRegisterServer ID=", gameId)

	registered, err := thordb.RegisterActiveGame(gameId, req.MachineId, req.Port)
	if err != nil || !registered {
		logerr("unable to register game", err)
		return 500, "Internal Server Error"
	}

	fmt.Println("Found game ", gameId)
	return 200, "OK"

}

func handleCharacterSelect(httpReq *http.Request) (int, string) {
	decoder := json.NewDecoder(httpReq.Body)
	var req request.SelectCharacter
	err := decoder.Decode(&req)
	if err != nil || req.ID == 0 || req.AccountToken == "" {
		log.Print("bad json request", httpReq.Body)
		return 400, "Bad Request"
	}

	charSession, err := thordb.SelectCharacter(req.AccountToken, req.ID)
	if err != nil {
		log.Print(err)
		return 500, "Internal Server Error"
	}

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(charSession.Token)
	if err != nil {
		log.Print(err)
		return 500, "Internal Server Error"
	}

	return 200, string(jsonBytes)
}

func sanitize(username string, password string) (string, string, error) {
	return username, password, nil
}

// TODO: Refactor into logging package
func logerr(msg string, err error) {
	fmt.Println("[ThoriumNET] ", msg)
	fmt.Println(err)
}
