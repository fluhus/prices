package aggregators

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"encoding/json"
)

// Cerberus login page.
const cerberusHome = "https://url.publishedprices.co.il/"

// Aggregates data files from the Cerberus server.
type CerberusAggregator struct {
	username string
}

// Returns a new Cerberus aggregator with the given user-name.
func NewCerberusAggregator(username string) *CerberusAggregator {
	return &CerberusAggregator{username}
}

func (a *CerberusAggregator) Aggregate(dir string) error {
	// Get login page.
	// fmt.Println("hi")
	r, err := http.Get("https://url.publishedprices.co.il/")
	checkError(err)
	body, err := ioutil.ReadAll(r.Body)
	checkError(err)
	
	// Parse token and cookie.
	cookie := find([]byte(r.Header["Set-Cookie"][0]), "cftpSID=(.*?);")
	token := find(body, "id=\"csrftoken\" value=\"(.*?)\"")
	fmt.Println("cookie:", string(cookie))
	fmt.Println("token:", string(token))
	r.Body.Close()
	
	
	
	// Request file list.
	r, err = cl.PostForm("https://url.publishedprices.co.il/file/ajax_dir?sEcho=2&iColumns=5&sColumns=%2C%2C%2C%2C&iDisplayStart=0&iDisplayLength=100000&mDataProp_0=fname&sSearch_0=&bRegex_0=false&bSearchable_0=true&bSortable_0=true&mDataProp_1=type&sSearch_1=&bRegex_1=false&bSearchable_1=true&bSortable_1=false&mDataProp_2=size&sSearch_2=&bRegex_2=false&bSearchable_2=true&bSortable_2=true&mDataProp_3=ftime&sSearch_3=&bRegex_3=false&bSearchable_3=true&bSortable_3=true&mDataProp_4=&sSearch_4=&bRegex_4=false&bSearchable_4=true&bSortable_4=false&sSearch=&bRegex=false&iSortingCols=0&cd=%2F",
			nil)
	checkError(err)
	var files struct {
		AaData []*struct {
			Value string
		}
	}
	
	err = json.NewDecoder(r.Body).Decode(&files)
	checkError(err)
	fmt.Println(files.AaData)
}

// Looks up a regular expression in the given sequence and returns the #1
// captured group. If not found, returns null - which is different from a 0-long
// array.
func find(text []byte, exp string) []byte {
	return regexp.MustCompile(exp).FindSubmatch(text)[1]
}

// Logs in to Cerberus and returns the login cookie value. The cookie's name is
// 'cftpSID'.
func (a *CerberusAggregator) login() (cookie []byte, err error) {
	// Get login page.
	res, err := http.Get(cerberusHome)
	if err != nil { return }
	body, err := ioutil.ReadAll(res.Body)
	if err != nil { return }
	res.Body.Close()
	
	// Parse token.
	token := find(body, "id=\"csrftoken\" value=\"(.*?)\"")
	if token == nil {
		return nil, fmt.Errorf("Could not parse login token.")
	}
	
	// Parse cookie.
	rawCookie, ok := res.Header["Set-Cookie"]
	if !ok {
		return nil, fmt.Errorf("Login page did not give a cookie.")
	}
	cookie := find([]byte(rawCookie[0]), "cftpSID=(.*?);")
	if cookie == nil {
		return nil, fmt.Errorf("Could not parse home-page cookie value. " +
				"Cookie: %v", rawCookie)
	}
	
	// Login.
	jar, err := singleCookieJar(cerberusHome, "cftpSID", string(cookie))
	if err != nil { return nil, err }
	urll, err := url.Parse("https://url.publishedprices.co.il/")
	checkError(err)
	cookies := []*http.Cookie{&http.Cookie{Name: "cftpSID",
			Value: string(cookie)}}
	jar.SetCookies(urll, cookies)
	cl := &http.Client{Jar: jar}
	r, err = cl.PostForm(
		"https://url.publishedprices.co.il/login/user",
		map[string][]string{
			"csrftoken": []string{string(token)},
			"username": []string{"doralon"},
			"password": []string{""},
			"Submit": []string{"Sign in"},
		})
	checkError(err)
	
	// Get login cookie.
	loginCookie := find([]byte(r.Header["Set-Cookie"][0]), "cftpSID=(.*?);")
	cookies = []*http.Cookie{&http.Cookie{Name: "cftpSID",
			Value: string(loginCookie)}}
	jar.SetCookies(urll, cookies)
}

// Returns a cookie-jar with a single cookie. Error shouldn't happen unless
// path is malformed.
func singleCookieJar(path, name, value string) (http.CookieJar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil { return nil, err }
	pathUrl, err := url.Parse(path)
	if err != nil { return nil, err }
	jar.AddCookies(pathUrl, []*http.Cookie{&http.Cookie{
		Name: name, Value: value}})
	return jar, nil
}
