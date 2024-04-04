package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var str = "a ability able about above accept according account across act action activity actually add address administration admit adult affect after again against age agency agent ago agree agreement ahead air all allow almost alone along already also although always American among amount analysis and animal another answer any anyone anything appear apply approach area argue arm around arrive art article artist as ask assume at attention attorney audience author authority available avoid away back bad bag ball bank bar base be beat beautiful because become bed before begin behavior behind believe benefit best better between beyond big bill billion bit black blood blue board body book born both box boy break bring brother budget build building business but buy by call camera campaign can cancer candidate capital car card care career carry case catch cause cell center central century certain certainly chair challenge chance change character charge check child choice choose church citizen city civil claim class clear clearly close coach cold collection college color come commercial common community company compare computer concern condition conference Congress consider consumer contain continue control cost could country couple course court cover create crime cultural culture cup current customer cut dark data daughter day dead deal debate decade decide decision deep defense degree democratic describe design despite detail determine develop development difference different difficult dinner direction director discover discuss discussion disease do doctor dog door down draw dream drive drop drug during each early east easy eat economic economy edge education effect effort eight either election else employee end energy enjoy enough enter entire environment environmental especially establish even evening event ever every everybody everyone everything evidence exactly example executive exist expect experience expert explain eye face fact factor fail fall family far fast father fear federal feel feeling few field fight figure fill film final finally financial find fine finger finish fire firm first fish five floor fly focus follow food foot for force foreign forget form former forward four free friend from front full fund future game garden gas general generation get girl give glass go goal good government great green ground group grow growth guess gun guy hair half hand hang happen happy hard have he head health hear heart heat heavy help her here herself high him himself his history hit hold home hope hospital hot hotel hour house how however huge human hundred husband idea identify if image imagine impact important improve in include including increase indeed indicate individual industry information inside instead institution interest interesting international interview into investment involve issue it item its itself job join just keep key kid kind kitchen know knowledge land language large last late later laugh law lawyer lay lead leader learn least leave left leg legal less let letter level lie life light like likely line list listen little live local long look lose loss lot love low machine magazine main maintain major majority make man manage management manager many market marriage material matter may maybe mean measure media medical meet meeting member memory mention message method middle might military million mind minute miss mission model modern moment money month more morning most mother mouth move movement movie much music must my myself name nation national natural nature near nearly necessary need network never new news newspaper next nice night no none nor north not note nothing notice now number occur of off offer office officer official often oh oil ok old on once one only onto open operation opportunity option or order organization other others our out outside over own owner page pain painting paper parent part participant particular particularly partner party pass past patient pattern pay peace people per perform performance perhaps period person personal phone physical pick picture piece place plan plant play player PM point police policy political politics poor popular population position positive possible power practice prepare present president pressure pretty prevent price private probably problem process produce product production professional professor program project property protect prove provide public pull purpose push put quality question quickly quite race radio raise range rate rather reach read ready real reality realize really reason receive recent recently recognize record red reduce reflect region relate relationship religious remain remember remove report represent require research resource respond response responsibility rest result return reveal rich right rise risk road rock role room rule run safe same save say scene school science scientist score sea season seat second section security see seek seem sell send senior sense series serious serve service set seven several shake share she shoot short shot should shoulder show side sign significant similar simple simply since sing single sister sit site situation six size skill skin small smile so social society soldier some somebody someone something sometimes son song soon sort sound source south southern space speak special specific speech spend sport spring staff stage stand standard star start state statement station stay step still stock stop store story strategy street strong structure student study stuff style subject success successful such suddenly suffer suggest summer support sure surface system table take talk task tax teach teacher team technology television tell ten tend term test than thank that the their them themselves then theory there these they thing think third this those though thought thousand threat three through throughout throw thus time to today together tonight too top total tough toward town trade traditional training travel treat treatment tree trial trip trouble true truth try turn TV two type under understand unit until up upon us use usually value various very victim view violence visit voice vote wait walk wall want war watch water way we weapon wear week weight well west western what whatever when where whether which while white who whole whom whose why wide wife will win wind window wish with within without woman wonder word work worker world worry would write writer wrong yard yeah year yes yet you young your yourself"
var words = strings.Fields(str)

// Manually test with hard-coded values. If set to empty strings, they will be
// set in-memory by other cp-admin functions.
var testEmail = ""
var testUserId = ""
var testToken = ""

// Manually test with hard-coded value. If set to 0, the login code will be
// retrieved from the api server (admin email bypass).
var testLoginCode = 0

