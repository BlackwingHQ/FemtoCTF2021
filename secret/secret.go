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
	//return hex.EncodeToString(append(
	//iv, CBCEncrypt(Key, PKCS7([]byte(input), 16), iv)...))
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
		/* 		(w).Header().Set("Access-Control-Allow-Origin", "*")
		   		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		   		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		*/switch r.Method {
		case "OPTIONS":
			return
		case "POST":
			// password := r.URL.Query().Get("password")
			var cred Credential

			// Try to decode the request body into the struct. If there is an error,
			// respond to the client with the error message and a 400 status code.
			err := json.NewDecoder(r.Body).Decode(&cred)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if cred.Password == "planet!!!11" {
				expire := time.Now().AddDate(0, 0, 1)
				userCookie := http.Cookie{
					Name:    "USER",
					Value:   encryptString("Zero Cool"),
					Path:    "/",
					Expires: expire,
					MaxAge:  86400}
				http.SetCookie(w, &userCookie)
				src_ip, err := requestSourceIp(r)
				if err != nil {
					panic(err)
				}
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
				//fmt.Fprintf(w, "{\"data\": \"Invalid login\"}")
			}
			return
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			/* fmt.Fprintf(w, "{\"data\": \"Sorry, only POST method supported.\"}") */
			return
		}

	})

	http.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		setupCORS(&w, r)
		/* 	(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		*/ //var cookie, err = r.Cookie("USER")
		switch r.Method {
		case "OPTIONS":
			return
		case "POST":
			// password := r.URL.Query().Get("password")
			var cred Credential

			// Try to decode the request body into the struct. If there is an error,
			// respond to the client with the error message and a 400 status code.
			err := json.NewDecoder(r.Body).Decode(&cred)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if cred.Token != "" {
				fmt.Println("cred.Token:", cred.Token)
				w.Header().Set("Content-Type", "application/json")
				// fmt.Fprintf(w, DecryptString(cookie.Value))
				m, _ := json.Marshal(
					map[string]string{"data": decryptString(cred.Token)})
				fmt.Fprintf(w, "%s", string(m))
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
			/* fmt.Fprintf(w, "{\"data\": \"Sorry, only POST method supported.\"}") */
		}
		// if err == nil {
		// 	// var value = cookie.Value
		// 	w.Header().Set("Content-Type", "application/json")
		// 	// fmt.Fprintf(w, DecryptString(cookie.Value))
		// 	m, _ := json.Marshal(
		// 		map[string]string{"user": DecryptString(cookie.Value)})
		// 	fmt.Fprintf(w, string(m))
		// }
	})

	// TODO: Remove
	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		var cookie, err = r.Cookie("TEST")
		if err == nil {
			// var cookievalue = cookie.Value
			// fmt.Println(DecryptString(cookievalue))
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
					// appending to existing query args
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

					// fmt.Println(resp.Status)
					// fmt.Println(string(responseBody))
					// fmt.Println("doing stuff with:", IPAddress.To4())
				} else {
					//fmt.Println(IPAddress.To4(), "good IP, but not what I'm looking for")
					http.Error(w, "Unauthorized", http.StatusForbidden)
					return
				}
				return
			} else {
				/* Should never get here... */
				http.Error(w, "Not a valid IP", http.StatusForbidden)
				return
			}
			//w.Header().Set("Content-Type", "text/plain; charset=us-ascii")
			// fmt.Fprintf(w, "Your cookie decodes to: %s\n", IPAddress.To4())
		}
		return
	})

	http.ListenAndServe(":8086", nil)
}
