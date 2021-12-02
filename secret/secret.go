package secret

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var Key []byte = make([]byte, 16)
var IV []byte = make([]byte, 16)

func decryptString(input string) string {
	s, err := hex.DecodeString(input)
	if err != nil {
		panic(err)
	}
	iv := s[0:16]
	ciphertext := s[16:]
	plaintext, err := aesDecrypt(Key, ciphertext, iv)
	if err != nil {
		panic(err)
	}
	ptext, err := pkcs7Unpad(plaintext)
	if err != nil {
		panic(err)
	}
	return string(ptext)
}

func encryptString(input string) string {
	iv, err := generateRandomBytes(16)
	if err != nil {
		panic(err)
	}
	ciphertext, err := aesEncrypt(Key, pkcs7Pad([]byte(input), 16), iv)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(append(iv, ciphertext...))
}

func genKey() {
	k, err := generateRandomBytes(16)
	if err != nil {
		panic(err)
	}
	Key = k
}

func updateKey() {
	for range time.Tick(time.Minute * 5) {
		fmt.Println("Regenerating Key")
		genKey()
	}
}

func encodeIP(ip string) string {
	ip_slice := strings.Split(ip, ".")
	trimmed_ip := make([]string, len(ip_slice))
	if len(trimmed_ip) != 4 {
		return ""
		//panic("bad ip")
	}
	for c, s := range ip_slice {
		trimmed_ip[c] = fmt.Sprintf("%03s", s)
	}
	return encryptString(strings.Join(trimmed_ip, "."))
}

func decodeIP(cookie string) net.IP {
	ip := decryptString(cookie)
	ip_slice := strings.Split(ip, ".")
	trimmed_ip := make([]string, len(ip_slice))
	if len(trimmed_ip) != 4 {
		return nil
		// panic("bad ip")
	}
	for c, s := range ip_slice {
		//fmt.Println(s)
		i, err := strconv.Atoi(s)
		if err != nil {
			fmt.Println(err)
			break
		}
		trimmed_ip[c] = strconv.Itoa(i)
	}
	str := strings.Join(trimmed_ip, ".")
	return net.ParseIP(str)
}

func requestSourceIp(req *http.Request) (string, error) {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return "", err
	}
	return host, nil
}

type Credential struct {
	Password string
	Token    string
}

func setupCORS(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, 	Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func Secret() {
	genKey()
	go updateKey()
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		setupCORS(&w, r)
		switch r.Method {
		case "OPTIONS":
			return
		case "POST":
			var cred Credential
			err := json.NewDecoder(r.Body).Decode(&cred)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			src_ip, err := requestSourceIp(r)
			if err != nil {
				panic(err)
			}
			fmt.Println("IP:", src_ip, "cred.Password:", cred.Password)
			if cred.Password == "planet!!!11" {
				expire := time.Now().AddDate(0, 0, 1)
				userCookie := http.Cookie{
					Name:    "USER",
					Value:   encryptString("Zero Cool"),
					Path:    "/",
					Expires: expire,
					MaxAge:  86400}
				http.SetCookie(w, &userCookie)
				testCookie := http.Cookie{
					Name:    "TEST",
					Value:   encodeIP(src_ip),
					Path:    "/",
					Expires: expire,
					MaxAge:  86400}
				http.SetCookie(w, &testCookie)
				fmt.Fprintf(w, "{\"data\": \"Success\", \"token\": \"%s\", \"note\": \"This is not the flag\"}",
					encryptString("Zero Cool"))
			} else {
				http.Error(w, "{\"data\": \"Invalid Password\"}", http.StatusUnauthorized)
			}
			return
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

	})

	http.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		setupCORS(&w, r)
		switch r.Method {
		case "OPTIONS":
			return
		case "POST":
			var cred Credential
			err := json.NewDecoder(r.Body).Decode(&cred)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if cred.Token != "" {
				fmt.Println("cred.Token:", cred.Token)
				w.Header().Set("Content-Type", "application/json")
				m, _ := json.Marshal(
					map[string]string{"data": decryptString(cred.Token)})
				fmt.Fprintf(w, "%s", string(m))
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

	})

	// TODO: Remove
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		var cookie, err = r.Cookie("TEST")
		if err == nil {
			IPAddress := decodeIP(cookie.Value)
			if IPAddress != nil {
				if IPAddress.IsLoopback() {
					client := &http.Client{}
					req, err := http.NewRequest(http.MethodGet, "http://"+IPAddress.String()+":1337/admin", nil)
					if err != nil {
						log.Fatal(err)
					}
					input := r.URL.Query().Get("input")
					dbg := r.URL.Query().Get("dbg")
					q := req.URL.Query()
					q.Add("input", input)
					q.Add("dbg", dbg)
					req.URL.RawQuery = q.Encode()
					resp, err := client.Do(req)
					if err != nil {
						fmt.Println("Error when sending request to the server")
						return
					}
					defer resp.Body.Close()
					responseBody, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Fprintf(w, string(responseBody))
				} else {
					http.Error(w, "Unauthorized", http.StatusForbidden)
					return
				}
				return
			} else {
				/* Should never get here... */
				http.Error(w, "Not a valid IP", http.StatusForbidden)
				return
			}
		}
		return
	})

	http.ListenAndServe(":8086", nil)
}