// TODO: Remove the above variables and related functionality once things are
// covered in E2E tests.

func generateRandomEmailAddress() string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	return words[rng.Intn(len(words))] + words[rng.Intn(len(words))] + words[rng.Intn(len(words))] + "@email.com"
}

func generatePlaceholderText(numWords int) string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	result := ""
	for i := 0; i < numWords; i++ {
		result += words[rng.Intn(len(words))] + " "
	}
	return result
}

func signup(email string) (string, error) {
	type responseBody struct {
		UserId string `json:"userId"`
		Error  string `json:"error"`
	}
	var responseBodyInst responseBody
	var url = "http://localhost:8000/api/user/signup/"
	var jsonData = []byte(`{"email":"` + email + `"}`)

	// Send POST request using the default http client.
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[err][admin] posting request: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("[admin] response status: %s [%s]\n", resp.Status, cts())

	unmarshalOrExit(resp.Body, &responseBodyInst)

	// Check if the server returned an error message.
	if responseBodyInst.Error != "" {
		fmt.Printf("[admin] api server returned error: %s [%s]\n", responseBodyInst.Error, cts())
		return responseBodyInst.UserId, fmt.Errorf(responseBodyInst.Error)
	}

	return responseBodyInst.UserId, nil
}

func wrappedSignup() {
	if testEmail == "" {
		testEmail = generateRandomEmailAddress()
	}
	var err error
	testUserId, err = signup(testEmail)
	if err != nil {
		return
	}
	fmt.Printf("[admin] successfully signed up email: %s, with userId: %s [%s]\n", testEmail, testUserId, cts())
}

func login(email string) (string, error) {
	type responseBody struct {
		UserId string `json:"userId"`
		Error  string `json:"error"`
	}
	var responseBodyInst responseBody
	var url = "http://localhost:8000/api/user/login/"
	var jsonData = []byte(fmt.Sprintf(`{"email":"%s"}`, email))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[err][admin] posting request: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("[admin] response status: %s [%s]\n", resp.Status, cts())

	unmarshalOrExit(resp.Body, &responseBodyInst)

	// Check if the server returned an error message.
	if responseBodyInst.Error != "" {
		fmt.Printf("[admin] api server returned error: %s [%s]\n", responseBodyInst.Error, cts())
		return responseBodyInst.UserId, fmt.Errorf(responseBodyInst.Error)
	}

	return responseBodyInst.UserId, nil
}

func wrappedLogin() {
	if testEmail == "" {
		fmt.Printf("[err][admin] no hardcoded test email - add email or signup a new user first [%s]\n", cts())
		return
	}
	var err error
	testUserId, err = login(testEmail)
	if err != nil {
		return
	}
	fmt.Printf("[admin] email: %s, userId: %s [%s]\n", testEmail, testUserId, cts())
}

// Get a loginCode for a given userId by posting a request to a restricted
// endpoint called bypass-email. Normally a code is emailed to users.
func getLoginCode(userId string) (int, error) {
	type responseBody struct {
		LoginCode     int       `json:"loginCode"`
		LoginAttempts int       `json:"loginAttempts"`
		LogoutTs      time.Time `json:"logoutTs"`
		Error         string    `json:"error"`
	}
	var responseBodyInst responseBody
	var url = fmt.Sprintf("http://localhost:8000/api/admin/bypass-email/%s", testUserId)

	// Create a new request using http.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("[err][admin] creating request: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Set custom admin auth header.
	req.Header.Set("Admin-Authorization", adminAuthToken)

	// Send the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[err][admin] posting request: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("[admin] response status: %s [%s]\n", resp.Status, cts())

	unmarshalOrExit(resp.Body, &responseBodyInst)

	// Check if the server returned an error message.
	if responseBodyInst.Error != "" {
		fmt.Printf("[admin] api server returned error: %s [%s]\n", responseBodyInst.Error, cts())
		return responseBodyInst.LoginCode, fmt.Errorf(responseBodyInst.Error)
	}

	return responseBodyInst.LoginCode, nil
}

