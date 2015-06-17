package aggregators

import (
	"net/http"
	"fmt"
	"io/ioutil"
	//"os"
	"encoding/json"
)

const (
	// Cerberus login page.
	cerberusHome = "https://url.publishedprices.co.il/"
	
	// Cerberus user page (with file list).
	cerberusUser = cerberusHome + "login/user"
)

// Aggregates data files from the Cerberus server.
type CerberusAggregator struct {
	username string
}

// Returns a new Cerberus aggregator with the given user-name.
func NewCerberusAggregator(username string) *CerberusAggregator {
	return &CerberusAggregator{username}
}

func (a *CerberusAggregator) Aggregate(dir string) error {
	client, err := a.login()
	if err != nil { return fmt.Errorf("Failed to login: %v", err) }
	
	
	
	// Request file list.
	r, err := client.PostForm("https://url.publishedprices.co.il/file/ajax_dir?sEcho=2&iColumns=5&sColumns=%2C%2C%2C%2C&iDisplayStart=0&iDisplayLength=100000&mDataProp_0=fname&sSearch_0=&bRegex_0=false&bSearchable_0=true&bSortable_0=true&mDataProp_1=type&sSearch_1=&bRegex_1=false&bSearchable_1=true&bSortable_1=false&mDataProp_2=size&sSearch_2=&bRegex_2=false&bSearchable_2=true&bSortable_2=true&mDataProp_3=ftime&sSearch_3=&bRegex_3=false&bSearchable_3=true&bSortable_3=true&mDataProp_4=&sSearch_4=&bRegex_4=false&bSearchable_4=true&bSortable_4=false&sSearch=&bRegex=false&iSortingCols=0&cd=%2F", nil)
	if err != nil { return err }
	var files struct {
		AaData []*struct {
			Value string
		}
	}
	
	err = json.NewDecoder(r.Body).Decode(&files)
	if err != nil { return err }
	fmt.Println(files.AaData)
	//*/
	
	return nil
}

// Returns a logged-in client.
func (a *CerberusAggregator) login() (*http.Client, error) {
	// Get login page.
	res, err := http.Get(cerberusHome)
	if err != nil { return nil, fmt.Errorf("Failed to get homepage: %v", err) }
	body, err := ioutil.ReadAll(res.Body)
	if err != nil { return nil, fmt.Errorf("Failed to read homepage: %v", err) }
	res.Body.Close()

	// Get token and cookie.	
	token, err := parseCerberusToken(body)
	if err != nil { return nil, err }
	cookie, err := parseCerberusCookie(res)
	if err != nil { return nil, err }
	
	// Login!
	jar, err := singleCookieJar(cerberusHome, "cftpSID", string(cookie))
	if err != nil { return nil, err }
	
	cl := &http.Client{Jar: jar}
	res, err = cl.PostForm(
		cerberusUser,
		map[string][]string{
			"csrftoken": []string{string(token)},
			"username": []string{a.username},
			"password": []string{""},
			"Submit": []string{"Sign in"},
		})
	if err != nil { return nil, fmt.Errorf("Failed to post: %v", err) }
	
	// Get second cookie.
	cookie, err = parseCerberusCookie(res)
	if err != nil { return nil, err }
	
	// Update client with new cookie.
	cl.Jar, err = singleCookieJar(cerberusHome, "cftpSID", string(cookie))
	if err != nil { return nil, err }
	
	return cl, nil
}

// Parses the Get-Cookie field of a Cerberus response.
func parseCerberusCookie(res *http.Response) ([]byte, error) {
	rawCookie, ok := res.Header["Set-Cookie"]
	if !ok {
		return nil, fmt.Errorf("Response does not contain a cookie.")
	}
	cookie := find([]byte(rawCookie[0]), "cftpSID=(.*?);")
	if cookie == nil {
		return nil, fmt.Errorf("Could not parse cookie value. " +
				"Cookie: %v", rawCookie)
	}
	return cookie, nil
}

// Parses the login token from a Cerberus response body.
func parseCerberusToken(body []byte) ([]byte, error) {
	token := find(body, "id=\"csrftoken\" value=\"(.*?)\"")
	if token == nil {
		return nil, fmt.Errorf("Could not parse login token.")
	}
	return token, nil
}

//func (a *CerberusAggregator)