func loginCode(userId string, code int) (string, int, error) {
	type responseBody struct {
		Token             string `json:"token"`
		RemainingAttempts int    `json:"remainingAttempts"`
		Error             string `json:"error"`
	}
	var responseBodyInst responseBody
	var url = "http://localhost:8000/api/user/login-code/"
	var jsonData = []byte(fmt.Sprintf(`{"userId":"%s","code":%d}`, userId, code))
	// Send POST request using the default http client.
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[err][admin] posting request: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("[admin] response status: %s [%s]\n", resp.Status, cts())

	unmarshalOrExit(resp.Body, &responseBodyInst)

	// Check if the server returned an error message.
	if responseBodyInst.Error != "" {
		fmt.Printf("[admin] api server returned error: %s [%s]\n", responseBodyInst.Error, cts())
		return "", responseBodyInst.RemainingAttempts, fmt.Errorf(responseBodyInst.Error)
	}

	return responseBodyInst.Token, responseBodyInst.RemainingAttempts, nil
}

func wrappedLoginCode() {
	// Check to make sure testUserId is set.
	if testUserId == "" {
		fmt.Printf("[err][admin] no hard-coded test userId - add manually, or signup/login a new user first [%s]\n", cts())
		return
	}

	// No hard-coded testLoginCode. Get code from the server.
	if testLoginCode == 0 {
		var err error
		testLoginCode, err = getLoginCode(testUserId)
		if err != nil {
			return
		}
	}

	// Proceed with api call to login-code.
	var err error
	var attempts int
	testToken, attempts, err = loginCode(testUserId, testLoginCode)
	if err != nil {
		fmt.Printf("[admin] userId: %s, remainingAttempts: %d [%s]\n", testUserId, attempts, cts())
		return
	}

	fmt.Printf("[admin] userId: %s, token: %s [%s]\n", testUserId, testToken, cts())
}

func logout() {
	type responseBody struct {
		Error string `json:"error"`
	}
	var responseBodyInst responseBody
	url := "http://localhost:8000/api/user/logout/"
	jsonData := []byte(fmt.Sprintf(`{"userId":"%s"}`, testUserId))

	// Create a new request using http.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[err][admin] creating new request: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Add authorization header to the request.
	req.Header.Add("Authorization", "Bearer "+testToken)
	// Set Content-Type header to application/json.
	req.Header.Set("Content-Type", "application/json")

	// Send request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[err][admin] posting request: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("[admin] response status: %s [%s]\n", resp.Status, cts())

	// A 204 status code (no content) is expected. If it's anything else,
	// proceed with unmarshaling the response body to get the error.
	if resp.StatusCode != http.StatusNoContent {
		unmarshalOrExit(resp.Body, &responseBodyInst)
		fmt.Printf("[admin] server returned error: %s [%s]\n", responseBodyInst.Error, cts())
	}

}

func createExim() {
	type requestBody struct {
		Target     string `json:"target"`
		Title      string `json:"title"`
		Summary    string `json:"summary"`
		Paragraph1 string `json:"paragraph1"`
		Paragraph2 string `json:"paragraph2"`
		Paragraph3 string `json:"paragraph3"`
		Link       string `json:"link"`
	}
	type responseBody struct {
		EximId string `json:"eximId"`
		Error  string `json:"error"`
	}
	var requestBodyInst requestBody
	var responseBodyInst responseBody
	var url = "http://localhost:8000/api/exim/create/"

	// Fill Exim with random, placeholder text.
	requestBodyInst.Target = "FEDERAL"
	requestBodyInst.Title = generatePlaceholderText(5)
	requestBodyInst.Summary = generatePlaceholderText(20)
	requestBodyInst.Paragraph1 = generatePlaceholderText(40)
	requestBodyInst.Paragraph2 = generatePlaceholderText(40)
	requestBodyInst.Paragraph3 = generatePlaceholderText(40)
	requestBodyInst.Link = fmt.Sprintf("https://%s.com", generatePlaceholderText(3))

	// Marshal the request body to JSON.
	jsonData, err := json.Marshal(requestBodyInst)
	if err != nil {
		fmt.Printf("[err][admin] marshaling request body: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Create a new request using http.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[err][admin] creating new request: %v [%s]\n", err, cts())
		os.Exit(1)
	}

	// Add authorization header to the request.
	req.Header.Add("Authorization", "Bearer "+testToken)
	// Set Content-Type header to application/json.
	req.Header.Set("Content-Type", "application/json")

	// Send request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[err][admin] posting request: %v [%s]\n", err, cts())
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("[admin] response status: %s [%s]\n", resp.Status, cts())

	unmarshalOrExit(resp.Body, &responseBodyInst)

	// Check if the server returned an error message.
	if responseBodyInst.Error != "" {
		fmt.Printf("[admin] api server returned error: %s [%s]\n", responseBodyInst.Error, cts())
	} else {
		fmt.Printf("[admin] ulid of new exim: %s [%s]\n", responseBodyInst.EximId, cts())
	}
}
